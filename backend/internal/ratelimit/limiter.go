package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store/goredisstore.v9"
)

type Rule struct {
	Name      string
	PerMinute int
	Burst     int
	KeyFunc   func(Request) string
	AppliesTo func(Request) bool
}

type Request struct {
	CallerClass    string
	CallerKey      string
	EndpointClass  string
	ClientIP       string
	OrganizationID string
}

type Decision struct {
	Allowed    bool
	Rule       string
	Limit      int
	Remaining  int
	ResetAfter time.Duration
	RetryAfter time.Duration
}

type Limiter struct {
	limiters map[string]*throttled.GCRARateLimiterCtx
	rules    []Rule
}

func New(client *redis.Client, keyPrefix string, rules []Rule) (*Limiter, error) {
	store, err := goredisstore.NewCtx(client, keyPrefix)
	if err != nil {
		return nil, fmt.Errorf("create rate limit store: %w", err)
	}

	limiters := make(map[string]*throttled.GCRARateLimiterCtx, len(rules))
	for _, rule := range rules {
		quota := throttled.RateQuota{
			MaxRate:  throttled.PerMin(rule.PerMinute),
			MaxBurst: rule.Burst,
		}
		limiter, limiterErr := throttled.NewGCRARateLimiterCtx(store, quota)
		if limiterErr != nil {
			return nil, fmt.Errorf("create limiter %q: %w", rule.Name, limiterErr)
		}
		limiters[rule.Name] = limiter
	}

	return &Limiter{limiters: limiters, rules: rules}, nil
}

func (l *Limiter) Check(ctx context.Context, request Request) (Decision, error) {
	tightest := Decision{Allowed: true, Remaining: -1}

	for _, rule := range l.rules {
		if rule.AppliesTo != nil && !rule.AppliesTo(request) {
			continue
		}

		key := rule.KeyFunc(request)
		if key == "" {
			continue
		}

		limited, result, err := l.limiters[rule.Name].RateLimitCtx(ctx, key, 1)
		if err != nil {
			return Decision{Allowed: true, Remaining: -1}, err
		}

		decision := Decision{
			Allowed:    !limited,
			Rule:       rule.Name,
			Limit:      result.Limit,
			Remaining:  result.Remaining,
			ResetAfter: result.ResetAfter,
			RetryAfter: result.RetryAfter,
		}

		if limited {
			return decision, nil
		}

		if tightest.Remaining < 0 || decision.Remaining < tightest.Remaining {
			tightest = decision
		}
	}

	return tightest, nil
}
