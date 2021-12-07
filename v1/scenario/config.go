package scenario

import (
	"io"
	"log"
	"net/url"
	"os"
	"time"
)

type Config struct {
	BaseURL        *url.URL
	Debug          bool
	LoadTimeout    time.Duration
	RequestTimeout time.Duration

	InfoLogger  *log.Logger
	DebugLogger *log.Logger
}

func NewConfig() *Config {
	baseURL, err := url.Parse("http://localhost")
	if err != nil {
		panic(err)
	}

	return &Config{
		BaseURL:        baseURL,
		Debug:          false,
		LoadTimeout:    60 * time.Second,
		RequestTimeout: 30 * time.Second,
	}
}

func (c *Config) SetupLogger() {
	c.InfoLogger = log.New(os.Stdout, "info  ", log.Ltime)
	c.DebugLogger = log.New(os.Stderr, "debug ", log.Ltime)
	if !c.Debug {
		c.DebugLogger.SetOutput(io.Discard)
	}
}
