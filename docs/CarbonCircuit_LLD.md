# CarbonCircuit — LLD Global Conventions & Cross-Cutting Standards

**Status:** Foundational document for all Low-Level Design work. Every domain-specific LLD that follows (Edge & Identity, Provenance, Sustainability & Credits, AI Review, Marketplace, Blockchain Integration, Trust & Safety/Platform, Billing) references this document instead of re-defining these rules per service.

This remains system design, not code — it defines patterns, structures, and conventions precisely enough to implement against, without writing implementation code itself.

---

## 1. API Response Envelope & Error Handling

### 1.1 Success Envelope

Every successful response, from every service exposed through the API Gateway, wraps its payload identically:

```
{ "data": <object | array> }
```

For paginated list endpoints, a `meta` sibling is added alongside `data`:

```
{ "data": [...], "meta": { "page": 1, "per_page": 25, "total_items": 340, "total_pages": 14 } }
```

For cursor-paginated endpoints:

```
{ "data": [...], "meta": { "next_cursor": "eyJ...", "has_more": true } }
```

No endpoint returns a bare object or array at the top level — everything is wrapped, with no exceptions, so client code never has to special-case a response shape.

### 1.2 Error Envelope

```
{ "error": { "code": "RESOURCE_NOT_FOUND", "message": "The requested batch could not be found.", "request_id": "01J..." } }
```

Where an error is field-specific, a `details` array is included carrying `{ "field": "...", "code": "..." }` entries — enough for a form to highlight the offending input, and never more:

```
{ "error": { "code": "VALIDATION_ERROR", "message": "One or more fields are invalid.", "request_id": "01J...",
  "details": [ { "field": "quantity", "code": "BELOW_MINIMUM" } ] } }
```

**Hard rule: the client never sees internal error detail.** No stack traces, no database error strings, no internal service names, no raw exception messages ever reach the `message` field. Every error surfaced to a client is one of a fixed, documented set of codes with a pre-written, generic, user-safe message. The `request_id` — the same correlation ID used in tracing and logging (Section 8) — is always included so a user can quote it to support and an Admin can look up the *actual* underlying error in Loki or Tempo. That is the mechanism giving admins full detail while clients get nothing sensitive.

### 1.3 Error Code Taxonomy

Maintained centrally so codes stay consistent across all fifteen services. Any new domain-specific case gets a new, equally generic, documented code added to this table rather than an ad hoc string invented per endpoint.

| Code | HTTP | Meaning |
|---|---|---|
| `VALIDATION_ERROR` | 400 | Request payload failed schema or field validation |
| `UNAUTHENTICATED` | 401 | Missing or invalid credentials |
| `TOKEN_REVOKED` | 401 | Credential was valid but has been revoked |
| `FORBIDDEN` | 403 | Authenticated, but not permitted to perform this action |
| `ORGANIZATION_NOT_VERIFIED` | 403 | Action requires a verified organization |
| `ORGANIZATION_RESTRICTED` | 403 | Organization is restricted following a fraud escalation |
| `ORGANIZATION_READ_ONLY` | 403 | Subscription lapsed past the grace period |
| `PLAN_LIMIT_EXCEEDED` | 403 | Plan quota exhausted; body names the specific limit |
| `MFA_REQUIRED` | 403 | Action requires multi-factor re-authentication |
| `RESOURCE_NOT_FOUND` | 404 | Resource doesn't exist or isn't visible to this caller |
| `CONFLICT` | 409 | Request conflicts with current state (claim already decided, listing already filled) |
| `IDEMPOTENCY_KEY_REUSED` | 409 | Same key presented with a different request body |
| `REQUEST_IN_PROGRESS` | 409 | An identical request with this key is currently executing |
| `EVIDENCE_NOT_READY` | 409 | Referenced evidence has not completed scanning |
| `PAYLOAD_TOO_LARGE` | 413 | Request body or uploaded file exceeds its limit |
| `UNSUPPORTED_MEDIA_TYPE` | 415 | File type not accepted, or declared type doesn't match content |
| `IDEMPOTENCY_KEY_REQUIRED` | 422 | A mutating request arrived without an `Idempotency-Key` header |
| `RATE_LIMITED` | 429 | Caller exceeded a rate limit; `Retry-After` always present |
| `INTERNAL_ERROR` | 500 | Unclassified server-side failure |
| `DEPENDENCY_UNAVAILABLE` | 503 | A required internal dependency is circuit-broken; retry later |
| `CAPACITY_SHED` | 503 | Global load shedding is active on this endpoint class |
| `GATEWAY_TIMEOUT` | 504 | Request exceeded the Gateway deadline |

Every service's internal errors are mapped to one of these at the boundary (handler layer, Section 3) before leaving the service — an internal error type never leaks through directly.

### 1.4 Internal (Admin-Visible) Error Detail

The full internal error — actual exception, stack trace, failing query, downstream response — is always logged with the same `request_id`, and never discarded. Admin Backend can look up full detail by `request_id` or by browsing structured logs. This is the only path to real error internals; there is no debug-mode flag on any client-facing endpoint that would expose more, since that would defeat the purpose.

---

## 2. Endpoint & Resource Naming

### 2.1 Pluralization

Endpoints are **plural by default**, representing a collection: `/batches`, `/checkpoints`, `/claims`, `/organizations`, `/facilities`, `/credits`, `/listings`, `/trades`, `/retirements`, `/notifications`, `/fraud-flags`.

**Singular is the documented exception**, used only for:
- A resource inherently one-per-caller or one-per-parent with no meaningful collection: `/me`, `/organizations/{orgId}/billing`, `/organizations/{orgId}/plan`, `/organizations/{orgId}/treasury-address`
- Action-style endpoints that don't map to a CRUD resource: `/auth/session`, `/claims/{claimId}/decision`, `/listings/{listingId}/retire`

