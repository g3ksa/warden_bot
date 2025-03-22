package storage

import (
	"context"
	"github.com/g3ksa/warden_bot/internal/warden_bot/service/model"
	"github.com/stretchr/testify/mock"
	"time"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) PutMessage(ctx context.Context, message *model.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockStorage) UpdateMessages(ctx context.Context, messages []*model.Message) error {
	args := m.Called(ctx, messages)
	return args.Error(0)
}

func (m *MockStorage) GetMessagesForLastDayByChat(ctx context.Context, chatID uint64) ([]model.Message, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).([]model.Message), args.Error(1)
}

func (m *MockStorage) SaveChatInfo(ctx context.Context, chatInfo *model.Chat) error {
	args := m.Called(ctx, chatInfo)
	return args.Error(0)
}

func (m *MockStorage) GetGroupChats(ctx context.Context) ([]model.Chat, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Chat), args.Error(1)
}

func (m *MockStorage) GetMessagesByChatAndPeriod(ctx context.Context, chatID uint64, date time.Time) ([]*model.Message, error) {
	args := m.Called(ctx, chatID, date)
	return args.Get(0).([]*model.Message), args.Error(1)
}

func (m *MockStorage) GetChatInfoByID(ctx context.Context, chatID uint64) (*model.Chat, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).(*model.Chat), args.Error(1)
}
