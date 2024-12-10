package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/g3ksa/warden_bot/internal/warden_bot/service/model"
	"gorm.io/gorm"
)

type Storage interface {
	PutMessage(ctx context.Context, message *model.Message) error
	UpdateMessages(ctx context.Context, messages []*model.Message) error
	GetMessagesForLastDayByChat(ctx context.Context, chatID uint64) ([]model.Message, error)
	SaveChatInfo(ctx context.Context, chatInfo *model.Chat) error
	GetGroupChats(ctx context.Context) ([]model.Chat, error)
	GetMessagesByChatAndPeriod(ctx context.Context, chatID uint64, date time.Time) ([]*model.Message, error)
	GetChatInfoByID(ctx context.Context, chatID uint64) (*model.Chat, error)
}

type DBStorage struct {
	db *gorm.DB
}

func NewDBStorage(db *gorm.DB) *DBStorage {
	return &DBStorage{db: db}
}

func (s *DBStorage) PutMessage(ctx context.Context, message *model.Message) error {
	err := s.db.Create(message).Error
	if err != nil {
		return err
	}
	return nil
}

func (s *DBStorage) UpdateMessages(ctx context.Context, messages []*model.Message) error {
	for _, msg := range messages {
		updates := make(map[string]interface{})

		updates["label"] = msg.Label

		if err := s.db.
			WithContext(ctx).
			Model(&model.Message{}).
			Where("message_id = ? and chat_id = ?", msg.MessageID, msg.ChatID).
			Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update message ID %d: %w", msg.MessageID, err)
		}
	}
	return nil
}

func (s *DBStorage) GetMessagesForLastDayByChat(ctx context.Context, chatID uint64) ([]model.Message, error) {
	messages := make([]model.Message, 0)
	err := s.db.WithContext(ctx).Where("date >= now() - interval '1 day' AND chat_id = ?", chatID).Order("date").Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *DBStorage) SaveChatInfo(ctx context.Context, chatInfo *model.Chat) error {
	err := s.db.Create(chatInfo).Error
	if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
		return err
	}
	return nil
}

func (s *DBStorage) GetGroupChats(ctx context.Context) ([]model.Chat, error) {
	chats := make([]model.Chat, 0)
	err := s.db.WithContext(ctx).Where("type LIKE ?", "%group%").Find(&chats).Error
	if err != nil {
		return nil, err
	}
	return chats, nil
}

func (s *DBStorage) GetMessagesByChatAndPeriod(ctx context.Context, chatID uint64, date time.Time) ([]*model.Message, error) {
	messages := make([]*model.Message, 0)

	truncatedDate := date.Truncate(24 * time.Hour)
	dateString := truncatedDate.Format("2006-01-02")

	err := s.db.WithContext(ctx).Where("chat_id = ? and DATE(date) = ?", chatID, dateString).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *DBStorage) GetChatInfoByID(ctx context.Context, chatID uint64) (*model.Chat, error) {
	chat := model.Chat{}
	err := s.db.WithContext(ctx).Where("chat_id = ?", chatID).Find(&chat).Error
	if err != nil {
		return nil, err
	}
	return &chat, nil
}