If in doubt, default to plural.

### 2.2 Structure

- All endpoints are versioned: `/v1/...`
- Nesting reflects real ownership, capped at two levels: `/v1/organizations/{orgId}/facilities` is fine; a batch's checkpoints are addressed as `/v1/batches/{batchId}/checkpoints`, not nested under org and facility as well.
- List endpoints accept `page` and `per_page` (default 25, max 100). High-volume, append-mostly collections — checkpoints, chain events, notifications, the retirement log — use **cursor pagination** (`after`, an opaque cursor) instead, since offset pagination degrades badly on large growing tables and can skip or repeat rows when the underlying set changes between pages.
- Filtering uses `filter[field]=value`; sorting uses `sort=field` / `sort=-field`.
- Every `POST`, `PUT`, and `PATCH` that isn't purely idempotent by nature requires an `Idempotency-Key` header (Section 7).
- Bulk endpoints use a `:batch` suffix — `/v1/checkpoints:batch` — signalling that the response is a per-item result array rather than a single resource.

### 2.3 Bulk Endpoint Semantics

Bulk ingestion is partial-success by design, because rejecting 499 valid checkpoints because one had a malformed timestamp would make the endpoint unusable for its actual purpose. A bulk request returns `207` with a per-item result:

```
{ "data": { "accepted": 498, "rejected": 2,
    "results": [ { "external_id": "WMS-88213", "status": "created", "id": "01J..." },
                 { "external_id": "WMS-88214", "status": "duplicate", "id": "01J..." },
                 { "external_id": "WMS-88215", "status": "rejected", "error": { "code": "VALIDATION_ERROR", "details": [...] } } ] } }
```

`duplicate` is a success outcome, not an error — it means this `external_id` was already ingested and the existing record's ID is returned. A partner replaying a day of events after an outage receives a response of entirely `duplicate` results and can reconcile confidently rather than having to guess whether anything was double-written.

---

## 3. Layered Architecture (Applies Inside Every Service)

Every Go service, regardless of domain, is structured in the same three layers, communicating only downward through interfaces.

**Handler layer** — parses and validates the incoming request, maps it to a typed input struct, calls exactly one method on a Service-layer interface, and maps the result or error to the standard envelope. Concretely: the **API Gateway** is the only service running a Gin HTTP handler layer, since it's the only service accepting external REST/JSON; every other service's handler layer is a **gRPC service handler** (Section 12). **No business logic, no direct database or cache access, no direct Kafka production ever happens in this layer**, regardless of which kind it is. Its only job is translation between wire format and a typed call into the service layer.

**Service layer** — owns all business logic: validation beyond schema shape, orchestration across repositories, calling other services through gRPC client interfaces, publishing events through a Publisher interface, enforcing business rules (credit ceilings, plan quotas, claim state transitions, remainder rules). The service layer depends only on interfaces — never a concrete GORM implementation, a concrete Kafka client, or a concrete Redis client. This is what makes it independently testable and is the core Dependency Inversion application throughout the codebase.

**Repository layer** — the only layer that talks to the database or Redis directly. One repository per aggregate. Repository interfaces expose domain-meaningful methods (`FindActiveByOrganization(ctx, orgID)`) and never leak GORM types or raw SQL structures upward.

`ai-agent-service`, being Python, follows the same three-layer separation with the same interface boundaries — a gRPC/consumer entry layer, a graph-orchestration layer holding the business rules, and a repository layer over Postgres and pgvector. The language differs; the layering discipline does not.

### 3.1 SOLID, Concretely Applied

- **Single Responsibility:** each Service-layer struct owns exactly one aggregate's business logic — `ClaimIntakeService`, `AIReviewOrchestrator`, `VerifierDecisionService` remain separate structs inside the merged Sustainability Service, not one god-object.
- **Open/Closed:** new behavior arrives as a new interface implementation. A new sustainability activity type's formula is a new `CreditFormulaCalculator` registered into a formula registry, not a new branch in existing formula code.
- **Liskov Substitution:** any implementation — a mock repository in tests, a GORM repository in production — is fully substitutable. Interfaces are defined by behavior contract, not by what one implementation happens to need.
- **Interface Segregation:** narrow, purpose-specific interfaces. A `ClaimReader` and a `ClaimWriter`, not one fat `ClaimRepository` with twenty methods most callers don't use.
- **Dependency Inversion:** every Service-layer struct receives its dependencies via constructor injection — nothing is instantiated internally or reached as a global singleton, except the logger and config loader, which are the one deliberate exception since they are genuinely cross-cutting and stateless.

### 3.2 DRY — Shared Internal Module

A shared internal Go module, imported by every Go service rather than duplicated, provides: the response envelope helper, the error-code mapper, request validation helpers, correlation-ID middleware, the GORM base model, the tenant-context propagator, a Redis client wrapper implementing the cache and distributed-lock interfaces, the transactional outbox and inbox helpers, the idempotency helper, and the gRPC interceptor stack. Any cross-cutting concern implemented twice across two services is treated as a bug to fix by extraction, not an acceptable duplication.

---

## 4. Complexity Guidelines

- Every list endpoint is paginated. There is no endpoint that can return an unbounded result set — a hard rule, not a style preference, since an unbounded query is both a latency risk and a trivial denial-of-service vector.
- Lookups that matter for request latency are backed by an index guaranteeing O(log n) or better. No code path relies on a full table scan for a request-path operation.
- N+1 query patterns are avoided deliberately: `Preload` or explicit joins wherever a list response needs related data, never a loop issuing one query per row.
- Bulk operations use bulk insert statements with `ON CONFLICT DO NOTHING` against the natural key, not per-row loops.
- Background and async processing always runs through a bounded worker pool — no code path spawns one goroutine per incoming item unboundedly. Concurrency is always capped and configurable.
- Any request-path query whose plan involves a sequential scan on a table expected to exceed 100,000 rows is a blocking code review finding.

