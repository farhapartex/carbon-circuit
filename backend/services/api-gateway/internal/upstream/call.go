package upstream

import (
	"context"
	"time"

	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/logging"
)

func callContext(
	ctx context.Context,
	idempotencyKey string,
	timeout time.Duration,
) (context.Context, context.CancelFunc) {
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	callCtx = grpcx.WithCorrelationID(callCtx, logging.CorrelationIDFrom(ctx))
	callCtx = grpcx.WithIdempotencyKey(callCtx, idempotencyKey)
	return callCtx, cancel
}
