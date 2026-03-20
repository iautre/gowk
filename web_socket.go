package gowk

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// wsOriginCheck 业务层可通过 SetWebSocketOriginCheck 替换，默认执行同源校验。
var wsOriginCheck func(r *http.Request) bool

// SetWebSocketOriginCheck 注册自定义 WebSocket 来源校验函数。
func SetWebSocketOriginCheck(f func(r *http.Request) bool) {
	wsOriginCheck = f
}

func defaultCheckOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	host := r.Host
	return strings.EqualFold(origin, "https://"+host) ||
		strings.EqualFold(origin, "http://"+host)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		if wsOriginCheck != nil {
			return wsOriginCheck(r)
		}
		return defaultCheckOrigin(r)
	},
}

var defaultMessage = &Server{
	clients:    make(map[string]*SocketClient),
	broadcast:  make(chan *Message),
	register:   make(chan *SocketClient),
	unregister: make(chan *SocketClient),
}

// wsInitOnce 确保 run goroutine 只启动一次。
var wsInitOnce sync.Once

type MessageInterface interface {
	HandlerMessage(*Message) error
}

func SendMessage(sm *Message) error {
	return defaultMessage.sendMessage(sm)
}

func WebSocketHandlerFunc(si MessageInterface) gin.HandlerFunc {
	Init()
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.Error(Error(err))
			return
		}
		defaultMessage.MessageInterface = si
		clientName := c.GetString(WEB_SOCKET_CLIENT_NAME)
		if clientName == "" {
			clientName = conn.RemoteAddr().String()
		}
		client := &SocketClient{
			conn:  conn,
			name:  clientName,
			send:  make(chan *Message),
			serve: defaultMessage,
		}
		go client.readPump()
		go client.writePump()
		defaultMessage.register <- client
	}
}

// Init 确保消息分发 goroutine 只启动一次。
func Init() {
	wsInitOnce.Do(func() {
		go defaultMessage.run()
	})
}

type Message struct {
	Sender   string     `json:"sender"`
	Receiver string     `json:"receiver"`
	Content  any        `json:"content"`
	err      *ErrorCode `json:"-"`
}

type Server struct {
	clients              map[string]*SocketClient
	broadcast            chan *Message
	register, unregister chan *SocketClient
	mu                   sync.Mutex
	MessageInterface     MessageInterface
}

func (s *Server) run() {
	for {
		select {
		case client := <-s.register:
			s.registerClient(client)
		case client := <-s.unregister:
			s.unRegisterClient(client)
		case message := <-s.broadcast:
			s.handlerMessage(message)
		}
	}
}

func (s *Server) unRegisterClient(client *SocketClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.clients[client.name]; ok {
		delete(s.clients, client.name)
		close(client.send)
	}
}

func (s *Server) registerClient(client *SocketClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client.name] = client
}

func (s *Server) handlerMessage(message *Message) {
	err := s.MessageInterface.HandlerMessage(message)
	if err != nil {
		slog.Error("消息处理失败，返回给发送者")
		var ec *ErrorCode
		if errors.As(err, &ec) {
			s.sendErrorMessage(message.Sender, ec)
		} else {
			s.sendErrorMessage(message.Sender, Error(err))
		}
	}
}

// sendErrorMessage 向指定客户端发送错误消息，不要求 Content 非空。
func (s *Server) sendErrorMessage(receiver string, ec *ErrorCode) {
	if receiver == "" {
		return
	}
	s.mu.Lock()
	client, ok := s.clients[receiver]
	s.mu.Unlock()
	if !ok {
		slog.Warn(fmt.Sprintf("sendErrorMessage: 客户端 %s 不在线", receiver))
		return
	}
	client.send <- &Message{Receiver: receiver, err: ec}
}

func (s *Server) sendMessage(message *Message) error {
	if message == nil {
		return NewError("没有发送内容")
	}
	if message.Receiver == "" {
		return NewError("没有指定接收者")
	}
	if message.Sender == "" {
		return NewError("没有发送者信息")
	}
	if message.Content == nil && message.err == nil {
		return NewError("没有发送内容")
	}
	if message.err == nil {
		message.err = SocketMsg(message.Content, message.Sender)
	}
	s.mu.Lock()
	client, ok := s.clients[message.Receiver]
	s.mu.Unlock()
	if !ok {
		return NewError(fmt.Sprintf("客户端%s不在线", message.Receiver))
	}
	client.send <- message
	return nil
}

type SocketClient struct {
	serve *Server
	conn  *websocket.Conn
	send  chan *Message
	name  string
}

func (c *SocketClient) writePump() {
	defer func() {
		c.serve.unregister <- c
		c.conn.Close()
	}()
	for {
		message, ok := <-c.send
		if !ok {
			return
		}
		if err := c.conn.WriteJSON(message.err); err != nil {
			return
		}
	}
}

func (c *SocketClient) readPump() {
	defer func() {
		c.serve.unregister <- c
		c.conn.Close()
	}()
	for {
		var message Message
		err := c.conn.ReadJSON(&message)
		if err != nil {
			var jsonErr *json.SyntaxError
			if errors.As(err, &jsonErr) {
				c.send <- &Message{Receiver: c.name, err: NewError("参数错误")}
				continue
			}
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				slog.Error(fmt.Sprintf("客户端连接异常断开: %v", err))
			}
			return
		}
		if message.Receiver == "" || message.Content == nil {
			c.send <- &Message{Receiver: c.name, err: NewError("参数错误")}
			continue
		}
		message.Sender = c.name
		c.serve.broadcast <- &message
	}
}
