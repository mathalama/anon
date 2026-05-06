package grpc

import (
	"context"
	"github.com/mathalama/nektokz/user-service/internal/domain"
	pb "github.com/mathalama/nektokz/proto/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	uc domain.UserUsecase
}

func NewUserHandler(uc domain.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := h.uc.GetMe(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}
	return &pb.GetUserResponse{
		Id:       user.ID,
		Username: user.Email, // or some other display name
		Gender:   user.Gender,
	}, nil
}

func (h *UserHandler) IsBanned(ctx context.Context, req *pb.IsBannedRequest) (*pb.IsBannedResponse, error) {
	_, banned, err := h.uc.GetBanStatus(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check ban status: %v", err)
	}
	return &pb.IsBannedResponse{Banned: banned}, nil
}

func (h *UserHandler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	err := h.uc.UpdateMe(ctx, req.UserId, req.Gender, req.Interests)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update profile: %v", err)
	}
	return &pb.UpdateProfileResponse{Success: true}, nil
}
