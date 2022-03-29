package main

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestSetModel struct {
	ID        int
	CreatedAt time.Time
}

func (m *TestSetModel) GetID() int {
	if m == nil {
		return 0
	}
	return m.ID
}

func (m *TestSetModel) GetCreatedAt() time.Time {
	if m == nil {
		return time.Unix(0, 0)
	}
	return m.CreatedAt
}

func (m *TestSetModel) String() string {
	return fmt.Sprintf("TestSetModel(ID: %d, %s)", m.ID, m.CreatedAt.Format(time.RFC3339))
}

var (
	generateTestSetModelID int32 = 0
)

func generateTestSetModel(createdAt time.Time) *TestSetModel {
	id := atomic.AddInt32(&generateTestSetModelID, 1)
	model := &TestSetModel{
		ID:        int(id),
		CreatedAt: createdAt,
	}

	return model
}

func generateTestSet(modelCount int) *Set[*TestSetModel] {
	set := &Set[*TestSetModel]{}

	createdAt := time.Now()
	for i := 0; i < modelCount; i++ {
		model := generateTestSetModel(createdAt)
		set.Add(model)
	}

	return set
}

func TestSet(t *testing.T) {
	model := generateTestSetModel(time.Now())

	set := &Set[*TestSetModel]{}

	ok := set.Add(model)
	assert.Truef(t, ok, "Could not add %s to %s", model, set)
	assert.Equal(t, 1, set.Len())

	empty, ok := set.Get(100)
	assert.Nil(t, empty)
	assert.False(t, ok)

	m, ok := set.Get(model.ID)
	assert.Equal(t, model, m)
	assert.True(t, ok)
}

func TestSetOrdered(t *testing.T) {
	set := generateTestSet(0)

	now := time.Now()
	ids := []int{}
	for i := 0; i < 5; i++ {
		now = now.Add(1 * time.Minute)
		model := generateTestSetModel(now)
		set.Add(model)
		ids = append([]int{model.GetID()}, ids...)
	}
	atDup1 := generateTestSetModel(now)
	atDup2 := generateTestSetModel(now)
	set.Add(atDup2)
	set.Add(atDup1)
	ids = append(ids[:2], ids[0:]...)
	ids[1] = atDup1.ID
	ids[2] = atDup2.ID

	assert.Equal(t, 7, set.Len())
	set.ForEach(func(i int, m *TestSetModel) {
		actual := ids[i]
		assert.Equal(t, actual, m.ID)
	})
}
