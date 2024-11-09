package bot

import (
	"database/sql"
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
	var urls []string
	var criterias = []bool{true, true, true, true, true, true}
	db, _ := sql.Open("sqlite3", "./users.db")
	m, err := NewSQLMap(db)
	if err != nil {
		panic(err)
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		panic(err)
	}

	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil {
			userID := int(update.Message.From.ID)

			// Инициализация пользователя, если он новый
			if _, exists := users[userID]; !exists {
				users[userID] = &User{ID: userID, State: StateStart}
			}

			user := users[userID]
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			if update.Message.Text == "/start" {
				msg.Text = StartMessage
				msg.ReplyMarkup = CreateInlineKeyboard(criterias)
				user.State = StateNone // Состояние остается тем же
				m.AddUser(userID)
			}

			switch user.State {
			default:
				if update.Message.Document != nil {
					// Обработка документа
					documentURLs, err := ExtractURLsFromFile(update.Message.Document.FileID, bot)
					if err == nil {
						urls = append(urls, documentURLs...)
						fmt.Println(urls)
					} else {
						log.Println("Error extracting URLs from file:", err)
					}
				} else {
					msg := strings.Split(update.Message.Text, " ")

					for _, word := range msg {
						if IsValidURL(word) {
							urls = append(urls, word)
						}
					}
				}
			}

			// Проверяем, что текст сообщения не пустой перед отправкой
			if msg.Text != "" {
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
			}
		} else if update.CallbackQuery != nil {
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

				criterias[i-1] = !criterias[i-1]
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
	}
}
