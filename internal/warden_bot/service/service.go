package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/g3ksa/warden_bot/internal/warden_bot/service/model"
	"github.com/g3ksa/warden_bot/internal/warden_bot/service/state"
	"github.com/g3ksa/warden_bot/internal/warden_bot/service/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type WardenBotService struct {
	tgBot           *tgbotapi.BotAPI
	storage         *storage.DBStorage
	modelServiceUrl string
	botState        state.BotState
}

func NewWardenBotService(modelServiceUrl string, bot *tgbotapi.BotAPI, storage *storage.DBStorage) *WardenBotService {
	return &WardenBotService{
		tgBot:           bot,
		storage:         storage,
		modelServiceUrl: modelServiceUrl,
		botState:        *state.NewBotState(),
	}
}

func (s *WardenBotService) ProcessUpdatesFromBot(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := s.tgBot.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	for update := range updates {
		if update.Message != nil {
			s.storage.SaveChatInfo(ctx, &model.Chat{
				ChatID: uint64(math.Abs((float64(update.Message.Chat.ID)))),
				Title:  update.Message.Chat.Title,
				Type:   update.Message.Chat.Type,
			})
			if update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup() {
				msg := &model.Message{
					MessageID:    uint64(update.Message.MessageID),
					UserFullName: fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName),
					Text:         update.Message.Text,
					Date:         time.Unix(int64(update.Message.Date), 0),
					Label:        0,
					ChatID:       uint64(math.Abs((float64(update.Message.Chat.ID)))),
				}

				err := s.SaveMessage(ctx, msg)
				if err != nil {
					log.Printf("Failed to save message to database: %v", err)
				}
			} else if update.Message.Chat.IsPrivate() {

				userID := update.Message.From.ID

				switch update.Message.Text {
				case "/report":
					s.processReportCommand(ctx, update.Message, userID)
				default:
					currentState, exists := s.botState.GetUserState(userID)

					if exists && currentState == "awaiting_chat_selection" {
						selectedText := update.Message.Text

						var chatID uint64
						_, err := fmt.Sscanf(selectedText, "%d:", &chatID)
						if err != nil {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неправильный выбор. Попробуйте снова.")
							s.tgBot.Send(msg)
							continue
						}

						s.botState.ClearUserState(userID)

						msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Вы выбрали чат: %d", chatID))
						msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
						s.tgBot.Send(msg)
					}
				}
			}
		}
	}

	return nil
}

func (s *WardenBotService) processReportCommand(ctx context.Context, message *tgbotapi.Message, userID int) {
	adminChats, err := s.GetAdminChats(ctx, message.From.ID)
	if err != nil {
		log.Printf("Failed to get admin chats: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Произошла ошибка при получении списка чатов.")
		s.tgBot.Send(msg)
		return
	}

	var buttons []tgbotapi.KeyboardButton
	for _, chat := range adminChats {
		buttons = append(buttons, tgbotapi.NewKeyboardButton(fmt.Sprintf("%d:%s", chat.ChatID, chat.Title)))
	}

	if len(buttons) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Вы не являетесь администратором ни в одном групповом чате.")
		s.tgBot.Send(msg)
		return
	}

	s.botState.SetUserState(userID, "awaiting_chat_selection")

	keyboard := tgbotapi.NewReplyKeyboard(buttons)
	keyboard.OneTimeKeyboard = true

	msg := tgbotapi.NewMessage(message.Chat.ID, "Выберите чат:")
	msg.ReplyMarkup = keyboard
	s.tgBot.Send(msg)
}

func (s *WardenBotService) SaveMessage(ctx context.Context, msg *model.Message) error {
	err := s.storage.PutMessage(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to save message: %s", err.Error())
	}
	return nil
}

func (s *WardenBotService) ProcessMessages(ctx context.Context) error {
	messages, err := s.storage.GetMessagesForLastDay(ctx)
	if err != nil {
		return err
	}

	var messageRequests []model.MessageRequest
	for _, msg := range messages {
		messageRequests = append(messageRequests, model.MessageRequest{
			Text:      msg.Text,
			MessageID: msg.MessageID,
			ChatID:    msg.ChatID,
		})
	}

	requestBody := model.MessagesRequest{Messages: messageRequests}
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(s.modelServiceUrl+"/classify", "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		return fmt.Errorf("failed to send classification request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("classification API returned non-200 status: %s", resp.Status)
	}

	var classifiedResponse model.ClassifiedMessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&classifiedResponse); err != nil {
		return fmt.Errorf("failed to decode classification response: %w", err)
	}

	messagesToUpdate := make([]*model.Message, 0, len(classifiedResponse.Messages))
	for _, message := range classifiedResponse.Messages {
		fmt.Printf("%+v", message)
		messagesToUpdate = append(messagesToUpdate, &model.Message{
			MessageID: message.MessageID,
			Label:     message.Label,
			ChatID:    message.ChatID,
		})
	}

	err = s.storage.UpdateMessages(ctx, messagesToUpdate)
	if err != nil {
		return err
	}

	return nil
}

func (s *WardenBotService) GetAdminChats(ctx context.Context, userID int) ([]model.Chat, error) {
	groupChats, err := s.storage.GetGroupChats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group chats: %w", err)
	}

	adminChats := make([]model.Chat, 0)

	for _, chat := range groupChats {
		chatAdmins, err := s.tgBot.GetChatAdministrators(tgbotapi.ChatConfig{ChatID: -int64(chat.ChatID)})
		if err != nil {
			log.Printf("Failed to fetch administrators for chat %d: %v", chat.ChatID, err)
			continue
		}

		for _, admin := range chatAdmins {
			if admin.User.ID == userID {
				adminChats = append(adminChats, chat)
				break
			}
		}
	}

	return adminChats, nil
}
