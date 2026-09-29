package server

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(path string) (*Database, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	database := &Database{
		db: db,
	}

	return database, nil
}

func (d *Database) Init() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY,
			sender TEXT NOT NULL,
			recipient TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)

	return err
}

func (d *Database) SaveMessage(msg Message) error {
	_, err := d.db.Exec(`
		INSERT INTO messages (id, sender, recipient, content)
		VALUES (?, ?, ?, ?)
	`, msg.ID, msg.Sender, msg.Recipient, msg.Content)

	return err
}