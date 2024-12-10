package bot

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/stretchr/testify/mock"
)

type MockTgBotAPI struct {
	mock.Mock
}

func (m *MockTgBotAPI) GetUpdatesChan(u tgbotapi.UpdateConfig) (tgbotapi.UpdatesChannel, error) {
	args := m.Called(u)
	return args.Get(0).(tgbotapi.UpdatesChannel), args.Error(1)
}

func (m *MockTgBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	args := m.Called(c)
	return args.Get(0).(tgbotapi.Message), args.Error(1)
}

func (m *MockTgBotAPI) GetChatAdministrators(config tgbotapi.ChatConfig) ([]tgbotapi.ChatMember, error) {
	args := m.Called(config)
	return args.Get(0).([]tgbotapi.ChatMember), args.Error(1)
}
