package bot

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

	//TODO: find & get user. Create if not exist

	if update.Message.Text == "/start" {
		msg.Text = StartMessage
		msg.ReplyMarkup = CreateInlineKeyboard(criterias) // from bd
		user.State = StateNone                            // Состояние остается тем же
		return
	}

	switch user.State {
	default:
		if update.Message.Document != nil {
			// Обработка документа
			documentURLs, err := ExtractURLsFromFile(update.Message.Document.FileID, bot)
			if err == nil {
				fmt.Println(documentURLs)
			} else {
				log.Println("Error extracting URLs from file:", err)
			}
		} else {
			msg := strings.Split(update.Message.Text, " ")

			for _, word := range msg {
				if IsValidURL(word) {
					// word = url
				}
			}
		}
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

	//user := users[userID]
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "")

	switch update.CallbackQuery.Data {
	case "1", "2", "3", "4", "5", "6":
		i, _ := strconv.Atoi(update.CallbackQuery.Data)

		criterias[i-1] = !criterias[i-1] // in bd
		editMsg := tgbotapi.NewEditMessageReplyMarkup(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			CreateInlineKeyboard(criterias), // new value from bd
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