---

## 5. Database Design Conventions

- **Primary keys are UUIDv7** (time-ordered) rather than sequential integers or random UUIDv4. This avoids a centralized sequence generator — important since many services generate IDs independently and concurrently — while giving good B-tree locality, since UUIDv7 sorts by creation time, unlike UUIDv4 which fragments indexes at scale.
- **Publicly exposed identifiers are never primary keys.** Any identifier appearing in an unauthenticated URL — the public batch reference above all — is a separately generated 128-bit random value rendered in base62, stored in its own uniquely-indexed column. UUIDv7's time ordering is a benefit internally and a leak externally: a party holding a handful of them can infer platform-wide creation ordering and volume, which is commercially sensitive on a platform where competitors are co-tenants.
- **Every table includes the standard columns** via the shared base model: `id`, `created_at`, `updated_at`, `deleted_at` (nullable, soft delete), and `version` (integer, starting at 1, incremented on every update — the backbone of optimistic concurrency, Section 10).
- **Every tenant-scoped table includes `organization_id`**, indexed, with Row-Level Security enabled (Section 11.2).
- **Every foreign key column is indexed.** Non-negotiable — an unindexed FK is among the most common causes of slow joins and slow cascading deletes.
- **Naming:** `snake_case` for tables and columns, plural table names.
- **Monetary and credit-amount fields use `NUMERIC`**, never `FLOAT` or `DOUBLE`. Credit amounts are `NUMERIC(28,6)` and USDC amounts `NUMERIC(20,6)`, matching the precision the contracts settle at so no rounding difference can arise between the two representations.
- **Status and enum-like fields** use a Postgres native `ENUM` type rather than free-text varchar, so invalid states are rejected at the database layer as well as the application layer.
- **Timestamps are `TIMESTAMPTZ`**, always stored in UTC. A supply chain spanning time zones has no room for an ambiguous local timestamp.
- **Category-specific flexible attributes** use a `JSONB` column with a `GIN` index only where that column is genuinely queried. It is never a substitute for proper relational columns on data that is structured and frequently filtered.
- **Vector columns** use pgvector's `halfvec` type with an HNSW index, in the `ai_agent` schema only (Section 11.5).

---

## 6. Indexing Strategy

- Every foreign key: indexed.
- Every column used in a `WHERE`, `JOIN`, or `ORDER BY` on a request-path query: indexed.
- Composite indexes are ordered **equality columns first, range columns last** — an index on `(organization_id, created_at DESC)` for "list this org's batches, newest first," since `organization_id` is always an equality filter and `created_at` is the range and sort column.
- Soft-deleted rows are excluded via a **partial index** (`WHERE deleted_at IS NULL`) on any index backing a frequently-run active-records query, keeping the index smaller and faster than indexing every historical row.
- Uniqueness constraints double as idempotency backstops: a unique constraint on `(claim_id)` in the credit issuance table physically prevents a second issuance row for the same claim regardless of what application logic does or doesn't catch. The full list of these backstops:

| Constraint | Table | Guarantees |
|---|---|---|
| `UNIQUE (claim_id)` | `credit_ledger.issuances` | One issuance per approved claim |
| `UNIQUE (organization_id, external_id)` | `provenance.batches` | Partner ingest replay safety |
| `UNIQUE (organization_id, external_id)` | `provenance.checkpoints` | Partner ingest replay safety |
| `UNIQUE (public_ref)` | `provenance.batches` | Public reference collision impossible |
| `UNIQUE (organization_id, endpoint, idempotency_key)` | every service's `idempotency_records` | Edge idempotency |
| `UNIQUE (event_id)` | every service's `inbox_events` | Consumer idempotency |
| `UNIQUE (epoch)` | `chain_observer.epoch_anchors` | One anchor per epoch |
| `UNIQUE (listing_id, fill_sequence)` | `marketplace.trades` | One trade row per fill |
| `UNIQUE (claim_id, verifier_user_id)` | `sustainability.claim_decisions` | A verifier cannot approve the same claim twice toward the dual-approval threshold |
| `UNIQUE (signing_key, nonce)` | `chain_writer.transactions` | One transaction per nonce per key |
| `UNIQUE (stripe_event_id)` | `billing.webhook_events` | Stripe redelivery safety |

- Every domain LLD specifies exact indexes per table following these rules rather than restating them.

---

## 7. Idempotency Implementation

This section makes HLD Section 9 concrete. It is the most safety-critical mechanism in the system, and the details that make it correct are exactly the ones easily lost.

### 7.1 The Idempotency Record

Every service exposing mutating operations owns an `idempotency_records` table in its own schema:

| Column | Type | Notes |
|---|---|---|
| `id` | UUID | |
| `organization_id` | UUID | Scope — keys are never global |
| `endpoint` | TEXT | Scope — the same key on a different endpoint is a different operation |
| `idempotency_key` | TEXT | Client-supplied |
| `request_hash` | BYTEA | SHA-256 of the canonicalized request body |
| `state` | ENUM | `processing` / `completed` / `failed` |
| `response_status` | INT | Populated on completion |
| `response_body` | JSONB | Populated on completion |
| `resource_id` | UUID | The created or affected resource, for fast linking |
| `created_at`, `completed_at` | TIMESTAMPTZ | |

`UNIQUE (organization_id, endpoint, idempotency_key)`.

### 7.2 The Flow

