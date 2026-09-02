package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/events"
	"github.com/carboncircuit/backend/internal/outbox"
	"github.com/carboncircuit/backend/services/provenance-service/internal/domain"
	"github.com/carboncircuit/backend/services/provenance-service/internal/repository"
)

const usage = `usage: republish -organization <uuid>

Re-appends batch.created and checkpoint.logged to the outbox for every batch
an organization owns, so a read-side projection can be rebuilt from scratch.
`

func main() {
	organization := flag.String("organization", "", "organization whose batches to republish")
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	flag.Parse()

	if err := run(*organization); err != nil {
		fmt.Fprintln(os.Stderr, "republish failed:", err)
		os.Exit(1)
	}
}

func run(organization string) error {
	organizationID, err := uuid.Parse(organization)
	if err != nil {
		return fmt.Errorf("organization must be a uuid: %w", err)
	}

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		return fmt.Errorf("DATABASE_DSN is required")
	}

	ctx := context.Background()

	store, err := database.Open(ctx, database.Options{
		DSN:             dsn,
		Schema:          "provenance",
		MaxOpenConns:    4,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: time.Minute,
		AcquireTimeout:  5 * time.Second,
	})
	if err != nil {
		return err
	}
	defer database.Close(store)

	batches := repository.NewBatchRepository()
	checkpoints := repository.NewCheckpointRepository()

	republished := 0

	err = database.WithinTenant(ctx, store, database.TenantContext{
		OrganizationID: organizationID.String(),
	}, func(tx database.Tx) error {
		owned, listErr := batches.List(tx, organizationID, "", 1000)
		if listErr != nil {
			return listErr
		}

		for _, batch := range owned {
			if appendErr := appendBatch(tx, batch); appendErr != nil {
				return appendErr
			}

			logged, checkpointErr := checkpoints.ListForBatch(tx, batch.ID)
			if checkpointErr != nil {
				return checkpointErr
			}

			for _, checkpoint := range logged {
				if appendErr := appendCheckpoint(tx, batch, checkpoint); appendErr != nil {
					return appendErr
				}
			}

			republished++
		}

		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("republish: organization=%s batches=%d\n", organizationID, republished)

	return nil
}

func componentsOf(batch domain.Batch) []events.ScoreComponent {
	var components []events.ScoreComponent
	if len(batch.ScoreComponents) > 0 {
		_ = json.Unmarshal(batch.ScoreComponents, &components)
	}
	return components
}

func appendBatch(tx database.Tx, batch domain.Batch) error {
	_, err := outbox.Append(tx, outbox.Envelope{
		AggregateType: "batch",
		AggregateID:   batch.ID,
		EventType:     events.TopicBatchCreated,
		Payload: events.BatchCreated{
			BatchID:                    batch.ID.String(),
			OrganizationID:             batch.OrganizationID.String(),
			PublicReference:            batch.PublicReference,
			ProductCategory:            string(batch.ProductCategory),
			ComponentType:              batch.ComponentType,
			OriginatingFacilityName:    batch.OriginatingFacilityName,
			OriginatingFacilityCountry: batch.OriginatingFacilityCountry,
			ProducedAt:                 batch.ProducedAt.UTC().Format(time.RFC3339),
			ProvenanceScore:            batch.ProvenanceScore,
			ScoreComponents:            componentsOf(batch),
			OccurredAt:                 time.Now().UTC().Format(time.RFC3339),
		},
	})
	return err
}

func appendCheckpoint(
	tx database.Tx,
	batch domain.Batch,
	checkpoint domain.Checkpoint,
) error {
	logged := events.CheckpointLogged{
		BatchID:         batch.ID.String(),
		CheckpointID:    checkpoint.ID.String(),
		OrganizationID:  batch.OrganizationID.String(),
		Type:            string(checkpoint.Type),
		LocationLabel:   checkpoint.LocationLabel,
		CountryCode:     checkpoint.CountryCode,
		OccurredAt:      checkpoint.OccurredAt.UTC().Format(time.RFC3339),
		AnchorStatus:    string(checkpoint.AnchorStatus),
		ProvenanceScore: batch.ProvenanceScore,
		ScoreComponents: componentsOf(batch),
	}

	if checkpoint.ShippingMethod != nil {
		logged.ShippingMethod = *checkpoint.ShippingMethod
	}
	if checkpoint.SupersedesCheckpointID != nil {
		logged.SupersedesCheckpointID = checkpoint.SupersedesCheckpointID.String()
	}

	_, err := outbox.Append(tx, outbox.Envelope{
		AggregateType: "batch",
		AggregateID:   batch.ID,
		EventType:     events.TopicCheckpointLogged,
		Payload:       logged,
	})

	return err
}
