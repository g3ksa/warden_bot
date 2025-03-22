package service

import (
	"context"
	"errors"
	"github.com/g3ksa/warden_bot/mocks/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"

	"github.com/g3ksa/warden_bot/internal/warden_bot/service/model"
	"github.com/stretchr/testify/assert"
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

func TestSaveMessage(t *testing.T) {
	ctx := context.Background()
	testMessage := &model.Message{
		MessageID:    1,
		UserFullName: "John Doe",
		Text:         "Hello, World!",
		Date:         time.Now(),
		Label:        0,
		ChatID:       123,
	}

	t.Run("Success", func(t *testing.T) {
		mockStorage, _, wardenBotService := setupTest()

		mockStorage.On("PutMessage", ctx, testMessage).Return(nil)
		err := wardenBotService.SaveMessage(ctx, testMessage)
		assert.NoError(t, err)

		mockStorage.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockStorage, _, wardenBotService := setupTest()

		mockStorage.On("PutMessage", ctx, testMessage).Return(errors.New("database error"))
		err := wardenBotService.SaveMessage(ctx, testMessage)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save message")

		mockStorage.AssertExpectations(t)
	})
}

func TestGetAdminChats(t *testing.T) {
	ctx := context.Background()
	userID := 123

	t.Run("Success", func(t *testing.T) {
		mockStorage, mockTgBot, wardenBotService := setupTest()

		groupChats := []model.Chat{
			{ChatID: 1, Title: "Chat 1", Type: "group"},
			{ChatID: 2, Title: "Chat 2", Type: "supergroup"},
		}

		chatAdmins := []tgbotapi.ChatMember{
			{User: &tgbotapi.User{ID: userID}},
		}

		mockStorage.On("GetGroupChats", ctx).Return(groupChats, nil)
		mockTgBot.On("GetChatAdministrators", tgbotapi.ChatConfig{ChatID: -1}).Return(chatAdmins, nil)
		mockTgBot.On("GetChatAdministrators", tgbotapi.ChatConfig{ChatID: -2}).Return([]tgbotapi.ChatMember{}, nil)

		adminChats, err := wardenBotService.GetAdminChats(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(adminChats))
		assert.Equal(t, uint64(1), adminChats[0].ChatID)

		mockStorage.AssertExpectations(t)
		mockTgBot.AssertExpectations(t)
	})

	t.Run("Error fetching group chats", func(t *testing.T) {
		mockStorage, _, wardenBotService := setupTest()

		mockStorage.On("GetGroupChats", ctx).Return([]model.Chat{}, errors.New("database error"))

		adminChats, err := wardenBotService.GetAdminChats(ctx, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch group chats")
		assert.Nil(t, adminChats)

		mockStorage.AssertExpectations(t)
	})

	t.Run("Error fetching chat administrators", func(t *testing.T) {
		mockStorage, mockTgBot, wardenBotService := setupTest()

		groupChats := []model.Chat{
			{ChatID: 1, Title: "Chat 1", Type: "group"},
		}

		mockStorage.On("GetGroupChats", ctx).Return(groupChats, nil)
		mockTgBot.On("GetChatAdministrators", tgbotapi.ChatConfig{ChatID: -1}).Return([]tgbotapi.ChatMember{}, errors.New("telegram API error"))

		adminChats, err := wardenBotService.GetAdminChats(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(adminChats))

		mockStorage.AssertExpectations(t)
		mockTgBot.AssertExpectations(t)
	})
}

func setupTest() (*MockStorage, *bot.MockTgBotAPI, *WardenBotService) {
	mockStorage := new(MockStorage)
	mockTgBot := new(bot.MockTgBotAPI)
	wardenBotService := &WardenBotService{
		tgBot:   mockTgBot,
		storage: mockStorage,
	}
	return mockStorage, mockTgBot, wardenBotService
}