1. **Reserve.** `INSERT ... ON CONFLICT DO NOTHING` a record in state `processing`. If the insert succeeds, this request owns the operation. If it doesn't, read the existing record and branch:
   - `state = processing` → return `409 REQUEST_IN_PROGRESS` with `Retry-After: 2`
   - `state = completed` and `request_hash` matches → replay the stored response verbatim
   - `state = completed` and `request_hash` differs → return `409 IDEMPOTENCY_KEY_REUSED`
   - `state = failed` → transition back to `processing` and proceed; a retry after a genuine failure should genuinely retry
2. **Execute.** Run the business operation.
3. **Complete.** Update the record to `completed` with the response, **in the same database transaction as the business write.** This is the property that makes the guarantee real: there is no window in which the effect exists and the record doesn't, or vice versa.
4. **Fail.** On error, mark the record `failed` in a separate transaction, so the failure is recorded even though the business transaction rolled back.

Step 1's atomic reserve is what handles the case a naive implementation misses. Two identical requests arriving five milliseconds apart both find no cached result and both execute — unless the reservation is a single atomic insert that exactly one of them can win.

**Redis fronts this as a read-through cache** on the completed-lookup path only, with a 48-hour TTL. The Postgres record is the source of truth. A financial idempotency guarantee cannot rest on a store that may lose its most recent writes on failover, and this is precisely the situation where those lost seconds would matter.

**Retention:** 30 days, swept by a scheduled job. Beyond that a retry is so far outside any plausible client retry window that treating it as a new request is correct.

### 7.3 The Inbox Record

Every service consuming events owns an `inbox_events` table:

| Column | Type |
|---|---|
| `event_id` | UUID, `UNIQUE` |
| `topic` | TEXT |
| `consumer_group` | TEXT |
| `processed_at` | TIMESTAMPTZ |

Consumption follows exactly this sequence, and no other:

```
BEGIN
  INSERT INTO inbox_events (event_id, topic, consumer_group) VALUES (...)
    -- unique violation here means already processed: ROLLBACK, ack, done
  <apply the business effect>
  <write any outbox events this effect produces>
COMMIT
ack the message
```

The insert, the effect, and any resulting outbox rows are one transaction. Recording the event before the effect gives at-most-once — a crash between them loses the effect permanently with no error raised anywhere, which is the worst failure mode available because it is silent. Recording after gives a duplicate on redelivery. Only the single-transaction form is correct, and it is the only form permitted.

Because the guard is the event ID rather than a processing timestamp, **DLQ replay is safe by construction**: replaying a message that was in fact already applied is a unique-violation no-op, not a duplicate effect.

### 7.4 Natural Keys

Partner-submitted batches and checkpoints carry a required `external_id`, uniquely constrained per organization. Bulk inserts use `ON CONFLICT (organization_id, external_id) DO NOTHING RETURNING`, and rows that conflict are reported as `duplicate` in the bulk response (Section 2.3). This survives a retry that regenerated its request headers, which a header-based key does not.

---

## 8. Correlation IDs & Tracing

Every request is assigned a `request_id` at the API Gateway if one isn't already present, propagated through every gRPC call header and every Kafka message header from that point on. Every log line, every trace span, and every error response includes it.

Kafka message headers carry `request_id`, `event_id`, `event_type`, `schema_version`, `produced_at`, and `organization_id` — so a consumer can trace, deduplicate, version-check, and tenant-scope without deserializing the payload.

This is the mechanism referenced throughout: how an Admin looks up what actually happened for an error a client saw only a generic message for, and how one user action is traced end to end across services.

---

## 9. Concurrency Patterns (Go-Specific Design)

- **Goroutines** are used in three deliberate patterns and only these three:
  1. A bounded pool of Kafka consumer worker goroutines per consumer group, sized to the partition count, not one goroutine per message.
  2. Scheduled background jobs on a ticker, each guarded by a distributed lock (Section 10) so only one instance executes a given job across all replicas.
  3. Bounded fan-out within a single request where independent sub-tasks can run in parallel, always using an errgroup-style pattern so a failure in one branch cancels the others through a shared context rather than leaking a goroutine.
- **Channels** are the bounded hand-off between a Kafka consumer goroutine and its worker pool. This is the backpressure mechanism: if workers fall behind, the channel fills and the consumer naturally slows its read rate rather than queuing unboundedly in memory. Channels also carry graceful-shutdown signalling, so in-flight work either completes or is cleanly abandoned within a shutdown deadline.
- **Mutexes** are reserved strictly for protecting in-memory state within a single process, and never as a substitute for cross-instance coordination. Concrete permitted uses: an in-memory hot-reloadable config and feature-flag cache (`RWMutex`, since reads vastly outnumber writes), and local metrics aggregation before flushing to Prometheus. If state must be consistent across multiple running instances, it does not belong behind a local mutex at all.
- **Transaction nonce sequencing is explicitly not a mutex case.** A per-process mutex over a shared signing key's nonce is correct with exactly one replica and silently wrong with two — both allocate the same nonce and one transaction is dropped. Nonces are allocated from a database row per signing key, taken under `SELECT ... FOR UPDATE` inside the same transaction that records the outbound transaction, so allocation and recording are atomic and correct at any replica count. Section 13 covers the full lifecycle.
- **Context propagation:** every request-scoped and job-scoped operation carries a `context.Context` for cancellation and deadline propagation. Every database call, gRPC call, and Kafka produce respects the inbound deadline — a slow downstream dependency causes a bounded failure mapped to `DEPENDENCY_UNAVAILABLE`, never an indefinite hang.

---

## 10. Locking & Concurrency Control

### 10.1 Optimistic Concurrency First

The default mechanism for concurrent updates to a single row is the `version` column: a write asserts the version it read and fails if it changed, and the service layer retries a bounded number of times. This costs nothing when there is no contention, which is the overwhelmingly common case, and it is the right default for entity updates.

