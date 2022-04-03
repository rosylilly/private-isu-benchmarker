package main

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
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

// GET / を送信
func GetRootAction(ctx context.Context, ag *agent.Agent) (*http.Response, error) {
	// リクエストを生成
	req, err := ag.GET("/")
	if err != nil {
		return nil, err
	}

	// リクエストを実行
	return ag.Do(ctx, req)
}

// POST / を送信
func PostRootAction(ctx context.Context, ag *agent.Agent, post *Post, csrfToken string) (*http.Response, error) {
	img, err := randomImage()
	if err != nil {
		return nil, err
	}

	body := bytes.NewBuffer([]byte{})
	form := multipart.NewWriter(body)

	form.WriteField("body", post.Body)
	form.WriteField("csrf_token", csrfToken)

	fileHeader := make(textproto.MIMEHeader)
	fileHeader.Set(
		"Content-Disposition",
		fmt.Sprintf(
			`form-data; name="%s"; filename="%s"`,
			"file", "image.png",
		),
	)
	fileHeader.Set("Content-Type", "image/png")
	file, err := form.CreatePart(fileHeader)
	if err != nil {
		return nil, err
	}
	if _, err := file.Write(img); err != nil {
		return nil, err
	}

	form.Close()

	// リクエストを生成
	req, err := ag.POST("/", body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", form.FormDataContentType())

	// リクエストを実行
	return ag.Do(ctx, req)
}
