package main

import (
	"sync"
	"time"
)

// Set の対象となるモデルのインターフェース
type Model interface {
	GetID() int
	GetCreatedAt() time.Time
}

// モデルの集合を表す構造体
// Set.Get(id) で ID からモデルを取る
// Set.At(index) で先頭から index 番目のモデルを取る
// Set.Add(model) で集合にモデルを追加
type Set[T Model] struct {
	mu   sync.RWMutex
	list []T
	dict map[int]T
}

// 集合に含まれるモデルの個数を返す
func (s *Set[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.list)
}

// 先頭から index 番目のモデルを取るメソッド
func (s *Set[T]) At(index int) T {
	// 読み取りロック
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Set.list が未初期化ならゼロ値を返す
	// *new(T) は T のゼロ値を返す
	// Set[*User] なら nil を返す
	if s.list == nil {
		return *new(T)
	}

	return s.list[index]
}

// ID からモデルを取るメソッド
func (s *Set[T]) Get(id int) (T, bool) {
	// 読み取りロック
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Set.dict が未初期化ならゼロ値と false を返す
	if s.dict == nil {
		return *new(T), false
	}

	model, ok := s.dict[id]
	return model, ok
}

// Set にモデルを追加するメソッド
// 追加時に CreatedAt でソート済みの位置に追加する
// CreatedAt が重複したら ID で昇順
func (s *Set[T]) Add(model T) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := model.GetID()
	if id == 0 {
		return false
	}

	if len(s.list) == 0 {
		// Set.list が空ならそのまま追加
		s.list = []T{model}
	} else {
		pos := 0

		// 先頭から順に探査していく
		for i := 0; i < len(s.list)-1; i++ {
			m := s.list[i]
			pos = i

			// 追加しようとしているモデルより CreatedAt が古ければその前に追加する
			if m.GetCreatedAt().Before(model.GetCreatedAt()) {
				break
			}

			// CreatedAt が同一で
			if m.GetCreatedAt().Equal(model.GetCreatedAt()) {
				// 追加しようとしているモデルより ID が大きければその前に追加する
				if m.GetID() > model.GetID() {
					break
				}
			}
		}

		s.list = append(s.list[:pos+1], s.list[pos:]...)
		s.list[pos] = model
	}

	// Set.dict が未初期化なら初期化
	if s.dict == nil {
		s.dict = make(map[int]T, 0)
	}
	// ID で対応するマップに保存
	s.dict[id] = model

	return true
}

// User の Set
type UserSet struct {
	Set[*User]
}

// Post の Set
type PostSet struct {
	Set[*Post]
}

// Comment の Set
type CommentSet struct {
	Set[*Comment]
}

type SetForEachFunc[T Model] func(idx int, model T)

func (s *Set[T]) ForEach(f SetForEachFunc[T]) {
	s.mu.RLock()
	list := s.list[:]
	s.mu.RUnlock()

	for idx, model := range list {
		f(idx, model)
	}
}
