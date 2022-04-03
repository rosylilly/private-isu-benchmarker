package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
)

// failure.NewError で用いるエラーコード定義
const (
	ErrInvalidStatusCode failure.StringCode = "status-code"
	ErrInvalidPath       failure.StringCode = "path"
	ErrNotFound          failure.StringCode = "not-found"
	ErrCSRFToken         failure.StringCode = "csrf-token"
	ErrInvalidPostOrder  failure.StringCode = "post-order"
	ErrInvalidAsset      failure.StringCode = "asset"
)

// 複数のエラーを持つ構造体
// 1度の検証で複数のエラーが含まれる場合があるため
type ValidationError struct {
	Errors []error
}

// error インターフェースを満たす Error メソッド
func (v ValidationError) Error() string {
	messages := []string{}

	for _, err := range v.Errors {
		if err != nil {
			messages = append(messages, fmt.Sprintf("%v", err))
		}
	}

	return strings.Join(messages, "\n")
}

// ValidationError が空かを判定する
func (v ValidationError) IsEmpty() bool {
	for _, err := range v.Errors {
		if err != nil {
			if ve, ok := err.(ValidationError); ok {
				if !ve.IsEmpty() {
					return false
				}
			} else {
				return false
			}
		}
	}
	return true
}

// isucandar.BenchmarkStep に自身の持つエラーをすべて追加する
func (v ValidationError) Add(step *isucandar.BenchmarkStep) {
	for _, err := range v.Errors {
		if err != nil {
			// 中身が ValidationError なら展開する
			if ve, ok := err.(ValidationError); ok {
				ve.Add(step)
			} else {
				step.AddError(err)
			}
		}
	}
}

// レスポンスを検証するバリデータ関数の型
type ResponseValidator func(*http.Response) error

// レスポンスを検証する関数
// 複数のバリデータ関数を受け取ってすべてでレスポンスを検証し、 ValidationError を返す
func ValidateResponse(res *http.Response, validators ...ResponseValidator) ValidationError {
	errs := []error{}

	for _, validator := range validators {
		if err := validator(res); err != nil {
			errs = append(errs, err)
		}
	}

	return ValidationError{
		Errors: errs,
	}
}

// ステータスコードコードを検証するバリデータ関数を返す高階関数
// 例: ValidateResponse(res, WithStatusCode(200))
func WithStatusCode(statusCode int) ResponseValidator {
	return func(r *http.Response) error {
		if r.StatusCode != statusCode {
			// ステータスコードが一致しなければ HTTP メソッド、URL パス、期待したステータスコード、実際のステータスコードを持つ
			// エラーを返す
			return failure.NewError(
				ErrInvalidStatusCode,
				fmt.Errorf(
					"%s %s : expected(%d) != actual(%d)",
					r.Request.Method,
					r.Request.URL.Path,
					statusCode,
					r.StatusCode,
				),
			)
		}
		return nil
	}
}

// レスポンスヘッダを検証するバリデータ関数を返す高階関数
func WithLocation(val string) ResponseValidator {
	return func(r *http.Response) error {
		target := r.Request.URL.ResolveReference(&url.URL{Path: val})
		if r.Header.Get("Location") != target.String() {
			// ヘッダーが一致しなければ HTTP メソッド、URL パス、期待したパス、実際の Location ヘッダを持つ
			// エラーを返す
			return failure.NewError(
				ErrInvalidPath,
				fmt.Errorf(
					"%s %s : %s, expected(%s) != actual(%s)",
					r.Request.Method,
					r.Request.URL.Path,
					"Location",
					val,
					r.Header.Get("Location"),
				),
			)
		}
		return nil
	}
}

// レスポンスボディに特定の文字列が含まれていることを検証するバリデータ関数を返す高階関数
func WithIncludeBody(val string) ResponseValidator {
	return func(r *http.Response) error {
		defer r.Body.Close()

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return failure.NewError(
				ErrInvalidResponse,
				fmt.Errorf(
					"%s %s : %s",
					r.Request.Method,
					r.Request.URL.Path,
					err.Error(),
				),
			)
		}

		if bytes.IndexAny(body, val) == -1 {
			return failure.NewError(
				ErrNotFound,
				fmt.Errorf(
					"%s %s : %s is not found in body",
					r.Request.Method,
					r.Request.URL.Path,
					val,
				),
			)
		}

		return nil
	}
}

