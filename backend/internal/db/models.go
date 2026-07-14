package db

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Role         Role   `json:"role"`
	HomePath     string `json:"homePath,omitempty"`
}

func (u *User) ReadOnly() bool {
	return u.Role == RoleViewer
}

type Session struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"userId"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
}

type ApiToken struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"userId"`
	Token     string `json:"token"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}