### 10.2 Distributed Locking Where Optimistic Fails

Redsync-based distributed locking is used where a critical section spans more than one row or more than one system, and optimistic concurrency therefore cannot express the invariant. It is used **only** for a named, enumerated list of critical sections — never as a blanket default around every write, which would reintroduce exactly the latency problem it exists to avoid.

A lock is acquired with a 5–10 second TTL before entering the critical section and explicitly released immediately after; the TTL is a safety net against a crashed holder, not the primary release mechanism. Each acquisition carries a unique fencing token, and any write validated against the lock checks the token is still current — so a process that lost its lock to a TTL expiry during a long GC pause cannot perform a stale write believing it still holds it.

| Lock key | Service | Critical section |
|---|---|---|
| `lock:credit-ledger:issue:{claimId}` | Credit Ledger | Claim-consumption check plus issuance record |
| `lock:marketplace:listing:{listingId}` | Marketplace | Listing quantity decrement on a fill |
| `lock:billing:usage:{orgId}:{period}` | Billing | Usage counter increment against a quota |
| `lock:chain-writer:signer:{keyId}` | Chain Writer | Nonce allocation and submission, per signing key |
| `lock:identity:treasury:{orgId}` | Identity | Treasury address change state machine |
| `lock:job:{jobName}` | All | Scheduled job run window |

### 10.3 The Belt-and-Suspenders Rule

Every one of these critical sections **also** has a database-level unique or check constraint as a final backstop (Section 6). The lock prevents the vast majority of races cheaply; the constraint guarantees correctness in the rare edge case where a lock is bypassed — a Redis failover window, a bug in acquisition logic. Neither layer is sufficient alone in this financial system, mirroring the same principle applied at the contract layer.

---

## 11. Database Architecture

### 11.1 Logical Isolation Within a Shared Physical Database

The system runs on **one physical PostgreSQL instance** shared by every service, with data ownership preserved through schema-level isolation rather than physical database separation.

- **Every service that owns data gets its own schema**, named after its service identifier. The API Gateway is the exception — it is stateless and holds no relational data, using only Redis for rate-limit counters and caches. The fourteen schemas are: `identity`, `billing`, `provenance`, `evidence`, `provenance_read`, `sustainability`, `credit_ledger`, `ai_agent`, `marketplace`, `chain_writer`, `chain_observer`, `fraud_detection`, `notification`, `admin`. A service's tables live only in its own schema.
- **Every service connects with its own dedicated Postgres role**, granted access only to its own schema. The "no service reaches into another service's tables" rule is therefore enforced by the database's own permission system, not by code review discipline — even a bug or a compromised service cannot query another schema, because its credentials are incapable of it. This is what keeps "shared database" from quietly becoming "shared data model" over time.
- Cross-service data access still only ever happens via gRPC or Kafka. The shared instance does not become a backdoor.

**An honest statement of the tradeoff:** this is a deliberate choice to accept a single point of failure and a shared resource budget in exchange for a dramatically simpler operational footprint. Schema separation and per-service roles give genuine *access* isolation. They do not give *resource* isolation: every service still shares one WAL, one autovacuum daemon, one buffer pool, one lock table, and one IOPS budget, so a runaway sequential scan in one service can degrade every other. The mitigations are the query discipline in Section 4, per-service connection segmentation below, and query-level monitoring — not a claim that the isolation is complete. Moving a service to its own instance later is a connection-string change plus a data migration, and the per-schema boundary is what keeps that option open.

### 11.2 Row-Level Security

Every tenant-scoped table has RLS enabled with a policy comparing `organization_id` against a transaction-local setting:

```
ALTER TABLE provenance.batches ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON provenance.batches
  USING (organization_id = current_setting('app.organization_id')::uuid);
```

The repository layer sets `app.organization_id` via `SET LOCAL` at the start of every transaction, taken **from the verified internal service token and never from a request parameter**. `SET LOCAL` is transaction-scoped, so it is compatible with PgBouncer's transaction pooling and cannot leak between pooled requests.

This is the structural half of the multi-tenancy guarantee. Application-layer scoping remains the first line and produces better errors; RLS is what makes a query that forgets it return zero rows instead of another tenant's data. A raw query written under time pressure, an ad hoc join, a reporting endpoint added in a hurry — none of them can bypass a policy the database enforces.

Two roles carry `BYPASSRLS`: the `admin` role, used only for genuinely cross-tenant operations and logged as such, and the migration role. No customer-facing service role has it.

### 11.3 Connection Budget and PgBouncer

With one shared instance, all fifteen services' connection pools draw from the same finite Postgres connection budget simultaneously. Even a modest pool of 10 per service across 15 services is 150 connections before any service runs a second replica — well past a stock default.

**PgBouncer in transaction pooling mode is mandatory, not an optimization.** All services connect to PgBouncer, never directly to Postgres. It multiplexes many application connections onto far fewer real backend connections, which is what lets fifteen independently-scaling services survive peak hour on one instance.

| Setting | Value |
|---|---|
| Postgres `max_connections` | 400 |
| PgBouncer `max_client_conn` | 2,000 |
| PgBouncer `default_pool_size` (per service) | 20 |
| PgBouncer `reserve_pool_size` | 5 |
| PgBouncer `pool_mode` | `transaction` |
| Service `MaxOpenConns` | 10 |
| Service `MaxIdleConns` | 5 |
| Service `ConnMaxLifetime` | 30 minutes |
| Service `ConnMaxIdleTime` | 5 minutes |
| Connection acquire timeout | 250 ms |

