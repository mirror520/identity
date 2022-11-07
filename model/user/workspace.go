package user

import "github.com/mirror520/jinte/model"

type WorkspaceID uint

type Workspace struct {
	ID      WorkspaceID        `json:"id" gorm:"primarykey"`
	Name    string             `json:"name"`
	Members []*WorkspaceMember `json:"members"`
	model.Time
}

func NewWorkspace(name string) *Workspace {
	return &Workspace{Name: name}
}

type WorkspaceMemberRole string

const (
	WorkspaceOwner  WorkspaceMemberRole = "owner"
	WorkspaceAdmin  WorkspaceMemberRole = "admin"
	WorkspaceNormal WorkspaceMemberRole = "normal"
)

type WorkspaceMemberID uint

type WorkspaceMember struct {
	ID        WorkspaceMemberID   `json:"id" gorm:"primarykey"`
	Role      WorkspaceMemberRole `json:"role"`
	Workspace *Workspace          `json:"workspace"`
	User      *User               `json:"user"`

	WorkspaceID WorkspaceID `json:"-"`
	UserID      UserID      `json:"-"`
	model.Time
}

func (w *Workspace) AddMember(u *User, role WorkspaceMemberRole) *WorkspaceMember {
	if w.Members == nil {
		w.Members = make([]*WorkspaceMember, 0)
	}

	member := &WorkspaceMember{User: u, Role: WorkspaceOwner}
	w.Members = append(w.Members, member)

	return member
}
