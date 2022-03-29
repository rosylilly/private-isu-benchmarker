package main

import (
	"context"
	"sync"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/score"
)

// シナリオレベルで発生するエラーコードの定義
const (
	ErrFailedLoadJSON failure.StringCode = "load-json"
	ErrCannotNewAgent failure.StringCode = "agent"
	ErrInvalidRequest failure.StringCode = "request"
)

// シナリオで発生するスコアのタグ
const (
	ScoreGETLogin  score.ScoreTag = "GET /login"
	ScorePOSTLogin score.ScoreTag = "POST /login"
)

// オプションと全データを持つシナリオ構造体
type Scenario struct {
	Option   Option
	Users    UserSet
	Posts    PostSet
	Comments CommentSet
}

// isucandar.PrepeareScenario を満たすメソッド
// isucandar.Benchmark の Prepare ステップで実行される
func (s *Scenario) Prepare(ctx context.Context, step *isucandar.BenchmarkStep) error {
	// User のダンプデータをロード
	if err := s.Users.LoadJSON("./dump/users.json"); err != nil {
		return failure.NewError(ErrFailedLoadJSON, err)
	}

	// Post のダンプデータをロード
	if err := s.Posts.LoadJSON("./dump/posts.json"); err != nil {
		return failure.NewError(ErrFailedLoadJSON, err)
	}

	// Comment のダンプデータをロード
	if err := s.Comments.LoadJSON("./dump/comments.json"); err != nil {
		return failure.NewError(ErrFailedLoadJSON, err)
	}

	// GET /initialize 用ユーザーエージェントの生成
	ag, err := s.Option.NewAgent(true)
	if err != nil {
		return failure.NewError(ErrCannotNewAgent, err)
	}

	// GET /initialize へのリクエストを実行
	res, err := GetInitializeAction(ctx, ag)
	if err != nil {
		return failure.NewError(ErrInvalidRequest, err)
	}
	// レスポンスの Body は必ず Close する
	defer res.Body.Close()

	// レスポンスを検証する
	ValidateResponse(
		res,
		// ステータスコードが 200 であることを検証する
		WithStatusCode(200),
	).Add(step)

	return nil
}

// isucandar.PrepeareScenario を満たすメソッド
// isucandar.Benchmark の Load ステップで実行される
func (s *Scenario) Load(ctx context.Context, step *isucandar.BenchmarkStep) error {
	wg := &sync.WaitGroup{}

	// 10秒おきにベンチマーク実行中であることを大会運営向けロガーに出力する
	// wg.Add(1)
	// go func() {
	// 	for {
	// 		AdminLogger.Print("ベンチマーク実行中")

	// 		select {
	// 		case <-ctx.Done():
	// 			return
	// 		case <-time.After(10 * time.Second):
	// 		}
	// 	}
	// }()

	if user, ok := s.Users.Get(1); ok {
		s.LoginSuccess(ctx, step, user)
	}

	wg.Wait()

	return nil
}

// isucandar.PrepeareScenario を満たすメソッド
// isucandar.Benchmark の Validation ステップで実行される
func (s *Scenario) Validation(ctx context.Context, step *isucandar.BenchmarkStep) error {
	return nil
}

// 成功するログインを実行するシナリオ
func (s *Scenario) LoginSuccess(ctx context.Context, step *isucandar.BenchmarkStep, user *User) bool {
	// User に紐づくユーザーエージェントを取得
	ag, err := user.GetAgent(s.Option)
	if err != nil {
		step.AddError(failure.NewError(ErrCannotNewAgent, err))
		return false
	}

	// ログインページへのリクエストを実行
	getRes, err := GetLoginAction(ctx, ag)
	if err != nil {
		step.AddError(failure.NewError(ErrInvalidRequest, err))
		return false
	}
	defer getRes.Body.Close()

	// レスポンスを検証
	getValidation := ValidateResponse(
		getRes,
		// ステータスコードは 200
		WithStatusCode(200),
		// 静的リソースを検証
		WithAssets(ctx, ag),
	)
	getValidation.Add(step)

	if getValidation.IsEmpty() {
		// 検証結果のエラーが空ならスコアを追加
		step.AddScore(ScoreGETLogin)
	} else {
		// エラーがあればここでシナリオは停止
		return false
	}

	// ここで context が終了している可能性があるのでチェックして終了していたら中断
	select {
	case <-ctx.Done():
		return false
	default:
	}

	// ログインするリクエストを実行
	postRes, err := PostLoginAction(ctx, ag, user.AccountName, user.Password)
	if err != nil {
		step.AddError(failure.NewError(ErrInvalidRequest, err))
		return false
	}
	defer postRes.Body.Close()

	// レスポンスを検証
	postValidation := ValidateResponse(
		postRes,
		// ステータスコードは 302
		WithStatusCode(302),
		// リダイレクト先はトップページ
		WithLocation("/"),
	)
	postValidation.Add(step)

	if postValidation.IsEmpty() {
		// 検証結果のエラーが空ならスコアを追加
		step.AddScore(ScorePOSTLogin)
	} else {
		return false
	}

	// ログインに成功したときだけ true を返す
	return true
}
