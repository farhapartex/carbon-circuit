package ratelimit_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/carboncircuit/backend/internal/ratelimit"
)

func redisClient(t *testing.T) *redis.Client {
	t.Helper()

	server := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: server.Addr()})
}

func TestADayScopedRuleDoesNotDivideByZero(t *testing.T) {
	limiter, err := ratelimit.New(redisClient(t), "test", []ratelimit.Rule{
		{
			Name:      "daily",
			PerDay:    10,
			Burst:     1,
			KeyFunc:   func(ratelimit.Request) string { return "org:one" },
			AppliesTo: func(ratelimit.Request) bool { return true },
		},
	})
	if err != nil {
		t.Fatalf("build limiter with a day rule: %v", err)
	}

	decision, err := limiter.Check(context.Background(), ratelimit.Request{})
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("the first request under a daily rule must be allowed")
	}
}

func TestARuleWithNoRateIsRefusedAtConstruction(t *testing.T) {
	_, err := ratelimit.New(redisClient(t), "test", []ratelimit.Rule{
		{
			Name:      "broken",
			Burst:     1,
			KeyFunc:   func(ratelimit.Request) string { return "key" },
			AppliesTo: func(ratelimit.Request) bool { return true },
		},
	})
	if err == nil {
		t.Fatal("a rule with neither rate set must be refused rather than panic")
	}
}

func TestADailyRuleExhaustsAfterItsAllowance(t *testing.T) {
	limiter, err := ratelimit.New(redisClient(t), "test", []ratelimit.Rule{
		{
			Name:      "daily",
			PerDay:    3,
			Burst:     0,
			KeyFunc:   func(ratelimit.Request) string { return "org:one" },
			AppliesTo: func(ratelimit.Request) bool { return true },
		},
	})
	if err != nil {
		t.Fatalf("build limiter: %v", err)
	}

	refused := false
	for attempt := 0; attempt < 10; attempt++ {
		decision, checkErr := limiter.Check(context.Background(), ratelimit.Request{})
		if checkErr != nil {
			t.Fatalf("check: %v", checkErr)
		}
		if !decision.Allowed {
			refused = true
			break
		}
	}

	if !refused {
		t.Fatal("a daily allowance must eventually refuse")
	}
}

func TestMinuteAndDayRulesCoexist(t *testing.T) {
	limiter, err := ratelimit.New(redisClient(t), "test", []ratelimit.Rule{
		{
			Name:      "per_minute",
			PerMinute: 60,
			Burst:     10,
			KeyFunc:   func(ratelimit.Request) string { return "ip:one" },
			AppliesTo: func(ratelimit.Request) bool { return true },
		},
		{
			Name:      "per_day",
			PerDay:    10,
			Burst:     1,
			KeyFunc:   func(ratelimit.Request) string { return "org:one" },
			AppliesTo: func(ratelimit.Request) bool { return true },
		},
	})
	if err != nil {
		t.Fatalf("a mix of minute and day rules must build: %v", err)
	}

	if _, err := limiter.Check(context.Background(), ratelimit.Request{}); err != nil {
		t.Fatalf("check: %v", err)
	}
}
