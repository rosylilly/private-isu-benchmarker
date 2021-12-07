package action

import (
	"context"
	"net/url"
	"strings"

	"github.com/rosylilly/private-isu-benchmarker/v1/model"
)

func PostLogin(ctx context.Context, user *model.User) error {
	form := url.Values{}
	form.Add("account_name", user.AccountName)
	form.Add("password", user.Password)

	req, err := user.Agent.POST("/login", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	res, err := user.Agent.Do(ctx, req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
