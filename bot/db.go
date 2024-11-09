package bot

import (
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"
)

type SQLMap struct {
	db *sql.DB
}

func NewSQLMap(db *sql.DB) (*SQLMap, error) {
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
	data, err := json.Marshal(criterias) // Преобразуем данные в JSON
	if err != nil {
		return err
	}
	_, err = m.db.Exec("UPDATE users SET data = ? WHERE tg_id = ?", data, tg_id)
	return err
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
