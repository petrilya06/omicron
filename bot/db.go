package bot

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func ConnectDB(tg_id int, default_criterias []byte) {
	db, err := sql.Open("sqlite3", "./users.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	makeTable := `CREATE TABLE IF NOT EXISTS users (tg_id integer primary key, data BLOB)`
	_, err = db.Exec(makeTable)
	if err != nil {
		log.Fatal(err)
	}

	addUser := `INSERT OR IGNORE INTO users (tg_id, data) VALUES (?, ?)`
	_, err = db.Exec(addUser, tg_id, default_criterias)
	if err != nil {
		log.Fatal(err)
	}
}

func GetUsers() {
	db, err := sql.Open("sqlite3", "./users.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT tg_id, data FROM users`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var tg_id int
		var data []byte
		err = rows.Scan(&tg_id, &data)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("tg_id: %d, data: %s\n", tg_id, data)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
}
