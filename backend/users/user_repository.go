package users

import (
	"fmt"
	"strings"
	"sync"
)

type UserRepository struct {
	users map[string]*User
	mu    *sync.RWMutex
}

func NewUserRepository() UserRepository {
	return UserRepository{
		users: make(map[string]*User),
		mu:    &sync.RWMutex{},
	}
}

func (db *UserRepository) GetUser(name string) (*User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	user, ok := db.users[name]
	if !ok {
		return &User{}, fmt.Errorf("error getting user '%s': does not exist", name)
	}

	return user, nil
}

func (db *UserRepository) PutUser(username string) {
	displayName := strings.Split(username, "@")[0]
	user := NewUser(username, displayName)
	db.mu.Lock()
	defer db.mu.Unlock()
	db.users[user.name] = user
}
