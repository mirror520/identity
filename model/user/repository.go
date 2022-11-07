package user

type Repository interface {
	// User operations
	Store(u *User) error
	FindBySocialID(id SocialAccountID) (*User, error)

	// Workspace operations
	StoreWorkspace(w *Workspace) error
	FindWorkspaces(id UserID) ([]*WorkspaceMember, error)
}
