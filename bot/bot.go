package bot

import (
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/petrilya06/omicron/db"
)

const (
	StateStart                = 1
	StateWaitingForTextOrFile = 2
	StateNone                 = 3
)

type User struct {
	ID    int
	State int
}

var users = make(map[int]*User)

func RunBot() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil {
			HandleMessage(bot, &update)
			continue
		}

		if update.CallbackQuery != nil {
			HandleCallbackQuery(bot, &update)
			continue
		}
	}
}

func HandleMessage(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	userID := int(update.Message.From.ID)

	// Инициализация пользователя, если он новый
	if _, exists := users[userID]; !exists {
		users[userID] = &User{ID: userID, State: StateStart}
	}

	user := users[userID]
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	criterias := [...]bool{true, true, true, true, true, true}

	if update.Message.Text == "/start" {
		msg.Text = StartMessage
		msg.ReplyMarkup = CreateInlineKeyboard(criterias) // Передаем текущие критерии
		user.State = StateNone
	} else {
		// Обработка других сообщений
	}

	if msg.Text != "" {
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

func HandleCallbackQuery(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
	if _, err := bot.Request(callback); err != nil {
		log.Panic(err)
	}

	userID := int(update.CallbackQuery.From.ID)

	// Инициализация пользователя, если он новый
	if _, exists := users[userID]; !exists {
		users[userID] = &User{ID: userID, State: StateStart}
	}

	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "")
	criterias, err := db.GetDataByUserID(userID) // Получаем текущие критерии
	if err != nil {
		log.Println("Error getting data from DB:", err)
		return
	}

	switch update.CallbackQuery.Data {
	case "1", "2", "3", "4", "5", "6":
		i, _ := strconv.Atoi(update.CallbackQuery.Data)

		// Изменяем критерий
		criterias[i-1] = !criterias[i-1] // Меняем значение критерия

		// Обновляем данные в базе данных
		if err := db.UpdateDataByUserID(userID, criterias); err != nil {
			log.Println("Error updating data in DB:", err)
			return
		}

		// Обновляем клавиатуру с новыми значениями
		editMsg := tgbotapi.NewEditMessageReplyMarkup(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			CreateInlineKeyboard(criterias),
		)

		if _, err := bot.Send(editMsg); err != nil {
			log.Println("Error sending edit message:", err)
		}
	}

	// Проверяем, что текст сообщения не пустой перед отправкой
	if msg.Text != "" {
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}
