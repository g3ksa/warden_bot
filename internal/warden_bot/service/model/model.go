package model

import "time"

type Message struct {
	MessageID    uint64    `json:"messageId"`
	UserFullName string    `json:"userName"`
	Text         string    `json:"text"`
	Date         time.Time `json:"date"`
	Label        uint      `json:"label"`
	ChatID       uint64    `json:"chatId"`
	Chat         Chat      `json:"chat" gorm:"foreignKey:ChatID;references:ChatID"`
}

func (m *Message) TableName() string {
	return "messages"
}

type Chat struct {
	ChatID   uint64    `json:"chatId" gorm:"primaryKey;autoIncrement"`
	Type     string    `json:"type" gorm:"type:varchar(50)"`
	Title    string    `json:"title" gorm:"type:varchar(255)"`
	Messages []Message `json:"messages" gorm:"foreignKey:ChatID;references:ChatID"`
}

func (ch *Chat) TableName() string {
	return "chats"
}

type MessageRequest struct {
	MessageID uint64 `json:"message_id"`
	Text      string `json:"text"`
	ChatID    uint64 `json:"chat_id"`
}

type MessagesRequest struct {
	Messages []MessageRequest `json:"messages"`
}

type ClassifiedMessage struct {
	MessageID uint64 `json:"message_id"`
	Text      string `json:"text"`
	Label     uint   `json:"label"`
	ChatID    uint64 `json:"chat_id"`
}

type ClassifiedMessagesResponse struct {
	Messages []ClassifiedMessage `json:"messages"`
}