PgBouncer is configured with **a separate pool per service**, so a connection-hungry service during an ingestion burst cannot starve an unrelated service. Without per-service segmentation, a shared instance plus a shared multiplexer would reintroduce exactly the noisy-neighbour problem schema separation was meant to avoid.

**Transaction pooling requires driver configuration, and getting this wrong produces confusing runtime failures rather than a clean error.** Transaction mode forbids session state, and pgx — underneath GORM — uses prepared statements by default, which are session state. Every service therefore sets pgx's query exec mode to `exec` (equivalently, disables implicit statement caching). Also unavailable under transaction pooling and used nowhere in the system: session-level advisory locks (Section 10 uses Redis instead), `LISTEN`/`NOTIFY`, and cursors held across transactions.

**Fail fast, don't queue indefinitely.** Connection acquisition has a 250 ms timeout; on exhaustion a service returns `DEPENDENCY_UNAVAILABLE` rather than letting requests pile up waiting for a connection that may not come. Paired with the circuit breaker, this keeps one service's connection pressure from degrading into a hung, cascading failure for its callers.

The **sum across all services' pools multiplied by replica count**, measured against PgBouncer's backend limit, is a first-class capacity planning number revisited whenever a service is added or a traffic profile changes materially — not a one-time calculation.

### 11.4 Migrations

Each service owns and runs its own migrations, scoped to its own schema, tracking its own history table (`identity.schema_migrations`, `provenance.schema_migrations`). There is no single shared migration set for "the database."

Migrations run against Postgres directly, not through PgBouncer, since DDL and transaction pooling interact badly. Every migration is required to be backward-compatible with the currently-deployed service version: add columns nullable or with defaults, never rename or drop in the same release that stops using something. A destructive change is always two releases — stop using it, then remove it — because a single-release rename makes rollback impossible exactly when it is most needed.

### 11.5 The `ai_agent` Schema and pgvector

The `ai_agent` schema holds the LangGraph checkpointer tables, run traces, and the vector store. The `vector` extension is enabled once at the instance level; only this schema uses it.

| Table | Purpose |
|---|---|
| `evidence_chunks` | Chunked evidence text with `halfvec(1024)` embeddings, HNSW-indexed, `organization_id`-scoped with RLS |
| `claim_summaries` | One embedding per decided claim, for precedent retrieval |
| `graph_checkpoints` | LangGraph state checkpoints keyed by run ID |
| `review_runs` | Full per-run audit trace — node states, model ID, program version, tool calls, token counts, cost |

HNSW parameters: `m = 16`, `ef_construction = 64`, query-time `ef_search = 40`. `halfvec` halves storage relative to full-precision vectors at negligible recall cost at this dimensionality, which matters because this is the fastest-growing table in the system.

Every similarity query carries an `organization_id` predicate and runs under the same RLS policy as every other tenant-scoped table. Vectors derived from confidential evidence are subject to exactly the same isolation guarantee as the evidence itself — this is the concrete reason the vector store lives here rather than in an external service whose tenant filtering would be a third party's implementation detail.

`review_runs` is retained 7 years. Everything else is retained 24 months.

### 11.6 Read Replicas

A read replica serves read-heavy paths: Provenance Read Service's query pool, reporting queries in Admin Backend, and data export generation. Writes always go to the primary. Services using the replica must tolerate replication lag; anything read-your-own-write goes to the primary explicitly. Public provenance reads tolerate lag comfortably, since the CDN in front of them already serves content up to 60 seconds old.

---

## 12. gRPC Service Design Conventions

### 12.1 What Runs gRPC

- **The API Gateway** is the only service accepting external traffic (REST/JSON over HTTPS). It translates every inbound request into a gRPC call to the appropriate internal service — there is no REST traffic anywhere behind the Gateway.
- **Every internal service exposes a gRPC server as its only inbound interface.** No internal service runs its own HTTP server for business traffic. This keeps the transport uniform system-wide and is what makes mTLS-everywhere straightforward to enforce consistently — one transport, one place to apply the security interceptor stack.
- **Kafka remains entirely separate from gRPC.** gRPC is exclusively the synchronous question-and-answer transport.

### 12.2 Proto Contract Organization

- One `.proto` package per service, named after its identifier (`carboncircuit.provenance.v1`, `carboncircuit.credit_ledger.v1`), living in a shared `/proto` directory versioned alongside the shared internal module, so every service compiles against the same generated stubs and no service hand-rolls a copy of another's client.
- Versioned in the package path (`.v1`) from day one, so a breaking change ships as `.v2` alongside the old version during a migration window rather than requiring a simultaneous deploy of every caller.
- **Shared message types**, defined once and imported everywhere: `CreditAmount` (string-encoded decimal, never a native float), `UsdcAmount`, `CreditClass` (facility, vintage, activity type), `PageRequest`/`PageResponse`, and an `ErrorDetail` carried in gRPC status metadata that maps directly to the REST taxonomy in Section 1.3 — so an error originating deep in a call chain surfaces at the Gateway as exactly the same code and shape a REST-only client would ever see.
- **The same protos define Kafka event payloads.** One schema language, one set of generated types, one compatibility policy (Section 14).

### 12.3 Method Naming

Standard CRUD verbs regardless of service: `Get{Noun}`, `List{Noun}`, `Create{Noun}`, `Update{Noun}`, plus domain-specific action verbs where a CRUD verb doesn't fit (`SubmitClaim`, `RecordVerifierDecision`, `RetireCredits`, `SubmitMintTransaction`) — mirroring the "plural is default, named actions are the documented exception" philosophy from Section 2.1, expressed as method names.

### 12.4 Cross-Cutting Interceptors

Applied identically on every service's gRPC server:

