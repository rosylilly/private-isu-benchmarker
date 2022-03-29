package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserLoadJSON(t *testing.T) {
	set := &UserSet{}

	err := set.LoadJSON("./dump/users.json")
	assert.NoError(t, err)
	assert.Equal(t, 1000, set.Len())

	model := set.At(0)
	assert.Equal(t, 1000, model.GetID())
	assert.Equal(t, "celina", model.AccountName)
}

func TestPostLoadJSON(t *testing.T) {
	set := &PostSet{}

	err := set.LoadJSON("./dump/posts.json")
	assert.NoError(t, err)
	assert.Equal(t, 10000, set.Len())

	model := set.At(0)
	assert.Equal(t, 10000, model.GetID())
	assert.Equal(t, "ﾄﾞﾓﾄﾞﾓ＼(^_^ ) ( ^_^)/ﾄﾞﾓﾄﾞﾓ", model.Body)
}

func TestCommentLoadJSON(t *testing.T) {
	set := &CommentSet{}

	err := set.LoadJSON("./dump/comments.json")
	assert.NoError(t, err)
	assert.Equal(t, 100000, set.Len())

	model := set.At(0)
	assert.Equal(t, 100000, model.GetID())
	assert.Equal(t, "ｵﾒｯﾄﾅｹﾞｷｯｽ♪(▼ヽ▼*)ﾝｰ.....ヾ(*▼・▼)ﾉ⌒☆ﾁｭｯ♪", model.Comment)
}
