package model

import (
	"github.com/isucon/isucandar/agent"
)

type User struct {
	AccountName string
	Password    string
	Agent       *agent.Agent
}

func NewUser(a *agent.Agent) (*User, error) {
	user := &User{
		Agent: a,
	}

	return user, nil
}
