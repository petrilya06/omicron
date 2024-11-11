package db

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type SQLMap struct {
	db *sql.DB
}

var DBMap *SQLMap

func MustInitDB() *SQLMap {
	db, _ := sql.Open("sqlite3", os.Getenv("DB_PATH"))
	m, err := newSQLMap(db)
	if err != nil {
		panic(err)
	}

	DBMap = m

	return m
}

func newSQLMap(db *sql.DB) (*SQLMap, error) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (tg_id integer primary key, data BLOB)`)
	if err != nil {
		return nil, err
	}
	return &SQLMap{db: db}, nil
}

func (m *SQLMap) AddUser(tg_id int) error {
	_, err := m.db.Exec("INSERT OR IGNORE INTO users (tg_id) VALUES (?)", tg_id)
	return err
}

func (m *SQLMap) SetCriterias(tg_id int, criterias []bool) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(criterias); err != nil {
		return err
	}

	_, err := m.db.Exec("UPDATE users SET data = ? WHERE tg_id = ?", buf.Bytes(), tg_id)
	return err
}

func (m *SQLMap) UpdateDataByUserID(tg_id int, newCriterias []bool) error {
	// Сначала кодируем новые данные в формат gob
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(newCriterias); err != nil {
		return err
	}

	// Затем обновляем данные в базе данных
	_, err := m.db.Exec("UPDATE users SET data = ? WHERE tg_id = ?", buf.Bytes(), tg_id)
	return err
}

func (m *SQLMap) GetCriteriasByUserID(tg_id int) ([]bool, error) {
	var data []byte
	query := "SELECT data FROM users WHERE tg_id = ?"
	err := m.db.QueryRow(query, tg_id).Scan(&data)

	if err == sql.ErrNoRows {
		// Пользователь не найден
		return nil, nil // Или верните ошибку, если это необходимо
	} else if err != nil {
		// Обработка других ошибок
		return nil, err
	}

	var criterias []bool
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&criterias); err != nil {
		return nil, err
	}

	return criterias, nil
}

// Метод для получения всех пользователей и их данных
func (m *SQLMap) PrintAllUsers() error {
	rows, err := m.db.Query("SELECT tg_id, data FROM users")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tgID int
		var data []byte
		if err := rows.Scan(&tgID, &data); err != nil {
			return err
		}

		// Декодируем данные из формата GOB
		var criterias []bool
		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		if err := dec.Decode(&criterias); err != nil {
			return err
		}

		// Выводим tg_id и data
		fmt.Printf("tg_id: %d, data: %v\n", tgID, criterias)
	}

	// Проверяем, произошла ли ошибка при чтении строк
	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}
