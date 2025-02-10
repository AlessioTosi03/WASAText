package database

func (db *appdbimpl) DoLogin(userID int) (User, error) {
	var user User
	err := db.c.QueryRow("SELECT name, profile_pic FROM users WHERE id = ?", userID).Scan(&user.Username, &user.ProfilePic)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
