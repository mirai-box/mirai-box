package model

// User represents an admin user.
type User struct {
	ID       string `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	Password string `db:"password" json:"-"`
	Role     string `db:"role" json:"role"`
}
