package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
)

var (
	ErrEmpty = errors.New("empty string")
)

type serviceLoggingMiddleware struct {
	logger log.Logger
	next   StringService
}

func (m serviceLoggingMiddleware) Uppercase(s string) (string, error) {
	begin := time.Now()
	output, err := m.next.Uppercase(s)
	m.logger.Log(
		"method", "uppercase",
		"input", s,
		"output", output,
		"err", err,
		"took", time.Since(begin),
	)
	return output, err
}

func (m serviceLoggingMiddleware) Count(s string) int {
	begin := time.Now()
	output := m.next.Count(s)
	m.logger.Log(
		"method", "count",
		"input", s,
		"output", output,
		"took", time.Since(begin),
	)
	return output
}

type instrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	countResult    metrics.Histogram
	next           StringService
}

func (mw instrumentingMiddleware) Uppercase(s string) (output string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "uppercase", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.next.Uppercase(s)
	return
}

func (mw instrumentingMiddleware) Count(s string) (n int) {
	defer func(begin time.Time) {
		lvs := []string{"method", "count", "error", "false"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
		mw.countResult.Observe(float64(n))
	}(time.Now())

	n = mw.next.Count(s)
	return
}

// StringService provides operations on strings
type StringService interface {
	Uppercase(s string) (string, error)
	Count(s string) int
}

type stringService struct{}

func (stringService) Uppercase(s string) (string, error) {
	if s == "" {
		return "", ErrEmpty
	}

	return strings.ToUpper(s), nil
}

func (stringService) Count(s string) int {
	return len(s)
}