func WithCSRFToken(user *User) ResponseValidator {
	return func(r *http.Response) error {
		defer r.Body.Close()

		user.SetCSRFToken("")

		doc, err := goquery.NewDocumentFromReader(r.Body)
		if err != nil {
			return failure.NewError(
				ErrInvalidResponse,
				fmt.Errorf(
					"%s %s : %s",
					r.Request.Method,
					r.Request.URL.Path,
					err.Error(),
				),
			)
		}

		node := doc.Find(`input[name="csrf_token"]`).Get(0)
		if node == nil {
			return failure.NewError(
				ErrCSRFToken,
				fmt.Errorf(
					"%s %s : CSRF token is not found",
					r.Request.Method,
					r.Request.URL.Path,
				),
			)
		}

		for _, attr := range node.Attr {
			if attr.Key == "value" {
				user.SetCSRFToken(attr.Val)
			}
		}

		if user.GetCSRFToken() == "" {
			return failure.NewError(
				ErrCSRFToken,
				fmt.Errorf(
					"%s %s : CSRF token is not found",
					r.Request.Method,
					r.Request.URL.Path,
				),
			)
		}

		return nil
	}
}

func WithOrderedPosts() ResponseValidator {
	return func(r *http.Response) error {
		defer r.Body.Close()
		doc, err := goquery.NewDocumentFromReader(r.Body)
		if err != nil {
			return failure.NewError(
				ErrInvalidResponse,
				fmt.Errorf(
					"%s %s : %s",
					r.Request.Method,
					r.Request.URL.Path,
					err.Error(),
				),
			)
		}

		errs := []error{}
		previousCreatedAt := time.Now()
		doc.Find(".isu-posts .isu-post").Each(func(_ int, s *goquery.Selection) {
			post := s.First()
			idAttr, exists := post.Attr("id")
			if !exists {
				return
			}
			creadtedAtAttr, exists := post.Attr("data-created-at")
			if !exists {
				return
			}

			id, _ := strconv.Atoi(strings.TrimPrefix(idAttr, "pid_"))
			createdAt, _ := time.Parse(time.RFC3339, creadtedAtAttr)

			if createdAt.After(previousCreatedAt) {
				errs = append(errs,
					failure.NewError(
						ErrInvalidPostOrder,
						fmt.Errorf(
							"%s %s : invalid order in top page: %s",
							r.Request.Method,
							r.Request.URL.Path,
							createdAt,
						),
					),
				)
				AdminLogger.Printf("isu-post: %d: %s", id, createdAt)
			}

		})

		return ValidationError{errs}
	}
}

// アセットの MD5 ハッシュ
var (
	assetsMD5 = map[string]string{
		"favicon.ico":       "ad4b0f606e0f8465bc4c4c170b37e1a3",
		"js/timeago.min.js": "f2d4c53400d0a46de704f5a97d6d04fb",
		"js/main.js":        "9c309fed7e360c57a705978dab2c68ad",
		"css/style.css":     "e4c3606a18d11863189405eb5c6ca551",
	}
)

// 静的ファイルを検証するバリデータ関数を返す高階関数
func WithAssets(ctx context.Context, ag *agent.Agent) ResponseValidator {
	return func(r *http.Response) error {
		resources, err := ag.ProcessHTML(ctx, r, r.Body)
		if err != nil {
			return failure.NewError(
				ErrInvalidAsset,
				fmt.Errorf(
					"%s %s : %v",
					r.Request.Method,
					r.Request.URL.Path,
					err,
				),
			)
		}

		errs := []error{}

		for uri, res := range resources {
			path := strings.TrimPrefix(uri, ag.BaseURL.String())
			// リソースの取得時にエラー
			if res.Error != nil {
				errs = append(errs,
					failure.NewError(
						ErrInvalidAsset,
						fmt.Errorf(
							"%s /%s : %v",
							"GET",
							path,
							res.Error,
						),
					),
				)
				continue
			}

			// http.Response.Body を閉じる
			defer res.Response.Body.Close()

			// ステータスコードが 304 ならこれ以上の検証をしない
			if res.Response.StatusCode == 304 {
				continue
			}

			expectedMD5, ok := assetsMD5[path]
			if !ok {
				// 定義にないリソースなら検証しない
				continue
			}

			hash := md5.New()
			io.Copy(hash, res.Response.Body)
			actualMD5 := hex.EncodeToString(hash.Sum(nil))

			if expectedMD5 != actualMD5 {
				errs = append(errs,
					failure.NewError(
						ErrInvalidAsset,
						fmt.Errorf(
							"%s /%s : expected(MD5 %s) != actual(MD5 %s)",
							"GET",
							path,
							expectedMD5,
							actualMD5,
						),
					),
				)
			}
		}

		return ValidationError{
			Errors: errs,
		}
	}
}
