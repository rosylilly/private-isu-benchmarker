package main

import (
	"context"
	"math/rand"
	"sync"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/score"
	"github.com/isucon/isucandar/worker"
)

// シナリオレベルで発生するエラーコードの定義
const (
	ErrFailedLoadJSON  failure.StringCode = "load-json"
	ErrCannotNewAgent  failure.StringCode = "agent"
	ErrInvalidRequest  failure.StringCode = "request"
	ErrInvalidResponse failure.StringCode = "response"
)

// シナリオで発生するスコアのタグ
const (
	ScoreGETLogin  score.ScoreTag = "GET /login"
	ScorePOSTLogin score.ScoreTag = "POST /login"
	ScoreGETRoot   score.ScoreTag = "GET /"
	ScorePOSTRoot  score.ScoreTag = "POST /"
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
	// レスポンスの Body は必ず Close
	defer res.Body.Close()

	// レスポンスを検証
	ValidateResponse(
		res,
		// ステータスコードが 200 であることを検証
		WithStatusCode(200),
	).Add(step)

	return nil
}

// isucandar.PrepeareScenario を満たすメソッド
// isucandar.Benchmark の Load ステップで実行される
func (s *Scenario) Load(ctx context.Context, step *isucandar.BenchmarkStep) error {
	wg := &sync.WaitGroup{}

	// 10秒おきにベンチマーク実行中であることを大会運営向けロガーに出力
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

	// 成功ケースのシナリオ
	successCase, err := worker.NewWorker(func(ctx context.Context, _ int) {
		if user, ok := s.Users.Get(rand.Intn(s.Users.Len())); ok {
			// 削除済みのユーザーを引いたらもう一回
			if user.DeleteFlag != 0 {
				return
			}

			// ログインに成功したら画像を投稿
			if s.LoginSuccess(ctx, step, user) {
				s.PostImage(ctx, step, user)
			}
			user.ClearAgent()
		}
	},
		// 無限回繰り返す
		worker.WithInfinityLoop(),
		// 4並列で実行
		worker.WithMaxParallelism(4),
	)
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		successCase.Process(ctx)
	}()

	// 失敗ケースのシナリオ
	failureCase, err := worker.NewWorker(func(ctx context.Context, _ int) {
		if user, ok := s.Users.Get(rand.Intn(s.Users.Len())); ok {
			// 削除済みのユーザーを引いたらもう一回
			if user.DeleteFlag != 0 {
				return
			}

			// ログインに失敗するだけ
			s.LoginFailure(ctx, step, user)
		}
	},
		// 20回繰り返す
		worker.WithLoopCount(20),
		// 2並列で実行
		worker.WithMaxParallelism(2),
	)
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		failureCase.Process(ctx)
	}()

	// トップページの並び順検証シナリオ
	orderedCase, err := worker.NewWorker(func(ctx context.Context, _ int) {
		if user, ok := s.Users.Get(rand.Intn(s.Users.Len())); ok {
			// トップページの並び順を検証
			s.OrderedIndex(ctx, step, user)
		}
	},
		// 無限回繰り返す
		worker.WithInfinityLoop(),
		// 2並列で実行
		worker.WithMaxParallelism(2),
	)
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		orderedCase.Process(ctx)
	}()

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

// 失敗するログインを実行するシナリオ
func (s *Scenario) LoginFailure(ctx context.Context, step *isucandar.BenchmarkStep, user *User) bool {
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
	// 本来のパスワードに間違った文字列を後付して間違ったパスワードにする
	postRes, err := PostLoginAction(ctx, ag, user.AccountName, user.Password+".invalid")
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
		// リダイレクト先はログインページ
		WithLocation("/login"),
	)
	postValidation.Add(step)

	if postValidation.IsEmpty() {
		// 検証結果のエラーが空ならスコアを追加
		step.AddScore(ScorePOSTLogin)
	} else {
		return false
	}

	// ここで context が終了している可能性があるのでチェックして終了していたら中断
	select {
	case <-ctx.Done():
		return false
	default:
	}

	// リダイレクト先となるログインページの取得
	redirectRes, err := GetLoginAction(ctx, ag)
	if err != nil {
		step.AddError(failure.NewError(ErrInvalidRequest, err))
		return false
	}
	defer redirectRes.Body.Close()

	// レスポンスを検証
	redirectValidation := ValidateResponse(
		redirectRes,
		// ステータスコードは 200
		WithStatusCode(200),
		// 適切なエラーメッセージが含まれていること
		WithIncludeBody("アカウント名かパスワードが間違っています"),
	)
	redirectValidation.Add(step)

	if redirectValidation.IsEmpty() {
		// 検証結果のエラーが空ならスコアを追加
		step.AddScore(ScoreGETLogin)
	} else {
		return false
	}

	// ログインに失敗したときだけ true を返す
	return true
}

// 画像を投稿するシナリオ
func (s *Scenario) PostImage(ctx context.Context, step *isucandar.BenchmarkStep, user *User) bool {
	// User に紐づくユーザーエージェントを取得
	ag, err := user.GetAgent(s.Option)
	if err != nil {
		step.AddError(failure.NewError(ErrCannotNewAgent, err))
		return false
	}

	// トップページへのリクエストを実行
	getRes, err := GetRootAction(ctx, ag)
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
		// CSRFToken を取得
		WithCSRFToken(user),
	)
	getValidation.Add(step)

	if getValidation.IsEmpty() {
		// 検証結果のエラーが空ならスコアを追加
		step.AddScore(ScoreGETRoot)
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

	// 画像を投稿
	post := &Post{
		Mime:   "image/png",
		Body:   randomText(),
		UserID: user.ID,
	}
	postRes, err := PostRootAction(ctx, ag, post, user.GetCSRFToken())
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
	)
	postValidation.Add(step)

	if postValidation.IsEmpty() {
		// 検証結果のエラーが空ならスコアを追加
		step.AddScore(ScorePOSTRoot)
	} else {
		return false
	}

	// ここで context が終了している可能性があるのでチェックして終了していたら中断
	select {
	case <-ctx.Done():
		return false
	default:
	}

	// トップページへ
	redirectRes, err := GetRootAction(ctx, ag)
	if err != nil {
		step.AddError(failure.NewError(ErrInvalidRequest, err))
		return false
	}
	defer redirectRes.Body.Close()

	redirectValidation := ValidateResponse(
		redirectRes,
		// ステータスコードは 200
		WithStatusCode(200),
		// 投稿した画像も含めリソースを取得
		WithAssets(ctx, ag),
	)
	redirectValidation.Add(step)

	if redirectValidation.IsEmpty() {
		// 検証結果のエラーが空ならスコアを追加
		step.AddScore(ScoreGETRoot)
	} else {
		return false
	}

	// 画像の投稿に成功したら true を返す
	return true
}

// トップページの並び順を検証するシナリオ
func (s *Scenario) OrderedIndex(ctx context.Context, step *isucandar.BenchmarkStep, user *User) bool {
	// User に紐づくユーザーエージェントを取得
	ag, err := user.GetAgent(s.Option)
	if err != nil {
		step.AddError(failure.NewError(ErrCannotNewAgent, err))
		return false
	}

	// トップページへのリクエストを実行
	getRes, err := GetRootAction(ctx, ag)
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
		// Post の並び順を検証
		WithOrderedPosts(),
	)
	getValidation.Add(step)

	if getValidation.IsEmpty() {
		// 検証結果のエラーが空ならスコアを追加
		step.AddScore(ScoreGETRoot)
	} else {
		// エラーがあればここでシナリオは停止
		return false
	}

	// 不備がなければ true を返す
	return true
}
