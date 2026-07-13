package db

type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	ReadOnly     bool   `json:"readOnly"`
}

type Session struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"userId"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
}
