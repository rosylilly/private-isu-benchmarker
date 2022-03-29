package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
)

var (
	// 選手向け情報を出力するロガー
	ContestantLogger = log.New(os.Stdout, "", log.Ltime|log.Lmicroseconds)
	// 大会運営向け情報を出力するロガー
	AdminLogger = log.New(os.Stderr, "[ADMIN] ", log.Ltime|log.Lmicroseconds)
)

// 各種オプションのデフォルト値
const (
	DefaultTargetHost               = "localhost:8080"
	DefaultRequestTimeout           = 3 * time.Second
	DefaultInitializeRequestTimeout = 10 * time.Second
	DefaultExitErrorOnFail          = true
)

func init() {
	failure.BacktraceCleaner.Add(failure.SkipGOROOT)
}

func main() {
	// ベンチマークオプションの生成
	option := Option{}

	// 各フラグとベンチマークオプションのフィールドを紐付ける
	flag.StringVar(&option.TargetHost, "target-host", DefaultTargetHost, "Benchmark target host with port")
	flag.DurationVar(&option.RequestTimeout, "request-timeout", DefaultRequestTimeout, "Default request timeout")
	flag.DurationVar(&option.InitializeRequestTimeout, "initialize-request-timeout", DefaultInitializeRequestTimeout, "Initialize request timeout")
	flag.BoolVar(&option.ExitErrorOnFail, "exit-error-on-fail", DefaultExitErrorOnFail, "Exit with error if benchmark fails")

	// コマンドライン引数のパースを実行
	// この時点で各フィールドに値が設定されます
	flag.Parse()

	// 現在の設定を大会運営向けロガーに出力する
	AdminLogger.Print(option)

	// シナリオの生成
	scenario := &Scenario{
		Option: option,
	}

	// ベンチマークの生成
	benchmark, err := isucandar.NewBenchmark(
		// isucandar.Benchmark はステップ内の panic を自動で recover する機能があるが、今回は利用しない
		isucandar.WithoutPanicRecover(),
		// 負荷試験の時間は1分間
		isucandar.WithLoadTimeout(1*time.Minute),
	)
	if err != nil {
		AdminLogger.Fatal(err)
	}

	// ベンチマークにシナリオを追加
	benchmark.AddScenario(scenario)

	// main で最上位の context.Context を生成
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ベンチマーク開始
	result := benchmark.Start(ctx)

	// エラーをすべて表示する
	for _, err := range result.Errors.All() {
		// 選手向けにエラーメッセージが表示される
		ContestantLogger.Printf("%v", err)
		// 大会運営向けにスタックトレース付きエラーメッセージが表示される
		AdminLogger.Printf("%+v", err)
	}

	// スコアをすべて表示する
	for tag, count := range result.Score.Breakdown() {
		AdminLogger.Printf("%s: %d", tag, count)
	}
}
