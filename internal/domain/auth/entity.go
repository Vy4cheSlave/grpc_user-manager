package auth

type User struct {
	Id           string  `json:"id"`
	Email        string  `json:"email"`
	Username     string  `json:"username"`
	PasswordHash []byte  `json:"password_hash"`
	FirstName    *string `json:"first_name"`
	LastName     *string `json:"last_name"`
	IsActive     bool    `json:"is_active"`
	Role         string  `json:"role"`
}
