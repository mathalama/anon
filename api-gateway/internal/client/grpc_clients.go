package client

import (
	pbUser "github.com/mathalama/nektokz/proto/user/v1"
	pbMatch "github.com/mathalama/nektokz/proto/matchmaking/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"github.com/sony/gobreaker"
	"time"
)

type GRPCClients struct {
	User pbUser.UserServiceClient
	Match pbMatch.MatchmakingServiceClient
	
	// Circuit Breakers
	UserBreaker *gobreaker.CircuitBreaker
	MatchBreaker *gobreaker.CircuitBreaker
}

func NewGRPCClients(userAddr, matchAddr string) (*GRPCClients, error) {
	uConn, err := grpc.Dial(userAddr, 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, err
	}
	mConn, err := grpc.Dial(matchAddr, 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, err
	}
	
	// Init Circuit Breakers
	userBreaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "user-service",
		MaxRequests: 5,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.6
		},
	})

	matchBreaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "matchmaking-service",
		MaxRequests: 5,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.6
		},
	})

	return &GRPCClients{
		User: pbUser.NewUserServiceClient(uConn),
		Match: pbMatch.NewMatchmakingServiceClient(mConn),
		UserBreaker: userBreaker,
		MatchBreaker: matchBreaker,
	}, nil
}
