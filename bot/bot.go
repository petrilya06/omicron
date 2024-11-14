package bot

import (
	"fmt"
	"log"
	"omicron/db"
	"os"
	"strconv"

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
	db.MustInitDB()

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

		// Добавляем пользователя в базу данных
		if err := db.DBMap.AddUser(userID); err != nil {
			log.Println("Error adding user to DB:", err)
			return
		}
	}

	user := users[userID]
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	criterias := []bool{true, true, true, true, true, true}

	a := db.DBMap.PrintAllUsers()
	fmt.Println(a)

	if update.Message.Text == "/start" {
		// Если пользователь только что добавлен, можно инициализировать его критерии
		if user.State == StateStart {
			// Здесь можно установить начальные критерии, если это необходимо
			if err := db.DBMap.SetCriterias(userID, criterias); err != nil {
				log.Println("Error setting criteria for user:", err)
				return
			}
		}

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

	criterias, err := db.DBMap.GetCriteriasByUserID(userID) // Получаем текущие критерии
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
		if err := db.DBMap.UpdateDataByUserID(userID, criterias); err != nil {
			log.Println("Error updating data in DB:", err)
			return
		}

		// Обновляем клавиатуру с новыми значениями
		editMsg := tgbotapi.NewEditMessageReplyMarkup(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			CreateInlineKeyboard(criterias),
		)

		a := db.DBMap.PrintAllUsers()
		fmt.Println("\n\n\n\n\n\nAll users:", a)

		if _, err := bot.Send(editMsg); err != nil {
			log.Println("Error sending edit message:", err)
		}
	}
}
