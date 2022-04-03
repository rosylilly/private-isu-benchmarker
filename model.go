package main

import (
	"sync"
	"time"

	"github.com/isucon/isucandar/agent"
)

// User の構造体
// 後ほど JSON 化したダンプデータから読み込めるようにタグを付与しています
type User struct {
	mu sync.RWMutex

	ID          int       `json:"id"`
	AccountName string    `json:"account_name"`
	Password    string    `json:"password"`
	Authority   int       `json:"authority"`
	DeleteFlag  int       `json:"del_flg"`
	CreatedAt   time.Time `json:"created_at"`

	csrfToken string
	Agent     *agent.Agent
}

// Model.GetID の実装
func (m *User) GetID() int {
	if m == nil {
		return 0
	}

	return m.ID
}

// Model.GetCreatedAt の実装
func (m *User) GetCreatedAt() time.Time {
	if m == nil {
		return time.Unix(0, 0)
	}

	return m.CreatedAt
}

// User に紐づく agent.Agent
func (m *User) GetAgent(o Option) (*agent.Agent, error) {
	m.mu.RLock()
	a := m.Agent
	m.mu.RUnlock()

	if a != nil {
		return a, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	a, err := o.NewAgent(false)
	if err != nil {
		return nil, err
	}
	m.Agent = a

	return a, nil
}

// ユーザーの agent.Agent を初期化
func (m *User) ClearAgent() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Agent = nil
}

// ユーザーごとの CSRF トークンのセット
func (m *User) SetCSRFToken(token string) {
	m.mu.Lock()
	m.csrfToken = token
	m.mu.Unlock()
}

// ユーザーごとの CSRF トークンの取得
func (m *User) GetCSRFToken() string {
	m.mu.RLock()
	token := m.csrfToken
	m.mu.RUnlock()

	return token
}

// Post の構造体
// 後ほど JSON 化したダンプデータから読み込めるようにタグを付与しています
type Post struct {
	mu sync.RWMutex

	ID          int       `json:"id"`
	Mime        string    `json:"mime"`
	Body        string    `json:"body"`
	ImgdataHash string    `json:"imgdata_hash"`
	UserID      int       `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// Model.GetID の実装
func (m *Post) GetID() int {
	if m == nil {
		return 0
	}

	return m.ID
}

// Model.GetCreatedAt の実装
func (m *Post) GetCreatedAt() time.Time {
	if m == nil {
		return time.Unix(0, 0)
	}

	return m.CreatedAt
}

// Comment の構造体
// 後ほど JSON 化したダンプデータから読み込めるようにタグを付与しています
type Comment struct {
	mu sync.RWMutex

	ID        int       `json:"id"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
}

// Model.GetID の実装
func (m *Comment) GetID() int {
	if m == nil {
		return 0
	}

	return m.ID
}

// Model.GetCreatedAt の実装
func (m *Comment) GetCreatedAt() time.Time {
	if m == nil {
		return time.Unix(0, 0)
	}

	return m.CreatedAt
}
