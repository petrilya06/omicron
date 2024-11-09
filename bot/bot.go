package bot

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotState int

const (
	StateStart BotState = iota
	StateWaitingForFile
	StateWaitingForText
)

type User struct {
	ID    int
	State BotState
}

var users = make(map[int]*User)

func downloadFile(fileName string, url string) error {
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func createInlineKeyboard(activatedButtons []bool) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for i := 0; i < 6; i++ {
		emoji := EmojiEnable
		if !activatedButtons[i] {
			emoji = EmojiDisable
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s - %d", emoji, i+1), strconv.Itoa(i+1)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Отправить файлом", "file"),
		tgbotapi.NewInlineKeyboardButtonData("Отправить текстом", "text"),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func RunBot() {
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

			switch user.State {
			case StateStart:
				if update.Message.Text == "/start" {
					msg.Text = StartMessage
					msg.ReplyMarkup = createInlineKeyboard([]bool{true, true, true, true, true, true})
					user.State = StateStart // Состояние остается тем же
					m.AddUser(userID)
				} else if update.Message.Text == "Отправить текстом" {
					msg.Text = "Введите ваши ссылки без лишнего текста:"
					user.State = StateWaitingForText
				} else if update.Message.Text == "Отправить файлом" {
					msg.Text = "Загрузите ваш файл:"
					user.State = StateWaitingForFile
				}

			case StateWaitingForText:
				// Обработка текста
				msg.Text = "Вы ввели: " + update.Message.Text
				user.State = StateStart // Возвращаемся в начальное состояние

			case StateWaitingForFile:
				// Обработка файла
				if update.Message.Document != nil {
					fileID := update.Message.Document.FileID
					fileName := update.Message.Document.FileName

					fileConfig := tgbotapi.FileConfig{FileID: fileID}
					file, err := bot.GetFile(fileConfig)
					if err != nil {
						log.Println("Error getting file:", err)
						continue
					}

					fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", bot.Token, file.FilePath)
					err = downloadFile(fileName, fileURL)
					if err != nil {
						log.Println("Error downloading file:", err)
						continue
					}

					msg.Text = "Файл успешно сохранен: " + fileName
					user.State = StateStart // Возвращаемся в начальное состояние
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

			user := users[userID]
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "")

			switch update.CallbackQuery.Data {
			case "1", "2", "3", "4", "5", "6":
				i, _ := strconv.Atoi(update.CallbackQuery.Data)

				criterias[i-1] = !criterias[i-1]
				editMsg := tgbotapi.NewEditMessageReplyMarkup(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					createInlineKeyboard(criterias),
				)

				if _, err := bot.Send(editMsg); err != nil {
					log.Println("Error sending edit message:", err)
				}

			case "text":
				m.SetCriterias(userID, criterias)
				fmt.Println("\n\n\n\n", userID, criterias)
				msg.Text = "Введите ваши ссылки без лишнего текста:"
				user.State = StateWaitingForText

			case "file":
				m.SetCriterias(userID, criterias)
				fmt.Println("\n\n\n\n", userID, criterias)
				msg.Text = "Загрузите ваш файл:"
				user.State = StateWaitingForFile
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
