package main

import (
	"context"
	"flag"
	"os"

	"github.com/isucon/isucandar"
	"github.com/rosylilly/private-isu-benchmarker/v0/scenario"
)

var (
	config *scenario.Config
)

func init() {
	config = scenario.NewConfig()

	flag.DurationVar(&config.LoadTimeout, "load-timeout", config.LoadTimeout, "set load timeout duration")
	flag.DurationVar(&config.RequestTimeout, "request-timeout", config.RequestTimeout, "set request timeout duration")
	flag.BoolVar(&config.Debug, "debug", false, "Enable debug mode")

	flag.Parse()

	config.SetupLogger()
}

func main() {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	benchmark, err := isucandar.NewBenchmark(
		isucandar.WithLoadTimeout(config.LoadTimeout),
	)
	if err != nil {
		panic(err)
	}

	benchmark.OnError(func(err error, bs *isucandar.BenchmarkStep) {
		config.DebugLogger.Printf("%v", err)
	})

	benchmark.AddScenario(scenario.NewScenario(config))

	result := benchmark.Start(ctx)
	exitWithResult(result)
}

func exitWithResult(result *isucandar.BenchmarkResult) {
	result.Score.DefaultScoreMagnification = 1

	additions := result.Score.Sum()
	criticals := int64(0)
	timeouts := int64(0)
	deductions := int64(0)

	for _, err := range result.Errors.All() {
		critical, timeout, deduction := errorClassification(err)
		if critical {
			criticals++
		}

		if timeout {
			timeouts++
		}

		if deduction {
			deductions++
		}
	}

	total := additions - deductions
	if total < 0 {
		total = 0
	}

	success := (total >= 0 && criticals <= 0 && timeouts <= 0)

	for tag, score := range result.Score.Breakdown() {
		config.DebugLogger.Printf("score: %s : %d", tag, score)
	}

	config.InfoLogger.Printf("score: %d (%d - %d)", total, additions, deductions)
	config.InfoLogger.Printf("success: %v", success)
	if !success {
		os.Exit(1)
	}
}

func errorClassification(err error) (bool, bool, bool) {
	// TODO: Implement error classification
	critical := false
	timeout := false
	deduction := true

	return critical, timeout, deduction
}
