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

	_ "github.com/mattn/go-sqlite3"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	// utilities
	GetName() (string, error)
	GetConversationsByUser(userID int) ([]Conversation, error)
	GetMessageUser(messageID int) (int, error)
	CheckUserParticipation(userID int, conversationID int) (bool, error)
	GetUserIDByUsername(username string) (int, error)
	GetGroupFromConversation(conversationID int) (Group, error)
	GetChatFromConversation(conversationID int) (Chat, error)
	GetOtherParticipant(conversationID int, userID int) (User, error)
	GetMessages(conversationID int) ([]Message, error)
	CreateGroup(userID int, groupName string, photoURL string) (int, error)
	CreateChat(userID int, otherUserID int) (int, error)
	GetUsername(userID int) (string, error)
	GetMyReaction(userID int, messageID int) (string, error)
	// mains
	GetUserByReaction(reactionID int) (int, error)
	SetMyUserName(userID int, name string) error
	SetMyPhoto(userID int, profile_pic string) error
	AddToGroup(userID int, groupID int) error
	LeaveGroup(userID int, groupID int) error
	SetGroupName(groupID int, groupName string) error
	SetGroupPhoto(groupID int, photoURL string) error
	GetConversation(conversationID int) (Conversation, error)
	SendMessage(userID int, conversationID int, message string, picture string) error
	ForwardMessage(userID int, conversationID int, forwardedID int) error
	DeleteMessage(messageID int) error
	CommentMessage(userID int, messageID int, emoji string) error
	UncommentMessage(userID int, messageID int) (int64, error)
	CreateUsername(username string) (int, error)
	DoLogin(userID int) (User, error)
	// others
	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Check if table exists. If not, the database is empty, and we need to create the structure
	var tableName string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='users';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `CREATE TABLE IF NOT EXISTS users (
					id INTEGER PRIMARY KEY,
					name TEXT NOT NULL,
					profile_pic BLOB
				);`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w", err)
		}
	}
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='conversations';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `CREATE TABLE IF NOT EXISTS conversations (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					type TEXT NOT NULL CHECK (type IN ('chat', 'group')), -- Enforces 'chat' or 'group'
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
					);`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w", err)
		}
	}
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='chats';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `CREATE TABLE IF NOT EXISTS chats (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					conversation_id INTEGER NOT NULL,
					FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
					UNIQUE (conversation_id) -- Garantisce che ogni chat sia unica
					);`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w", err)
		}
	}
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='groups';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `CREATE TABLE IF NOT EXISTS groups (
						id INTEGER PRIMARY KEY,
						conversation_id INTEGER NOT NULL UNIQUE,
						group_name TEXT NOT NULL,
						group_pic BLOB,
						FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
					);`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w", err)
		}
	}
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='participation_relation';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `CREATE TABLE IF NOT EXISTS participant_relation (
						conversation_id INTEGER NOT NULL,
						user_id INTEGER NOT NULL,
						PRIMARY KEY (conversation_id, user_id),
						FOREIGN KEY (user_id) REFERENCES users(id),
						FOREIGN KEY (conversation_id) REFERENCES conversations(id)
					);`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w", err)
		}
	}
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='messages';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `
				CREATE TABLE IF NOT EXISTS messages (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					conversation_id INTEGER NOT NULL,
					user_id INTEGER NOT NULL,
					forwarded BOOLEAN,
					image BLOB,
					message_text TEXT NOT NULL,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
				);`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w", err)
		}
	}
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='message_read_status';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `CREATE TABLE IF NOT EXISTS message_read_status (
						message_id INTEGER NOT NULL UNIQUE,
						user_id INTEGER NOT NULL UNIQUE,
						read_timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
						PRIMARY KEY (message_id, user_id),
						FOREIGN KEY (user_id) REFERENCES users(id),
						FOREIGN KEY (message_id) REFERENCES messages(id)
					);`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w", err)
		}
	}
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='message_received_status';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `CREATE TABLE IF NOT EXISTS message_received_status (
					message_id INTEGER NOT NULL UNIQUE,
					user_id INTEGER NOT NULL UNIQUE,
					read_timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
					PRIMARY KEY (message_id, user_id),
					FOREIGN KEY (user_id) REFERENCES users(id),
					FOREIGN KEY (message_id) REFERENCES messages(id)
				);`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w", err)
		}
	}
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='reaction_relation';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		sqlStmt := `CREATE TABLE IF NOT EXISTS reaction_relation (
					message_id INTEGER NOT NULL,
					user_id INTEGER NOT NULL,
					emoji TEXT,
					PRIMARY KEY (message_id, user_id),
					FOREIGN KEY (user_id) REFERENCES users(id),
					FOREIGN KEY (message_id) REFERENCES messages(id)
					);`
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
