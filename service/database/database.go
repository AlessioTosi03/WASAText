/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Example configuration struct, normally you'd load this from a config file or environment variables
	cfg := struct {
		DB struct {
			Filename string
		}
	}{}
	// Use the path to your goland identifier.sqlite file here
	cfg.DB.Filename = "./identifier.sqlite" // Replace with the actual path to your SQLite file

	// Open database connection
	db, err := sql.Open("sqlite3", cfg.DB.Filename)
	if err != nil {
		log.Fatalf("error opening SQLite DB: %v", err)
	}
	defer db.Close()

	// Check if 'users' table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users';").Scan(&tableName)
	if err != nil {
		log.Println("Table 'users' not found, creating it...")
		_, err := db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, username TEXT, profile_pic TEXT)`)
		if err != nil {
			log.Fatalf("error creating 'users' table: %v", err)
		}
	}

	// Now you can use appDB for your application logic
	fmt.Println("AppDatabase initialized and ready to use")
}

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	GetName() (string, error)
	SetMyUserName(userID int, name string) error
	SetMyPhoto(userID int, profile_pic string) error
	//AddToGroup
	//LeaveGroup
	//SetGroupName
	//SetGroupPhoto
	//GetConversation
	//SendMessage
	//ForwardMessage
	//DeleteMessage
	//CommentMessage
	//UncommentMesage
	//GetMyConversations
	//DoLogin
	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

func (db *appdbimpl) GetUserIDByUsername(username string) (int, error) {
	var userID int
	err := db.c.QueryRow("SELECT id FROM users WHERE name = ?", username).Scan(&userID)
	if err != nil {
		return 0, err // Handle no user found or other errors
	}
	return userID, nil
}

func (db *appdbimpl) SetMyPhoto(userID int, profile_pic string) error {
	_, err := db.c.Exec("UPDATE users SET profile_pic = ? WHERE id = ?", profile_pic, userID)
	return err
}

func (db *appdbimpl) AddToGroup(userID int, groupID int) error {
	_, err := db.c.Exec("INSERT INTO groups (user_id, group_id) VALUES (?, ?)", userID, groupID)
	return err
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Check if table exists. If not, the database is empty, and we need to create the structure
	var tableName string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='example_table';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `CREATE TABLE example_table (id INTEGER NOT NULL PRIMARY KEY, name TEXT);`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w", err)
		}
	}
	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
