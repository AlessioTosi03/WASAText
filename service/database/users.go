package database

// SetName updates the username for a given user ID
func (db *appdbimpl) SetMyUserName(userID int, name string) error {
	_, err := db.c.Exec("UPDATE users SET name = ? WHERE id = ?", name, userID)
	return err
}