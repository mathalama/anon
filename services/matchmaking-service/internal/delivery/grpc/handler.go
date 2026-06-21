package grpc

import (
	"context"
	"github.com/mathalama/nektokz/matchmaking-service/internal/domain"
	pb "github.com/mathalama/nektokz/proto/matchmaking/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MatchmakingHandler struct {
	pb.UnimplementedMatchmakingServiceServer
	usecase domain.MatchUsecase
}

func NewMatchmakingHandler(usecase domain.MatchUsecase) *MatchmakingHandler {
	return &MatchmakingHandler{
		usecase: usecase,
	}
}

func (h *MatchmakingHandler) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	filter := domain.Filter{
		Mode:      req.Mode,
		Gender:    req.Gender,
		Interests: req.Interests,
		RoomTopic: req.RoomTopic,
	}
	// Note: we might need MyGender here, but SearchRequest doesn't have it yet.
	// In the real app, we get MyGender from the user profile.
	// For now, let's assume it's part of the filter or fetched internally.

	if err := h.usecase.Search(ctx, req.UserId, filter); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start search: %v", err)
	}

	return &pb.SearchResponse{Success: true}, nil
}

func (h *MatchmakingHandler) Cancel(ctx context.Context, req *pb.CancelRequest) (*pb.CancelResponse, error) {
	if err := h.usecase.Cancel(ctx, req.UserId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to cancel search: %v", err)
	}
	return &pb.CancelResponse{Success: true}, nil
}

func (h *MatchmakingHandler) GetStatus(ctx context.Context, req *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {
	room, err := h.usecase.GetStatus(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get status: %v", err)
	}

	if room == nil {
		return &pb.GetStatusResponse{Status: "searching"}, nil
	}

	partnerID := room.UserA
	if partnerID == req.UserId {
		partnerID = room.UserB
	}

	return &pb.GetStatusResponse{
		RoomId:    room.ID,
		PartnerId: partnerID,
		Status:    "matched",
	}, nil
}
