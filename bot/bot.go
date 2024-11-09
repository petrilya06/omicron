package bot

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func createInlineKeyboard(activatedButtons []bool) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for i := 0; i < 6; i++ {
		emoji := EmojiEnable
		if !activatedButtons[i] {
			emoji = EmojiDisable
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d - %s", i+1, emoji), strconv.Itoa(i)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Продолжить", "continue"),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func RunBot() {
	var userID int
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
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			switch update.Message.Text {
			case "/start":
				msg.Text = StartMessage
				msg.ReplyMarkup = createInlineKeyboard(criterias)

				userID = int(update.Message.From.ID)
				m.AddUser(userID)
			}
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		} else if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				panic(err)
			}

			switch update.CallbackQuery.Data {
			case "0", "1", "2", "3", "4", "5":
				i, _ := strconv.Atoi(update.CallbackQuery.Data)
				criterias[i] = !criterias[i] // Переключаем состояние кнопки

				fmt.Println("\n\n\n\n\n", update.CallbackQuery.Data, criterias)
				editMsg := tgbotapi.NewEditMessageReplyMarkup(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					createInlineKeyboard(criterias),
				)

				if _, err := bot.Send(editMsg); err != nil {
					log.Println("Error sending edit message:", err)
				}

			case "continue":
				m.SetCriterias(userID, criterias)
				users, err := m.GetAllUsers()
				if err != nil {
					log.Fatal(err)
				}
				for _, user := range users {
					log.Printf("User ID: %d, Data: %v", user.TgID, user.Data)
				}
			}
		}
	}
}
