package client

import (
	pbUser "github.com/mathalama/nektokz/proto/user/v1"
	pbMatch "github.com/mathalama/nektokz/proto/matchmaking/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClients struct {
	User pbUser.UserServiceClient
	Match pbMatch.MatchmakingServiceClient
}

func NewGRPCClients(userAddr, matchAddr string) (*GRPCClients, error) {
	uConn, err := grpc.Dial(userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	mConn, err := grpc.Dial(matchAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	
	return &GRPCClients{
		User: pbUser.NewUserServiceClient(uConn),
		Match: pbMatch.NewMatchmakingServiceClient(mConn),
	}, nil
}
