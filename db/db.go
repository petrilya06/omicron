package db

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"encoding/json"
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

func (m *SQLMap) GetDataByUserID(tg_id int) ([]bool, error) {
	var data []byte
	err := m.db.QueryRow("SELECT data FROM users WHERE tg_id = ?", tg_id).Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var criterias []bool
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&criterias); err != nil {
		return nil, err
	}

	return criterias, nil
}

// Метод для получения всех пользователей и их данных
func (m *SQLMap) GetAllUsers() ([]struct {
	TgID int
	Data []bool
}, error) {
	rows, err := m.db.Query("SELECT tg_id, data FROM users")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []struct {
		TgID int
		Data []bool
	}

	for rows.Next() {
		var user struct {
			TgID int
			Data []byte
		}
		if err := rows.Scan(&user.TgID, &user.Data); err != nil {
			return nil, err
		}

		var criterias []bool
		if err := json.Unmarshal(user.Data, &criterias); err != nil {
			return nil, err
		}

		users = append(users, struct {
			TgID int
			Data []bool
		}{
			TgID: user.TgID,
			Data: criterias,
		})
	}

	return users, nil
}
