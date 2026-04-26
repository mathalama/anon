package client

import (
	"context"
)

type UserClient struct {
	baseURL string
}

func NewUserClient(baseURL string) *UserClient {
	return &UserClient{baseURL: baseURL}
}

func (c *UserClient) IsBanned(ctx context.Context, userID string) (bool, error) {
	// Stub: in a real app, this would make an HTTP request to user-service
	return false, nil
}

type ChatClient struct {
	baseURL string
}

func NewChatClient(baseURL string) *ChatClient {
	return &ChatClient{baseURL: baseURL}
}

func (c *ChatClient) CreateRoom(ctx context.Context, roomID string, userA, userB string) error {
	// Stub: in a real app, this would make an HTTP request to chat-service
	return nil
}
