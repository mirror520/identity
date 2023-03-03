package user

type Repository interface {
	// Command
	Store(u *User) error

	// Query
	Find(id UserID) (*User, error)
	FindByUsername(username string) (*User, error)
	FindBySocialID(socialID string) (*User, error)
}
