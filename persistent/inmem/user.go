package inmem

import (
	"sync"

	"github.com/mirror520/identity/user"
)

type userRepository struct {
	users     map[user.UserID]*user.User   // map[UserID]*user.User
	usernames map[string]*user.User        // map[Username]*user.User
	socials   map[user.SocialID]*user.User // map[SocialID]*user.User
	sync.RWMutex
}

func NewUserRepository() (user.Repository, error) {
	repo := new(userRepository)
	repo.users = make(map[user.UserID]*user.User)
	repo.usernames = make(map[string]*user.User)
	repo.socials = make(map[user.SocialID]*user.User)
	return repo, nil
}

func (repo *userRepository) Store(u *user.User) error {
	repo.Lock()

	repo.users[u.ID] = u
	repo.usernames[u.Username] = u

	for _, account := range u.Accounts {
		repo.socials[account.SocialID] = u
	}

	repo.Unlock()
	return nil
}

func (repo *userRepository) Find(id user.UserID) (*user.User, error) {
	repo.RLock()
	defer repo.Unlock()

	u, ok := repo.users[id]
	if !ok {
		return nil, user.ErrUserNotFound
	}

	return u, nil
}

func (repo *userRepository) FindByUsername(username string) (*user.User, error) {
	repo.RLock()
	defer repo.Unlock()

	u, ok := repo.usernames[username]
	if !ok {
		return nil, user.ErrUserNotFound
	}

	return u, nil
}

func (repo *userRepository) FindBySocialID(socialID user.SocialID) (*user.User, error) {
	repo.RLock()
	defer repo.RUnlock()

	u, ok := repo.socials[socialID]
	if !ok {
		return nil, user.ErrUserNotFound
	}

	return u, nil
}
