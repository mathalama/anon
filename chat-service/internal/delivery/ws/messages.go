package ws

type ClientMessage struct {
	Type     string `json:"type"`
	Content  string `json:"content,omitempty"`
	IsTyping bool   `json:"is_typing,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type ServerMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content,omitempty"`
	Sender    string `json:"sender,omitempty"`
	IsTyping  bool   `json:"is_typing,omitempty"`
	RoomID    string `json:"room_id,omitempty"`
	Timestamp int64  `json:"ts,omitempty"`

	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

