package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/isucon/isucandar/agent"
)

// ベンチマークオプションを保持する構造体
type Option struct {
	TargetHost               string
	RequestTimeout           time.Duration
	InitializeRequestTimeout time.Duration
	ExitErrorOnFail          bool
}

// fmt.Stringer インターフェースを実装
// log.Print などに渡した際、このメソッドが実装されていれば返した文字列が出力される
func (o Option) String() string {
	args := []string{
		"benchmarker",
		fmt.Sprintf("--target-host=%s", o.TargetHost),
		fmt.Sprintf("--request-timeout=%s", o.RequestTimeout.String()),
		fmt.Sprintf("--initialize-request-timeout=%s", o.InitializeRequestTimeout.String()),
		fmt.Sprintf("--exit-error-on-fail=%v", o.ExitErrorOnFail),
	}

	return strings.Join(args, " ")
}

// Option の内容に沿った agent.Agent を生成
func (o Option) NewAgent(forInitialize bool) (*agent.Agent, error) {
	agentOptions := []agent.AgentOption{
		// リクエストのベース URL は Option.TargetHost かつ HTTP
		agent.WithBaseURL(fmt.Sprintf("http://%s/", o.TargetHost)),
		// agent.DefaultTransport を都度クローンして利用
		agent.WithCloneTransport(agent.DefaultTransport),
	}

	// initialize 用の agent.Agent かによってタイムアウト時間が違うのでオプションを調整
	if forInitialize {
		agentOptions = append(agentOptions, agent.WithTimeout(o.InitializeRequestTimeout))
	} else {
		agentOptions = append(agentOptions, agent.WithTimeout(o.RequestTimeout))
	}

	// オプションに従って agent.Agent を生成
	return agent.NewAgent(agentOptions...)
}