- **mTLS enforcement** — connections without a valid peer certificate are rejected before reaching any handler.
- **Internal auth-token validation** — the short-lived signed token carrying caller identity, organization, roles, plan tier, verification status, and restriction state is validated and unpacked into the request context once, so every handler downstream can trust `ctx` already carries verified caller identity.
- **Tenant context establishment** — the organization ID from that token is what the repository layer uses for `SET LOCAL app.organization_id` (Section 11.2). A handler cannot substitute a different value.
- **Correlation ID propagation** — read from metadata (or generated if this is somehow the first hop) and attached to every log line, span, and outbound call.
- **Panic recovery** — a recovered panic becomes `INTERNAL_ERROR` rather than crashing the process or leaking a stack trace to a caller.
- **Structured logging and OpenTelemetry span** — every method call is logged and traced automatically, so handlers don't need to remember to instrument themselves.

### 12.5 Timeouts, Retries, Health

- Every gRPC client call sets an explicit deadline derived from the inbound request's remaining context deadline. A call never blocks indefinitely on a downstream service.
- Retries are applied automatically only to calls safe to retry — idempotent reads, `Get*` and `List*` — with bounded exponential backoff. **Mutating calls are never automatically retried by the gRPC client layer.** If a mutating call needs retry safety, that is handled explicitly at the caller's business-logic level using the mechanisms in Section 7, not silently retried underneath by transport plumbing. A silent transport retry on a non-idempotent call is exactly the kind of thing that causes a duplicate mint.
- Every service implements the standard gRPC Health Checking Protocol for readiness and liveness.

### 12.6 Cross-Service Call Inventory

The running master list; each domain LLD adds to it.

| Method | Server | Called by | Purpose |
|---|---|---|---|
| `IdentityService.GetOrganization` | Identity | API Gateway | Resolve org context once per request; cache-backed |
| `IdentityService.GetFacility` | Identity | Provenance, Sustainability | Facility detail and ownership check |
| `IdentityService.ValidateAPIKey` | Identity | API Gateway | Partner ingestion auth; result cached 60s at the Gateway |
| `IdentityService.GetTreasuryAddress` | Identity | Credit Ledger, Marketplace | Mint recipient and listing-authorization check |
| `BillingService.GetPlanSnapshot` | Billing | API Gateway | Plan tier, quotas, and rate limits, stamped into the internal token |
| `EvidenceService.CreateUploadTarget` | Evidence | Provenance, Sustainability | Issues a signed direct-upload target |
| `EvidenceService.GetEvidenceMetadata` | Evidence | Provenance, Sustainability, AI Agent, Fraud Detection | Scan status and content hash without fetching the file |
| `EvidenceService.GetEvidenceText` | Evidence | AI Agent | Extracted text for a page range, returned marked as untrusted |
| `SustainabilityService.ComputeCreditCeiling` | Sustainability | AI Agent | The authoritative formula, so the AI path and the issuance path can never diverge |
| `CreditLedgerService.CheckAvailableBalance` | Credit Ledger | Marketplace, Sustainability | Pre-flight plausibility check before a listing or issuance |
| `CreditLedgerService.GetFacilityCreditSummary` | Credit Ledger | Sustainability, Admin Backend | Read-side dashboard summary |
| `ChainWriterService.SubmitMintTransaction` | Chain Writer | Credit Ledger | The only path by which a mint is ever submitted |
| `ChainWriterService.SubmitAnchorTransaction` | Chain Writer | Chain Observer | Anchoring an epoch root |
| `ChainWriterService.SubmitRegisterBatchTransaction` | Chain Writer | Provenance | On-demand batch NFT minting |
| `ChainWriterService.SubmitFreezeTransaction` | Chain Writer | Admin Backend | Compliance freeze following escalation |
| `ChainWriterService.SubmitSetFeeTransaction` | Chain Writer | Billing | Per-seller fee override on a plan change |
| `ChainWriterService.SubmitPauseTransaction` | Chain Writer | Admin Backend | Emergency pause |
| `ChainWriterService.GetTransactionStatus` | Chain Writer | Credit Ledger, Marketplace, Admin Backend | Lifecycle state for a submitted transaction |
| `ChainObserverService.GetConfirmationStatus` | Chain Observer | Credit Ledger, Marketplace | Confirmation depth for a transaction hash when a caller needs it synchronously |
| `AIAgentService.GetReviewResult` | AI Agent | Sustainability, Admin Backend | Retrieve a completed assessment and its trace |

Note what is absent: no service calls Billing to check a quota on the request path, and no service calls Identity to re-resolve an organization it was already told about. Both are resolved once at the Gateway and carried in the internal token, which is what keeps the write path inside the two-hop budget in HLD Section 7.

---

## 13. Chain Transaction Lifecycle

Chain Writer Service owns a `transactions` table that is the durable record of every outbound on-chain operation. This exists because "submit a transaction" is not a request-response operation — it is a multi-minute state machine with several distinct failure modes, and treating it as a function call loses transactions.

| Column | Purpose |
|---|---|
| `id` | UUID |
| `signing_key` | Which of the three keys (`minter` / `anchor` / `ops`) |
| `nonce` | Allocated at queue time, `UNIQUE (signing_key, nonce)` |
| `operation_type` | `mint` / `anchor` / `register_batch` / `freeze` / `set_fee` / `pause` |
| `business_ref` | `claimId`, `epoch`, `listingId` — the correlation key to off-chain logs |
| `payload` | Encoded call data, including any authorization signatures |
| `state` | See below |
| `transaction_hash` | Populated at submission; changes on replacement |
| `gas_price_gwei`, `attempt_count` | Escalation tracking |
| `submitted_at`, `confirmed_at`, `block_number` | |
| `last_error` | For operator diagnosis |

**States:** `queued` → `signed` → `submitted` → `mined` → `confirmed`, with `failed` and `deferred` as terminal-ish branches.

