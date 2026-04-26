package ws

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mathalama/nektokz/chat-service/internal/domain"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var in ClientMessage
		if err := json.Unmarshal(message, &in); err != nil {
			c.sendError("BAD_REQUEST", "invalid json")
			continue
		}

		switch in.Type {
		case "ping":
			c.sendJSON(ServerMessage{Type: "pong", Timestamp: time.Now().Unix()})

		case "typing":
			c.hub.sendToRoom(c.RoomID, func(recipientID string) ServerMessage {
				if recipientID == c.UserID {
					return ServerMessage{Type: ""}
				}
				return ServerMessage{Type: "partner_typing", IsTyping: in.IsTyping, Timestamp: time.Now().Unix()}
			})

		case "message":
			if in.Content == "" {
				continue
			}

			toxic, err := c.isToxic(context.Background(), in.Content)
			if err != nil {
				c.sendError("MODERATION_ERROR", "failed to moderate message")
				continue
			}
			if toxic {
				c.sendError("TOXIC_MESSAGE", "message rejected")
				continue
			}

			if err := c.repo.SaveMessage(context.Background(), &domain.Message{
				RoomID:   c.RoomID,
				SenderID: c.UserID,
				Content:  in.Content,
				SentAt:   time.Now(),
			}); err != nil {
				c.sendError("DB_ERROR", "failed to save message")
				continue
			}

			ts := time.Now().Unix()
			c.hub.sendToRoom(c.RoomID, func(recipientID string) ServerMessage {
				if recipientID == "" {
					return ServerMessage{Type: ""}
				}
				sender := "partner"
				if recipientID == c.UserID {
					sender = "me"
				}
				return ServerMessage{Type: "message", Content: in.Content, Sender: sender, Timestamp: ts}
			})

		case "next":
			if err := c.endRoom(context.Background()); err != nil {
				c.sendError("ROOM_END_ERROR", err.Error())
				continue
			}
			c.hub.DisconnectRoom(c.RoomID)

		default:
			c.sendError("UNKNOWN_TYPE", "unknown message type")
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendJSON(msg ServerMessage) {
	if msg.Type == "" {
		return
	}
	b, _ := json.Marshal(msg)
	select {
	case c.send <- b:
	default:
	}
}

func (c *Client) sendError(code, message string) {
	c.sendJSON(ServerMessage{Type: "error", Code: code, Message: message, Timestamp: time.Now().Unix()})
}

func (c *Client) isToxic(ctx context.Context, content string) (bool, error) {
	if c.moderation == nil {
		return false, nil
	}
	return c.moderation.IsToxic(ctx, content)
}

func (c *Client) endRoom(ctx context.Context) error {
	room, err := c.repo.GetRoom(ctx, c.RoomID)
	if err != nil {
		return err
	}
	if room == nil {
		return errors.New("room not found")
	}
	now := time.Now()
	room.Status = "ended"
	room.EndedAt = &now
	return c.repo.UpdateRoom(ctx, room)
}
