package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/billing-service/internal/domain"
	"github.com/carboncircuit/backend/services/billing-service/internal/repository"
	"github.com/carboncircuit/backend/services/billing-service/internal/service"
)

func store(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_BILLING_DSN")
	if dsn == "" {
		t.Skip("set TEST_BILLING_DSN to run subscription integration tests")
	}

	opened, err := database.Open(context.Background(), database.Options{
		DSN:             dsn,
		Schema:          "billing",
		MaxOpenConns:    8,
		MaxIdleConns:    4,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: time.Minute,
		AcquireTimeout:  5 * time.Second,
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	return opened
}

func creator(handle *gorm.DB) *service.SubscriptionCreator {
	return service.NewSubscriptionCreator(
		handle,
		repository.NewSubscriptionRepository(handle),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func organization(t *testing.T, handle *gorm.DB) uuid.UUID {
	t.Helper()

	id := uuid.New()

	t.Cleanup(func() {
		err := database.WithinTenant(
			context.Background(),
			handle,
			database.TenantContext{OrganizationID: id.String()},
			func(tx database.Tx) error {
				if err := tx.Session().Exec(
					`DELETE FROM billing.subscriptions WHERE organization_id = ?`, id).Error; err != nil {
					return err
				}
				return tx.Session().Exec(
					`DELETE FROM billing.idempotency_records WHERE organization_id = ?`, id).Error
			},
		)
		if err != nil {
			t.Errorf("clean subscriptions for %s: %v", id, err)
		}
	})

	return id
}

func enrolment(id uuid.UUID, orgType domain.OrganizationType, tier domain.PlanTier, key string) service.Enrolment {
	return service.Enrolment{
		OrganizationID:   id,
		OrganizationType: orgType,
		Tier:             tier,
		IdempotencyKey:   key,
		RequestBody:      []byte(id.String() + "|" + string(tier)),
	}
}

func TestPaidPlanCreatesActiveSubscription(t *testing.T) {
	handle := store(t)
	id := organization(t, handle)

	enrolled, err := creator(handle).Create(context.Background(),
		enrolment(id, domain.OrganizationManufacturer, domain.TierGrowth, "sub-key-1"))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if enrolled.Subscription.State != domain.SubscriptionActive {
		t.Fatalf("expected an active subscription, got %q", enrolled.Subscription.State)
	}
	if enrolled.Subscription.Plan.Tier != domain.TierGrowth {
		t.Fatalf("expected growth, got %q", enrolled.Subscription.Plan.Tier)
	}

	period := enrolled.Subscription.CurrentPeriodEnd.Sub(enrolled.Subscription.CurrentPeriodStart)
	if period < 27*24*time.Hour || period > 32*24*time.Hour {
		t.Fatalf("expected roughly a one month period, got %s", period)
	}
}

func TestFreePlanIsOnlyForCreditBuyers(t *testing.T) {
	handle := store(t)

	buyer := organization(t, handle)
	if _, err := creator(handle).Create(context.Background(),
		enrolment(buyer, domain.OrganizationCreditBuyer, domain.TierBuyer, "sub-key-2")); err != nil {
		t.Fatalf("credit buyer should be able to take the free plan: %v", err)
	}

	manufacturer := organization(t, handle)
	_, err := creator(handle).Create(context.Background(),
		enrolment(manufacturer, domain.OrganizationManufacturer, domain.TierBuyer, "sub-key-3"))
	if !errors.Is(err, service.ErrPlanNotAllowed) {
		t.Fatalf("expected ErrPlanNotAllowed for a manufacturer on the buyer tier, got %v", err)
	}
}

func TestSecondSubscriptionIsRefused(t *testing.T) {
	handle := store(t)
	id := organization(t, handle)
	subscribe := creator(handle)

	if _, err := subscribe.Create(context.Background(),
		enrolment(id, domain.OrganizationManufacturer, domain.TierStarter, "sub-key-4")); err != nil {
		t.Fatalf("first subscription: %v", err)
	}

	_, err := subscribe.Create(context.Background(),
		enrolment(id, domain.OrganizationManufacturer, domain.TierGrowth, "sub-key-5"))
	if !errors.Is(err, service.ErrSubscriptionExists) {
		t.Fatalf("expected ErrSubscriptionExists, got %v", err)
	}
}

func TestRetryReplaysWithoutASecondSubscription(t *testing.T) {
	handle := store(t)
	id := organization(t, handle)
	subscribe := creator(handle)
	request := enrolment(id, domain.OrganizationManufacturer, domain.TierStarter, "sub-key-6")

	first, err := subscribe.Create(context.Background(), request)
	if err != nil {
		t.Fatalf("first: %v", err)
	}

	second, err := subscribe.Create(context.Background(), request)
	if err != nil {
		t.Fatalf("retry: %v", err)
	}

	if !second.Replayed {
		t.Fatal("expected the retry to replay")
	}
	if second.Subscription.ID != first.Subscription.ID {
		t.Fatal("expected the same subscription to come back")
	}

	var count int64
	err = database.WithinTenant(context.Background(), handle,
		database.TenantContext{OrganizationID: id.String()},
		func(tx database.Tx) error {
			return tx.Session().Table("billing.subscriptions").
				Where("organization_id = ?", id).Count(&count).Error
		})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly one subscription, got %d", count)
	}
}
