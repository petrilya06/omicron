package bot

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func IsValidURL(url string) bool {
	re := regexp.MustCompile(`https:\/\/zakupki\.mos\.ru\/auction\/\d{7}`)
	return re.MatchString(url)
}

func ExtractURLsFromText(text string) []string {
	var urls []string
	words := strings.Split(text, " ")
	for _, word := range words {
		if IsValidURL(word) {
			urls = append(urls, word)
		}
	}
	return urls
}

func ExtractURLsFromFile(fileID string, bot *tgbotapi.BotAPI) ([]string, error) {
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return nil, err
	}

	// Скачиваем файл
	resp, err := http.Get(file.Link(bot.Token))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Читаем содержимое файла
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Извлекаем ссылки из текста файла
	return ExtractURLsFromText(string(body)), nil
}

func CreateInlineKeyboard(activatedButtons [6]bool) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for i := 0; i < 2; i++ {
		var row []tgbotapi.InlineKeyboardButton
		for j := 0; j < 3; j++ {
			emoji := EmojiEnable
			if !activatedButtons[i*3+j] {
				emoji = EmojiDisable
			}
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s %d", emoji, i*3+j+1), strconv.Itoa(i*3+j+1)))
		}
		rows = append(rows, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