| Situation | Handling |
|---|---|
| Nonce allocation | `SELECT ... FOR UPDATE` on the signing key's counter row, inside the same transaction that inserts the record. Allocation and recording are atomic, so no nonce is ever allocated and lost. |
| Stuck transaction | Unconfirmed after 3 minutes → resubmit the same nonce with gas price increased 25%, up to 5 escalations. Replacement rather than a new nonce, so the original cannot later confirm alongside it. |
| Nonce gap | A gap blocking later transactions is filled with a zero-value self-transfer at that nonce, after operator confirmation. Alerted, never silent. |
| Revert on rate limit | State `deferred`, re-queued with backoff, Organization notified of delayed issuance. A safety limit must never become a data-loss path. |
| Revert for any other reason | State `failed`, `chain.transaction.failed` published, surfaced to Admin Backend. Never auto-retried, since a revert usually means the request itself was wrong. |
| RPC provider unavailable | Circuit breaker opens; transactions accumulate in `queued` and drain when it recovers. Nothing is lost. |
| Reorg after `mined` | Detected by Chain Observer, state returns to `submitted`, re-tracked to confirmation depth. |

The invariant this table protects: **a mint that was authorized is never lost, and is never applied twice.** Every failure mode above resolves to one of those two outcomes.

---

## 14. Kafka Conventions

| Setting | Value |
|---|---|
| Partitions, default | 12 |
| Partitions, `checkpoint.logged` | 24 |
| Replication factor | 3 |
| `min.insync.replicas` | 2 |
| Producer `acks` | `all` |
| Producer `enable.idempotence` | `true` |
| Producer `max.in.flight.requests.per.connection` | 5 |
| Compression | `zstd` |
| Retention, standard topics | 7 days |
| Retention, financially significant topics | 30 days |
| Consumer `isolation.level` | `read_committed` |
| Consumer `enable.auto.commit` | `false` — offsets commit after the inbox transaction commits |

**Partition keys** are listed per topic in HLD Section 4.2 and are a correctness requirement rather than a tuning choice: Kafka guarantees ordering only within a partition, so a topic keyed wrongly produces out-of-order processing that is invisible under light load and corrupting under real load.

**Schemas are the same protobuf definitions used for gRPC** (Section 12.2), registered in a schema registry with a **backward-compatible** evolution policy: new fields must be optional, existing field numbers are never reused, and a field is deprecated for a full release before removal. A consumer running an older schema version must always be able to read a newer producer's message.

**Every event carries** `event_id` (UUID, the inbox deduplication key), `event_type`, `schema_version`, `occurred_at`, `organization_id`, and `request_id` in its headers.

**DLQ topics** are named `{topic}.dlq` and carry the original message plus failure metadata — the exception, the attempt count, and the consumer group that gave up. Replay is an explicit Admin Backend action that re-publishes to the source topic preserving the original `event_id`, which the inbox pattern makes safe.

---

## 15. Background & Scheduled Jobs

Kafka consumers are the primary background-processing mechanism system-wide. For work that isn't event-triggered, each service runs a lightweight in-process scheduler, and **every scheduled job acquires a distributed lock for its run window** before executing — preventing the same job from running redundantly, and in the case of anything touching the ledger dangerously, across every replica simultaneously.

| Job | Service | Cadence |
|---|---|---|
| On-chain ledger reconciliation | Credit Ledger | Every 15 minutes |
| Marketplace listing reconciliation | Marketplace | Every 15 minutes |
| Expired listing sweep | Marketplace | Hourly |
| Epoch anchor construction | Chain Observer | Every 10 minutes |
| Stuck transaction escalation | Chain Writer | Every minute |
| Usage counter reset | Billing | Monthly, per billing anchor |
| Trust Tier recomputation | Identity | Hourly, plus event-driven |
| Provenance Score recomputation backstop | Provenance Read | Every 6 hours |
| Idempotency and inbox record sweep | All | Daily |
| Evidence retention sweep | Evidence | Daily |
| Deletion request processing | Identity | Daily |
| Stale AI review promotion to human-only | Sustainability | Every 30 minutes |

Every job emits a Prometheus metric for last successful run, duration, and items processed, so a job that silently stops running is an alertable condition rather than something discovered when its output is missed.

---

## 16. Testing & Verification Standards

- **Unit tests** at the service layer against mocked repository and client interfaces — which is what the Dependency Inversion discipline in Section 3.1 exists to enable. Target 80% coverage on service-layer packages; handler and repository layers are covered by integration tests instead.
- **Integration tests** against a real Postgres and Redis in containers, per service, covering repository behaviour, RLS policy enforcement, migration up-and-down, and the idempotency and inbox flows including their concurrent cases.
- **Contract tests** verifying every gRPC server satisfies its proto contract and every consumer handles every schema version it claims to support.
- **RLS tests are mandatory and explicit**: for every tenant-scoped table, a test asserting that a query under organization A's context returns zero of organization B's rows. This is the kind of guarantee that is easy to assume and easy to lose silently in a refactor.
- **Load tests** against the HLD Section 3.2 capacity model, gating release on the Section 7.1 SLOs. Minimum scenarios: sustained checkpoint ingest at target rate, public read burst, marketplace browse under concurrent load, and bulk ingest of a full backlog.
- **Chaos scenarios** exercised at least once per release: Redis unavailable, one Kafka broker down, RPC provider unavailable, model provider circuit open, replica lag spike. Each has a defined expected degradation in HLD Section 8, and the test asserts that degradation rather than merely that nothing crashed.

---

*This document is the shared foundation for every domain-specific LLD that follows. Domain LLDs define their own tables, indexes, endpoints, and named lock keys, but inherit every rule above without restating it.*
