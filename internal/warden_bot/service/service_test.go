package service

import (
	"context"
	"errors"
	"github.com/g3ksa/warden_bot/mocks/bot"
	"github.com/g3ksa/warden_bot/mocks/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"testing"
	"time"

	"github.com/g3ksa/warden_bot/internal/warden_bot/service/model"
	"github.com/stretchr/testify/assert"
)

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

func TestClassifyMessages(t *testing.T) {
	ctx := context.Background()
	testMessage := &model.Message{
		MessageID:    1,
		UserFullName: "John Doe",
		Text:         "Hello, World!",
		Date:         time.Now(),
		Label:        0,
		ChatID:       123,
	}

	t.Run("Big messages", func(t *testing.T) {
		mockStorage, _, wardenBotService := setupTest()

		mockStorage.On("PutMessage", ctx, testMessage).Return(errors.New("mocked error"))
		err := wardenBotService.SaveMessage(ctx, testMessage)

		assert.NoError(t, err)
	})

	t.Run("Small messages", func(t *testing.T) {
		mockStorage, _, wardenBotService := setupTest()

		mockStorage.On("PutMessage", ctx, testMessage).Return(errors.New("connection lost"))
		err := wardenBotService.SaveMessage(ctx, testMessage)

		assert.NoError(t, err)
	})
}

func setupTest() (*storage.MockStorage, *bot.MockTgBotAPI, *WardenBotService) {
	mockStorage := new(storage.MockStorage)
	mockTgBot := new(bot.MockTgBotAPI)
	wardenBotService := &WardenBotService{
		tgBot:   mockTgBot,
		storage: mockStorage,
	}
	return mockStorage, mockTgBot, wardenBotService
}
