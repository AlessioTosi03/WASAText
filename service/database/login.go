package database

func (db *appdbimpl) DoLogin(username string) (int, error) {
	var userID int
	err := db.c.QueryRow("SELECT id FROM users WHERE name = ?", userID).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}
