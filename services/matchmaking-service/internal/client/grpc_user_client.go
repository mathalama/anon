package client

import (
	"context"
	pb "github.com/mathalama/nektokz/proto/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCUserClient struct {
	client pb.UserServiceClient
}

func NewGRPCUserClient(addr string) (*GRPCUserClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &GRPCUserClient{
		client: pb.NewUserServiceClient(conn),
	}, nil
}

func (c *GRPCUserClient) IsBanned(ctx context.Context, userID string) (bool, error) {
	resp, err := c.client.IsBanned(ctx, &pb.IsBannedRequest{UserId: userID})
	if err != nil {
		return false, err
	}
	return resp.Banned, nil
}
