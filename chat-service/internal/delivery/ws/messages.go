package ws

type ClientMessage struct {
	Type     string      `json:"type"`
	Content  string      `json:"content,omitempty"`
	IsTyping bool        `json:"is_typing,omitempty"`
	Reason   string      `json:"reason,omitempty"`
	Payload  interface{} `json:"payload,omitempty"`
}

type ServerMessage struct {
	Type      string      `json:"type"`
	Content   string      `json:"content,omitempty"`
	Sender    string      `json:"sender,omitempty"`
	IsTyping  bool        `json:"is_typing,omitempty"`
	RoomID    string      `json:"room_id,omitempty"`
	Timestamp int64       `json:"ts,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`

	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

