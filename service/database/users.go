package database

import (
	"database/sql"
	"fmt"
	"errors"
)

func (db *appdbimpl) GetUserIDByUsername(username string) (int, error) {
	var userID int
	err := db.c.QueryRow("SELECT id FROM users WHERE name = ?", username).Scan(&userID)
	if err != nil {
		return 0, err // Handle no user found or other errors
	}
	return userID, nil
}

func (db *appdbimpl) CreateUsername(username string) (int, error) {
	result, err := db.c.Exec("INSERT INTO users (name, profile_pic) VALUES (?, 'default')", username)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

// SetName updates the username for a given user ID
func (db *appdbimpl) SetMyUserName(userID int, name string) error {
	err := db.c.QueryRow("SELECT name FROM users WHERE name = ?", name).Scan(&name)

	if errors.Is(err, nil) {
		// Se err è nil, significa che il nome esiste già
		return fmt.Errorf("username '%s' is already taken", name)
	}

	if err != sql.ErrNoRows {
		// Se c'è un errore diverso da sql.ErrNoRows, ritorniamolo
		return err
	}
	if errors.Is(err, sql.ErrNoRows) {
		_, err := db.c.Exec("UPDATE users SET name = ? WHERE id = ?", name, userID)
		return err
	}
	return nil
}

func (db *appdbimpl) SetMyPhoto(userID int, profile_pic string) error {
	_, err := db.c.Exec("UPDATE users SET profile_pic = ? WHERE id = ?", profile_pic, userID)
	return err
}

func (db *appdbimpl) AddToGroup(userID int, groupID int) error {
	_, err := db.c.Exec("INSERT INTO participant_relation (user_id, conversation_id) VALUES (?, ?)", userID, groupID)
	return err
}

func (db *appdbimpl) LeaveGroup(userID int, groupID int) error {
	// Attempt to remove the user from the group
	_, err := db.c.Exec("DELETE FROM participant_relation WHERE user_id = ? AND conversation_id = ?", userID, groupID)
	return err
}

func (db *appdbimpl) CheckUserParticipation(userID int, conversationID int) (bool, error) {
	var count int
	err := db.c.QueryRow("SELECT COUNT(*) FROM participant_relation WHERE user_id = ? AND conversation_id = ?", userID, conversationID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (db *appdbimpl) GetUsername(userID int) (string, error) {
	var username string
	err := db.c.QueryRow("SELECT name FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
}

func (db *appdbimpl) GetAllUsers() ([]string, error) {
	var usernames []string

	// Esegui la query per ottenere tutti gli utenti
	rows, err := db.c.Query("SELECT name FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Assicura che le righe vengano chiuse alla fine

	// Itera sui risultati e aggiungi i nomi utente alla slice
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		usernames = append(usernames, username)
	}

	// Controlla eventuali errori dopo l'iterazione
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return usernames, nil
}
