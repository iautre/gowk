package auth

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/iautre/gowk"
	"github.com/iautre/gowk/auth/proto"
)

// AuthServer implements the AuthService gRPC interface
type AuthServer struct {
	proto.UnimplementedAuthServiceServer
}

// NewAuthServer creates a new AuthServer
func NewAuthServer() *AuthServer {
	return &AuthServer{}
}

// Login handles user login
func (s *AuthServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	if req.Account == "" || req.Code == "" {
		return nil, status.Errorf(codes.InvalidArgument, "account and code are required")
	}

	userService := UserService{}
	user, err := userService.Login(ctx, &LoginParams{
		Account: req.Account,
		Code:    req.Code,
	})
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}

	// Create token directly for gRPC context
	token := gowk.UUID()

	// TODO: Store token in database/cache for validation

	return &proto.LoginResponse{
		Token:    token,
		UserId:   user.ID,
		Nickname: user.Nickname.String,
	}, nil
}

// Register handles user registration
func (s *AuthServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.LoginResponse, error) {
	// TODO: Implement user registration logic
	return nil, status.Errorf(codes.Unimplemented, "registration not implemented")
}

// UserInfo handles user info request
func (s *AuthServer) UserInfo(ctx context.Context, req *proto.UserInfoRequest) (*proto.UserInfoResponse, error) {
	if req.UserId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	userService := UserService{}
	user, err := userService.GetById(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &proto.UserInfoResponse{
		Id:       user.ID,
		Phone:    user.Phone.String,
		Email:    user.Email.String,
		Nickname: user.Nickname.String,
		Group:    user.Group.String,
	}, nil
}

// TODO: Implement OAuth2/OIDC methods after proto file is updated
// OAuth2Auth, OAuth2Token, OIDCDiscovery, OIDCUserInfo, OIDCJwks
