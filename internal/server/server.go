package server

import (
	"context"
	"errors"
	"oauth/internal/entity"
	jwtutil "oauth/internal/jwt"
	pb "oauth/internal/pb/proto"
	"oauth/internal/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

type oauthServer struct {
	pb.UnimplementedOAuthServer
	store      storage.Storage
	jwtManager *jwtutil.Manager
}

func NewServer(store storage.Storage) (*oauthServer, error) {
	jm := jwtutil.NewManagerFromEnv()
	return &oauthServer{
		store:      store,
		jwtManager: jm,
	}, nil

}

func (s *oauthServer) RegisterGRPC(grpcSrv *grpc.Server) {
	pb.RegisterOAuthServer(grpcSrv, s)
}

func (s *oauthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Email == "" || req.Password == "" {
		return &pb.RegisterResponse{
			Ok:      false,
			Message: "email and password required",
		}, nil
	}
	user, err := entity.NewUser(req.Email, req.Password)
	if err != nil {
		return &pb.RegisterResponse{}, err
	}
	user.SetFullName(req.FullName)
	err = user.SetBirthday(req.Birthday)
	if err != nil {
		return &pb.RegisterResponse{}, err
	}
	user.Phone = req.Phone
	if err := s.store.AddUser(user); err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return &pb.RegisterResponse{
				Ok:      false,
				Message: "user exists",
			}, nil
		}
		return &pb.RegisterResponse{
			Ok:      false,
			Message: err.Error(),
		}, nil
	}
	return &pb.RegisterResponse{Ok: true, Message: "created"}, nil
}

func (s *oauthServer) Token(ctx context.Context, req *pb.TokenRequest) (*pb.TokenResponse, error) {
	u, err := s.store.GetUser(req.Email)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(req.Password)); err != nil {
		return nil, err
	}
	access, exp, err := s.jwtManager.GenerateAccessToken(u.Email, u.Role)
	if err != nil {
		return nil, err
	}
	refresh, _, err := s.jwtManager.GenerateRefreshToken(u.Email, u.Role)
	if err != nil {
		return nil, err
	}
	return &pb.TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    exp - time.Now().Unix(),
	}, nil
}

func (s *oauthServer) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.TokenResponse, error) {
	claims, err := s.jwtManager.Parse(req.RefreshToken)
	if err != nil {
		return nil, err
	}
	if t, ok := claims["typ"].(string); !ok || t != "refresh" {
		return nil, errors.New("not a refresh token")
	}
	sub, _ := claims["sub"].(string)
	role, _ := claims["role"].(string)
	access, exp, err := s.jwtManager.GenerateAccessToken(sub, role)
	if err != nil {
		return nil, err
	}

	refresh, _, err := s.jwtManager.GenerateRefreshToken(sub, role)
	if err != nil {
		return nil, err
	}
	return &pb.TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    exp - time.Now().Unix(),
	}, nil
}

func (s *oauthServer) Verify(ctx context.Context, req *pb.VerifyRequest) (*pb.VerifyResponse, error) {

	claims, err := s.jwtManager.Parse(req.AccessToken)
	if err != nil {
		return &pb.VerifyResponse{Valid: false}, nil
	}
	typ, _ := claims["typ"].(string)
	if typ != "access" {
		return &pb.VerifyResponse{Valid: false}, nil
	}
	sub, _ := claims["sub"].(string)
	expF, _ := claims["exp"].(float64)
	return &pb.VerifyResponse{
		Valid: true,
		Email: sub,
		Exp:   int64(expF),
	}, nil
}

func (s *oauthServer) FetchList(ctx context.Context, in *pb.FetchListRequest) (*pb.FetchListResponse, error) {
	pageSize := int64(in.PageSize)
	pageNumber := int64(in.PageNumber)
	if pageNumber <= 0 || pageSize <= 0 {
		return nil, errors.New("page size and number must be positve")
	}
	users, totalCount, err := s.store.GetUserList(pageSize, (pageNumber-1)*pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]*pb.UserItem, 0)
	if len(users) == 0 {
		if pageNumber > 1 {
			return nil, err
		}
		return &pb.FetchListResponse{
			Items:      items,
			TotalCount: 0,
			TotalPages: 1,
			PageNumber: 1,
		}, nil
	}

	for _, user := range users {
		items = append(items, &pb.UserItem{
			Id:       user.Id,
			Email:    user.Email,
			FullName: user.FullName,
			Phone:    user.Phone,
			Birthday: user.Birthday.Format(entity.DateLayout),
			Created:  user.Created.Format(entity.DateLayout),
		})
	}
	totalPages := totalCount / pageSize
	if totalCount > totalPages*pageSize {
		pageSize++
	}
	return &pb.FetchListResponse{
		Items:      items,
		TotalCount: int64(totalCount),
		TotalPages: totalPages,
		PageNumber: int64(in.PageNumber),
	}, nil
}

func (s *oauthServer) DeleteUser(ctx context.Context, req *pb.DeleteUsreRequest) (*pb.DeleteUserResponse, error) {
	userId := req.Id

	err := s.store.DeleteUser(userId)

	if err != nil {
		return &pb.DeleteUserResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}
	return &pb.DeleteUserResponse{
		Message: "Removed",
		Success: true,
	}, nil
}
