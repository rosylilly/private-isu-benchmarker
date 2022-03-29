package main

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/isucon/isucandar/agent"
)

// GET /initialize を送信
// 第一引数に context.context を取ることで外からリクエストをキャンセルできるようにしている
func GetInitializeAction(ctx context.Context, ag *agent.Agent) (*http.Response, error) {
	// リクエストを生成
	req, err := ag.GET("/initialize")
	if err != nil {
		return nil, err
	}

	// リクエストを実行
	return ag.Do(ctx, req)
}

// GET /login を送信
func GetLoginAction(ctx context.Context, ag *agent.Agent) (*http.Response, error) {
	// リクエストを生成
	req, err := ag.GET("/login")
	if err != nil {
		return nil, err
	}

	// リクエストを実行
	return ag.Do(ctx, req)
}

// POST /login を送信
func PostLoginAction(ctx context.Context, ag *agent.Agent, accountName, password string) (*http.Response, error) {
	values := url.Values{}
	values.Add("account_name", accountName)
	values.Add("password", password)

	// リクエストを生成
	req, err := ag.POST("/login", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// リクエストを実行
	return ag.Do(ctx, req)
}
