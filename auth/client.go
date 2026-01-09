package auth

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authpb "github.com/iautre/gowk/auth/proto"
)

// AuthClient OAuth2客户端
type AuthClient struct {
	conn       *grpc.ClientConn
	authClient authpb.AuthServiceClient
	clientID   string
	secret     string
}

// NewAuthClient 创建OAuth2客户端
func NewAuthClient(addr, clientID, secret string) (*AuthClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %v", err)
	}

	client := authpb.NewAuthServiceClient(conn)

	return &AuthClient{
		conn:       conn,
		authClient: client,
		clientID:   clientID,
		secret:     secret,
	}, nil
}

// Login 用户登录
func (c *AuthClient) Login(ctx context.Context, account, code string) (*authpb.LoginResponse, error) {
	req := &authpb.LoginRequest{
		Account: account,
		Code:    code,
	}

	resp, err := c.authClient.Login(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("login failed: %v", err)
	}

	return resp, nil
}

// Register 用户注册
func (c *AuthClient) Register(ctx context.Context, phone, email, name string) (*authpb.LoginResponse, error) {
	req := &authpb.RegisterRequest{
		Phone: phone,
		Email: email,
		Name:  name,
	}

	resp, err := c.authClient.Register(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("register failed: %v", err)
	}

	return resp, nil
}

// UserInfo 获取用户信息
func (c *AuthClient) UserInfo(ctx context.Context, userID int64) (*authpb.UserInfoResponse, error) {
	req := &authpb.UserInfoRequest{
		UserId: userID,
	}

	resp, err := c.authClient.UserInfo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get user info failed: %v", err)
	}

	return resp, nil
}

// TODO: Implement OAuth2 methods after authpb file is updated with OAuth2 messages
// Authorize, GetToken, RefreshToken, GetUserInfo (OAuth2), RevokeToken

// Close 关闭客户端连接
func (c *AuthClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
