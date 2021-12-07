package action

import (
	"context"

	"github.com/rosylilly/private-isu-benchmarker/v1/model"
)

func GetTopPage(ctx context.Context, user *model.User) error {
	req, err := user.Agent.GET("/")
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

func GetLoginPage(ctx context.Context, user *model.User) error {
	req, err := user.Agent.GET("/login")
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
