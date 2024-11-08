package bot

import (
	"fmt"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func createInlineKeyboard(activatedButtons map[int]bool) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for i := 1; i <= 6; i++ {
		emoji := EmojiEnable
		if !activatedButtons[i] {
			emoji = EmojiDisable
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d - %s", i, emoji), strconv.Itoa(i)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Продолжить", "continue"),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func RunBot() {
	var criterias []byte = []byte{1, 1, 1, 1, 1, 1, 1}
	activatedButtons := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true}

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
				msg.ReplyMarkup = createInlineKeyboard(activatedButtons)
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
			case "1", "2", "3", "4", "5", "6":
				i, _ := strconv.Atoi(update.CallbackQuery.Data)
				criterias[i] = 0
				activatedButtons[i] = !activatedButtons[i] // Переключаем состояние кнопки
				editMsg := tgbotapi.NewEditMessageReplyMarkup(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					createInlineKeyboard(activatedButtons),
				)

				if _, err := bot.Send(editMsg); err != nil {
					log.Println("Error sending edit message:", err)
				}

			case "continue":
				userID := int(update.CallbackQuery.From.ID)
				fmt.Println(userID)
				ConnectDB(userID, criterias)
				GetUsers()
			}
		}
	}
}
