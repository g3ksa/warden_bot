package report

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/g3ksa/warden_bot/internal/warden_bot/service/storage"
)

type ReportGenerator struct {
	storage storage.Storage
}

func NewReportGenerator(storage storage.Storage) *ReportGenerator {
	return &ReportGenerator{
		storage,
	}
}

type Report struct {
	Date                       time.Time
	ChatID                     uint64
	ChatTitle                  string
	TotalMessages              int
	ProductiveMessages         int
	UnproductiveMessages       int
	UnproductivePercentage     float64
	UnproductiveMessageSamples []string
	TopDistractingUsers        []UserActivity
	ActivityTimeline           []ActivityPoint
	ProductivityIndicator      string
}

type UserActivity struct {
	UserName string
	Count    int
}

type ActivityPoint struct {
	Timestamp time.Time
	Count     int
}

func (g *ReportGenerator) GenerateReport(ctx context.Context, chatID uint64, date time.Time) (*Report, error) {
	// Получение всех сообщений за период
	messages, err := g.storage.GetMessagesByChatAndPeriod(ctx, chatID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("В чате не было сообщений за запрашиваемую дату")
	}

	chatInfo, err := g.storage.GetChatInfoByID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chat info: %w", err)
	}

	// Инициализация переменных для анализа
	totalMessages := len(messages)
	productiveMessages := 0
	unproductiveMessages := 0
	unproductiveSamples := []string{}
	userActivity := make(map[string]int)
	timeline := make(map[time.Time]int)

	// Анализ сообщений
	for _, msg := range messages {
		if msg.Label == 1 { // Label 1 - продуктивное сообщение
			productiveMessages++
		} else if msg.Label == 0 { // Label 0 - непродуктивное сообщение
			unproductiveMessages++
			unproductiveSamples = append(unproductiveSamples, msg.Text)
			userActivity[msg.UserFullName]++
			timeline[msg.Date.Truncate(time.Hour)]++ // Группировка по часам
		}
	}

	// Расчет процента непродуктивных сообщений
	unproductivePercentage := 0.0
	if totalMessages > 0 {
		unproductivePercentage = float64(unproductiveMessages) / float64(totalMessages) * 100
	}

	// Сортировка пользователей с наибольшим количеством непродуктивных сообщений
	topUsers := []UserActivity{}
	for user, count := range userActivity {
		topUsers = append(topUsers, UserActivity{UserName: user, Count: count})
	}
	// Сортировка по убыванию
	sort.Slice(topUsers, func(i, j int) bool {
		return topUsers[i].Count > topUsers[j].Count
	})

	// Подготовка временной активности
	activityTimeline := []ActivityPoint{}
	for timestamp, count := range timeline {
		activityTimeline = append(activityTimeline, ActivityPoint{Timestamp: timestamp, Count: count})
	}
	// Сортировка по времени
	sort.Slice(activityTimeline, func(i, j int) bool {
		return activityTimeline[i].Timestamp.Before(activityTimeline[j].Timestamp)
	})

	// Индикатор продуктивности
	productivityIndicator := "Нормально"
	if unproductivePercentage > 50 {
		productivityIndicator = "Низкая продуктивность"
	} else if unproductivePercentage < 20 {
		productivityIndicator = "Высокая продуктивность"
	}

	// Создание отчета
	report := &Report{
		Date:                       date,
		ChatID:                     chatID,
		ChatTitle:                  chatInfo.Title,
		TotalMessages:              totalMessages,
		ProductiveMessages:         productiveMessages,
		UnproductiveMessages:       unproductiveMessages,
		UnproductivePercentage:     unproductivePercentage,
		UnproductiveMessageSamples: unproductiveSamples,
		TopDistractingUsers:        topUsers,
		ActivityTimeline:           activityTimeline,
		ProductivityIndicator:      productivityIndicator,
	}

	return report, nil
}

func (r *Report) String() string {
	reportText := fmt.Sprintf(
		"📊 *Отчет по чату*: %s за %s\n\n"+
			"📋 *Общая статистика:*\n"+
			"Всего сообщений: %d\n"+
			"Продуктивных сообщений: %d\n"+
			"Непродуктивных сообщений: %d\n"+
			"Процент непродуктивных сообщений: %.2f%%\n\n"+
			"🚫 *Примеры непродуктивных сообщений:*\n%s\n\n"+
			"👥 *Пользователи с наибольшим количеством непродуктивных сообщений:*\n%s\n\n"+
			"📈 *Активность непродуктивных сообщений по времени:*\n%s\n\n"+
			"📌 *Индикатор продуктивности*: %s",
		r.ChatTitle,
		r.Date.Format("02.01.2006"),
		r.TotalMessages,
		r.ProductiveMessages,
		r.UnproductiveMessages,
		r.UnproductivePercentage,
		formatExamples(r.UnproductiveMessageSamples),
		formatTopUsers(r.TopDistractingUsers),
		formatActivityTimeline(r.ActivityTimeline),
		r.ProductivityIndicator,
	)

	return reportText
}

func formatExamples(examples []string) string {
	if len(examples) == 0 {
		return "Нет примеров."
	}
	result := ""
	for i, example := range examples {
		if i >= 10 {
			result += "и другие...\n"
			break
		}
		result += fmt.Sprintf("- %s\n", example)
	}
	return result
}

func formatTopUsers(users []UserActivity) string {
	if len(users) == 0 {
		return "Нет пользователей с непродуктивными сообщениями."
	}
	result := ""
	for _, user := range users {
		result += fmt.Sprintf("- %s: %d сообщений\n", user.UserName, user.Count)
	}
	return result
}

func formatActivityTimeline(timeline []ActivityPoint) string {
	if len(timeline) == 0 {
		return "Нет данных о временной активности."
	}
	result := ""
	for _, point := range timeline {
		result += fmt.Sprintf("- %s: %d сообщений\n", point.Timestamp.Format("01.02.2006 15:00"), point.Count)
	}
	return result
}
