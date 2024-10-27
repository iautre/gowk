package gowk

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境中应该更严格
	},
}

var defaultMessage = &Server{
	clients:    make(map[string]*Client),
	broadcast:  make(chan *Message),
	register:   make(chan *Client),
	unregister: make(chan *Client),
}

type MessageInterface interface {
	ReadMessage(*Message) *ErrorCode
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
		client_name := c.GetString(WEB_SOCKET_CLIENT_NAME)
		if client_name == "" {
			client_name = conn.RemoteAddr().String()
		}
		client := &Client{
			conn:  conn,
			name:  client_name,
			send:  make(chan *Message),
			serve: defaultMessage,
		}
		go client.readPump()
		go client.writePump()
		defaultMessage.register <- client
	}
}

func Init() {
	go defaultMessage.run()
}

type Message struct {
	Sender   string     `json:"sender"`
	Receiver string     `json:"receiver"`
	Content  any        `json:"content"`
	err      *ErrorCode `json:"-"`
}

type Server struct {
	clients              map[string]*Client
	broadcast            chan *Message
	register, unregister chan *Client
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

func (s *Server) unRegisterClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.clients[client.name]; ok {
		delete(s.clients, client.name)
		close(client.send)
	}
}
func (s *Server) registerClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client.name] = client
}
func (s *Server) handlerMessage(message *Message) {
	err := s.MessageInterface.ReadMessage(message)
	if err != nil {
		slog.Error("消息处理失败，返回给发送者")
		s.sendMessage(&Message{Receiver: message.Sender, err: err})
	}
}
func (s *Server) sendMessage(message *Message) error {
	if message == nil || message.Content == nil {
		return NewError("没有发送内容")
	}
	if message.Receiver == "" {
		return NewError("没有指定接收者")
	}
	if message.Sender == "" {
		return NewError("没有发送者信息")
	}
	if message.err == nil {
		message.err = SocketMsg(message.Content, message.Sender)
	}
	if client, ok := s.clients[message.Receiver]; ok {
		client.send <- message
	} else {
		return NewError(fmt.Sprintf("客户端%s不在线", message.Receiver))
	}
	return nil
}

type Client struct {
	serve *Server
	conn  *websocket.Conn
	send  chan *Message
	name  string
}

func (c *Client) writePump() {
	defer func() {
		c.serve.unregister <- c
		c.conn.Close()
	}()
	for {
		message, ok := <-c.send
		if !ok {
			return
		}
		err := c.conn.WriteJSON(message.err)
		if err != nil {
			return
		}
	}
}
func (c *Client) readPump() {
	defer func() {
		c.serve.unregister <- c
		c.conn.Close()
	}()
	for {
		var message Message
		err := c.conn.ReadJSON(&message)
		if err != nil {
			if errors.Is(err, &json.SyntaxError{}) {
				// 判断错误是否为无效的 JSON
				c.send <- &Message{Receiver: c.name, err: NewError("参数错误")}
				continue
			}
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				slog.Error(fmt.Sprintf("客户端连接异常断开: %v\n", err))
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
