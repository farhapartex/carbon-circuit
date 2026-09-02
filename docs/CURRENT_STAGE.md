# Current stage

[← Back to the README](../README.md)

What actually runs today, as distinct from what is designed. The design describes fifteen backend services, four smart contracts, and the full customer-facing product; the sections below are the subset that is built.

Login and onboarding work end to end against real services. Everything past onboarding is still interface without a backend behind it.

## Frontend — `frontend/`

Next.js 16 (App Router), TypeScript in strict mode, Tailwind v4. **41 routes**. Five call live APIs; eleven still render typed fixtures; the rest are page shells awaiting their service.

| Area | State |
|---|---|
| Design system | **Done.** All 24 palette tokens, semantic type scale, shadcn/ui on Radix. Tailwind's default colour namespace is cleared so a stray `text-emerald-500` cannot bypass the palette. |
| Component tiers | **Done.** 15 primitives, 25 shared components, 7 layout shells, 6 feature domains. ESLint enforces the one-directional tier graph rather than leaving it to review. |
| Authentication | **Live.** Auth0 authorization-code flow with PKCE through a backend-for-frontend. The access token never reaches the browser — it lives in an encrypted session cookie, and `import "server-only"` makes a client component that imports the API client fail the build. |
| Route protection | **Live.** `proxy.ts` gates every app route on onboarding progress read from the session cookie, so an un-onboarded user cannot reach the dashboard by typing the URL. Public surfaces stay public. |
| Onboarding | **Partial.** Organization creation and plan selection write to real services and drive the gate. The registry-verification screen still renders a fixture rather than the outcome actually computed, and Treasury Address designation has no endpoint. |
| Pricing | **Live.** Plans come from Postgres through the gateway, filtered server-side to the organization types eligible for them. |
| App shell | **Live.** Sidebar, account menu, and verification banner read the signed-in user and organization from the session. |
| Public surfaces | **Done.** Home, both solution pages, pricing, the QR-scan provenance page, and the public retirement log — the last two on fixtures. |
| Settings | **UI only.** Profile and MFA, organization details, members with invite and revoke, billing with usage and cancellation, API keys, Treasury Address — all built, all on fixtures. |
| Remaining product surfaces | **Page shells only.** Batches, checkpoints, claims, the verifier workflow, credits, marketplace, dashboard, notifications render a heading and nothing else. |
| Wallet | **Partial.** wagmi, viem, and RainbowKit connect and switch networks. Proving an address by signature lands with the backend. |
| Server state | **Not started.** TanStack Query. Server Components and Server Actions cover what exists so far. |

## Backend — `backend/`

Go 1.26, single module. Gin at the edge only; everything behind it speaks gRPC over mutual TLS, per the low-level design.

| Component | State |
|---|---|
| `api-gateway` | **Running.** The only public entry point. Verifies Auth0 JWTs against cached JWKS, enforces per-caller rate limits, requires an `Idempotency-Key` on every mutation, resolves caller context once and stamps it into a signed internal token. Seven routes. |
| `identity-service` | **Running.** Provisions users from verified Auth0 claims, creates organizations with registry verification, resolves session context. gRPC only, never published to the host. |
| `billing-service` | **Running.** Serves the plan catalogue with a 24-hour Redis read-through cache, and creates subscriptions. gRPC only. |
| Database schema | **Done for two services.** 22 migrations across `identity` and `billing` — 31 tables, Row-Level Security with `FORCE` on every tenant-scoped table, run through golang-migrate as a Compose job rather than by hand. |
| Idempotency | **Done.** The four-layer scheme from the design: reservation, request-hash comparison, byte-for-byte response replay, and reuse detection — all inside the same transaction as the business write. |
| Transactional outbox | **Done.** Writer bound to a transaction, publisher draining to Kafka with `FOR UPDATE SKIP LOCKED` so replicas never double-publish, keyed by aggregate for per-partition ordering. |
| Service-to-service security | **Done.** Mutual TLS on every gRPC hop plus a short-lived Ed25519 token that carries caller context. Only the gateway holds a signing key, so a compromised service can verify tokens but never mint one. |
| Local infrastructure | **Running.** Postgres, PgBouncer in transaction pooling mode, Redis, Kafka in KRaft mode. One `make up`. |
| Remaining 12 services | **Not started.** Provenance and its read side, evidence, sustainability, credits, marketplace, chain writer and observer, AI agent, fraud detection, notification, admin. |

Four properties worth calling out, because each is verified rather than assumed. Every service holds its **own Postgres role and PgBouncer pool**, so "no service reaches into another service's tables" is enforced by database permissions — `identity_service` is denied on the `billing` schema and vice versa. **Row-Level Security is proven, not presumed**: with only a user context set, a user reads their own memberships and *zero* organizations. A **plaintext gRPC connection to either service is refused**, and a call carrying a valid certificate but no service token is rejected with `Unauthenticated`. And **organization identity cannot be asserted by a caller** — the fields that once carried it are `reserved` in the protos, so it comes from the verified token or not at all.

Twelve test files, the security-critical ones checked by removing the guard and confirming the test fails.

## Smart contracts — `smart-contract/`

**Not started.** The design is complete in [Smart Contract Design](./CarbonCircuit_SmartContract_Design.md). Credit Ledger, Marketplace, Chain Writer, and Chain Observer all depend on the contracts being deployable to a local Anvil node, which puts the Solidity work on the critical path for those services.

## Known gaps in what is built

Stated plainly rather than discovered later:

- **The registry-verification screen shows a fixture, not your result.** The outcome is computed correctly and stored — `verified`, `unverified`, or `rejected` with its reason — but the screen that should display it was built before the endpoint existed.
- **Treasury Address designation has no endpoint.** The onboarding step renders and connects a wallet, but nothing proves ownership by signature or persists the address, so `is_treasury_designated` is always false.
- **Stripe is not wired.** Selecting a paid plan creates an active subscription as though payment had succeeded. Deliberate for now; the subscription state machine and grace periods exist in the schema.
- **Caller context invalidation has no callers.** The cache is invalidated explicitly on demand, but no endpoint changes a role or revokes a membership yet, so nothing calls it. When member management arrives it must, or a demoted user keeps their old role for up to sixty seconds.
- **`plan_tier` in the internal token is always empty.** The claim exists because the design lists it, but populating it would mean a billing call during context resolution, which would reintroduce the hop the cache removes. Nothing consumes it yet.
- **Development certificates and signing keys are generated locally.** `docker compose run --rm devcerts` issues a dev CA and per-service certificates into a gitignored volume; `go run ./cmd/devkeys` mints the token keypair. Production needs real certificate issuance and a secrets manager.
- **`identity-service` has no container healthcheck.** Its gRPC health service is registered and reports real database reachability, but the distroless image carries no shell or probe binary, so Compose cannot exercise it directly. The gateway's `/readyz` does.
- **The business registry dataset has 100 rows, not the ~500 the PRD specifies.** Every verification branch is reachable, but the count is a documented divergence.
