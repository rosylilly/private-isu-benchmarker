package scenario

import (
	"context"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/rosylilly/private-isu-benchmarker/v0/action"
	"github.com/rosylilly/private-isu-benchmarker/v0/model"
)

type Scenario struct {
	config *Config
}

func NewScenario(config *Config) *Scenario {
	return &Scenario{
		config: config,
	}
}

func (s *Scenario) Prepare(ctx context.Context, step *isucandar.BenchmarkStep) error {
	s.config.DebugLogger.Printf("prepare")

	return nil
}

func (s *Scenario) Load(ctx context.Context, step *isucandar.BenchmarkStep) error {
	s.config.DebugLogger.Printf("load")

	for {
		client, err := agent.NewAgent(
			agent.WithCloneTransport(agent.DefaultTransport),
			agent.WithBaseURL(s.config.BaseURL.String()),
			agent.WithTimeout(s.config.RequestTimeout),
		)

		if err != nil {
			return err
		}

		user, err := model.NewUser(client)
		if err != nil {
			return err
		}

		if err := action.GetTopPage(ctx, user); err != nil {
			return err
		}

		step.AddScore(SCORE_GET_TOPPAGE)

		select {
		case <-ctx.Done():
			return nil
		default:
		}
	}
}

func (s *Scenario) Validation(ctx context.Context, step *isucandar.BenchmarkStep) error {
	s.config.DebugLogger.Printf("validation")
	return nil
}
