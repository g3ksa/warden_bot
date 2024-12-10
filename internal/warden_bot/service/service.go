package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/g3ksa/warden_bot/internal/warden_bot/service/model"
	"github.com/g3ksa/warden_bot/internal/warden_bot/service/report"
	"github.com/g3ksa/warden_bot/internal/warden_bot/service/state"
	"github.com/g3ksa/warden_bot/internal/warden_bot/service/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramBotAPI interface {
	GetUpdatesChan(u tgbotapi.UpdateConfig) (tgbotapi.UpdatesChannel, error)
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	GetChatAdministrators(config tgbotapi.ChatConfig) ([]tgbotapi.ChatMember, error)
}

type WardenBotService struct {
	tgBot           TelegramBotAPI
	storage         storage.Storage
	modelServiceUrl string
	botState        *state.BotState
	reportGenerator *report.ReportGenerator
}

func NewWardenBotService(modelServiceUrl string, bot TelegramBotAPI, storage storage.Storage) *WardenBotService {
	return &WardenBotService{
		tgBot:           bot,
		storage:         storage,
		modelServiceUrl: modelServiceUrl,
		botState:        state.NewBotState(),
		reportGenerator: report.NewReportGenerator(storage),
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
				ChatID: uint64(math.Abs(float64(update.Message.Chat.ID))),
				Title:  update.Message.Chat.Title,
				Type:   update.Message.Chat.Type,
			})
			if update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup() {
				msg := &model.Message{
					MessageID:    uint64(update.Message.MessageID),
					UserFullName: fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName),
					Text:         strings.ReplaceAll(update.Message.Text, "\n", " "),
					Date:         time.Unix(int64(update.Message.Date), 0),
					Label:        0,
					ChatID:       uint64(math.Abs(float64(update.Message.Chat.ID))),
				}

				err := s.SaveMessage(ctx, msg)
				if err != nil {
					log.Printf("Failed to save message to database: %v", err)
				}
			} else if update.Message.Chat.IsPrivate() {

				userID := update.Message.From.ID

				switch update.Message.Command() {
				case "report":
					s.processReportCommand(ctx, update.Message, userID)
				case "help":
					s.processHelpCommand(ctx, update.Message.Chat.ID)
				default:
					currentState, exists := s.botState.GetUserState(userID)

					if exists {
						selectedText := update.Message.Text

						var chatID uint64
						_, err := fmt.Sscanf(selectedText, "%d:", &chatID)
						if err != nil {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неправильный выбор. Попробуйте снова.")
							s.tgBot.Send(msg)
							continue
						}

						s.botState.ClearUserState(userID)

						report, err := s.reportGenerator.GenerateReport(ctx, chatID, currentState)
						if err != nil {
							log.Printf("Failed to generate report: %v", err)
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%s %s", "Произошла ошибка при генерации отчета.", err.Error()))
							s.tgBot.Send(msg)
							continue
						}

						reportMsg := report.String()
						chatId := update.Message.Chat.ID

						msg := tgbotapi.NewMessage(chatId, reportMsg)
						msg.ParseMode = "Markdown"
						msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
						s.tgBot.Send(msg)
					}
				}
			}
		}
	}

	return nil
}

func (s *WardenBotService) processHelpCommand(ctx context.Context, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Доступные команды:\n\n"+
		"/report [YYYY-MM-DD] - создать отчет (по умолчанию за предыдущий день)\n"+
		"/help - помощь\n"+
		"Contact: @nit3bo1")
	s.tgBot.Send(msg)
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

	commandArgs := strings.Split(message.CommandArguments(), " ")[0]

	var reportDate time.Time

	if commandArgs != "" {
		reportDate, err = time.Parse("2006-01-02", commandArgs)
		if err != nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Некорректный формат даты. Используйте формат YYYY-MM-DD.")
			s.tgBot.Send(msg)
			return
		}
	} else {
		reportDate = time.Now().Truncate(24 * time.Hour)
	}

	s.botState.SetUserState(userID, reportDate)

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
	chats, err := s.storage.GetGroupChats(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch chats: %w", err)
	}

	for _, chat := range chats {
		messages, err := s.storage.GetMessagesForLastDayByChat(ctx, chat.ChatID)
		if err != nil {
			return err
		}

		if len(messages) == 0 {
			continue
		}

		var messageRequests []model.MessageRequest
		for _, msg := range messages {
			messageRequests = append(messageRequests, model.MessageRequest{
				Text:      msg.Text,
				MessageID: msg.MessageID,
				ChatID:    msg.ChatID,
			})
		}

		messagesToUpdate, err := s.RequestToModel(ctx, messageRequests)
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		err = s.storage.UpdateMessages(ctx, messagesToUpdate)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *WardenBotService) RequestToModel(ctx context.Context, messages []model.MessageRequest) ([]*model.Message, error) {
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
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(s.modelServiceUrl+"/classify", "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to send classification request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("classification API returned non-200 status: %s", resp.Status)
	}

	var classifiedResponse model.ClassifiedMessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&classifiedResponse); err != nil {
		return nil, fmt.Errorf("failed to decode classification response: %w", err)
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

	return messagesToUpdate, nil
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
