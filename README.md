# CarbonCircuit

**A multi-industry supply-chain provenance and carbon credit platform — launching with electronics components.**

CarbonCircuit does two connected things for physical supply chains. It records the verifiable journey of a batch of goods from raw material through manufacturing, assembly, and shipment. And it lets the facilities in that chain have their sustainability practices independently reviewed and converted into carbon credits that other companies can buy, hold, and permanently retire against their own emissions.

Provenance answers *"where did this come from, and is it authentic?"* Carbon credits answer *"how much verified environmental benefit did this facility create, and who now owns it?"*

> **Status — in development.** The design is complete and documented: product requirements, backend architecture, low-level conventions, smart contract design, and frontend architecture. Implementation is under way — see [Current stage](#current-stage) for exactly what runs today. Every document here is written to be built against directly, and where code and a document disagree, that is a defect in one of them.

---

## Why I'm building this

This is a personal project to build a genuinely production-shaped Web3 system rather than a demo — combining smart contracts, event-driven backend services, and an AI agent embedded in a real workflow instead of a chatbot bolted onto the side.

The problem is real. Electronics supply chains pass through 5–8 facilities across multiple countries with no verifiable way to confirm origin claims. Sustainability claims are usually self-reported with no independent check. And carbon markets globally have a credibility problem driven by credits that could not be traced or were counted twice.

So the system is designed around three properties that are easy to claim and hard to actually deliver:

- **A credit is issued exactly once**, against exactly one approved claim, enforced on-chain rather than by a database row someone could edit.
- **A credit never loses its identity.** Every credit permanently carries the facility that earned it, the year, and the activity type — because a buyer's entire reason for paying is being able to name that in their own ESG report.
- **A retired credit is gone forever.** No resale, no re-transfer, no second retirement.

The design is also explicit about what it *doesn't* guarantee — it cannot detect that the same real-world reduction was also registered with Verra or Gold Standard. That boundary is documented, disclosed to buyers in the product, and mitigated rather than hidden.

---

## Current stage

What actually runs today, as distinct from what is designed. The design describes fifteen backend services, four smart contracts, and the full customer-facing product; the sections below are the subset that is built.

Login and onboarding work end to end against real services. Everything past onboarding is still interface without a backend behind it.

### Frontend — `frontend/`

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

### Backend — `backend/`

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

### Smart contracts — `smart-contract/`

**Not started.** The design is complete in [Smart Contract Design](./docs/CarbonCircuit_SmartContract_Design.md). Credit Ledger, Marketplace, Chain Writer, and Chain Observer all depend on the contracts being deployable to a local Anvil node, which puts the Solidity work on the critical path for those services.

### Known gaps in what is built

Stated plainly rather than discovered later:

- **The registry-verification screen shows a fixture, not your result.** The outcome is computed correctly and stored — `verified`, `unverified`, or `rejected` with its reason — but the screen that should display it was built before the endpoint existed.
- **Treasury Address designation has no endpoint.** The onboarding step renders and connects a wallet, but nothing proves ownership by signature or persists the address, so `is_treasury_designated` is always false.
- **Stripe is not wired.** Selecting a paid plan creates an active subscription as though payment had succeeded. Deliberate for now; the subscription state machine and grace periods exist in the schema.
- **Caller context invalidation has no callers.** The cache is invalidated explicitly on demand, but no endpoint changes a role or revokes a membership yet, so nothing calls it. When member management arrives it must, or a demoted user keeps their old role for up to sixty seconds.
- **`plan_tier` in the internal token is always empty.** The claim exists because the design lists it, but populating it would mean a billing call during context resolution, which would reintroduce the hop the cache removes. Nothing consumes it yet.
- **Development certificates and signing keys are generated locally.** `docker compose run --rm devcerts` issues a dev CA and per-service certificates into a gitignored volume; `go run ./cmd/devkeys` mints the token keypair. Production needs real certificate issuance and a secrets manager.
- **`identity-service` has no container healthcheck.** Its gRPC health service is registered and reports real database reachability, but the distroless image carries no shell or probe binary, so Compose cannot exercise it directly. The gateway's `/readyz` does.
- **The business registry dataset has 100 rows, not the ~500 the PRD specifies.** Every verification branch is reachable, but the count is a documented divergence.

---

## Architecture

### System design — request and event flow

![CarbonCircuit system design: request and event flow](./System-design.png)

Fifteen services across eight domains. Requests enter through a single API Gateway that resolves caller context **once** and stamps it into an internal service token, so no write path needs more than two synchronous hops. Everything that isn't a question-and-answer runs through Kafka with a transactional outbox on produce and a transactional inbox on consume. The public QR-scan page is served from a pre-computed read model behind a CDN and never touches the write side or the chain.

One boundary is drawn harder than any other: **Chain Writer Service is the only component in the system with access to signing keys**, and even it cannot authorize anything — a mint carries EIP-712 verifier signatures that the contract verifies for itself.

### Smart contract design

![CarbonCircuit smart contract design on Base L2](./Smart-Contract-Design.png)

Four upgradeable contracts on Base, each deployed as a proxy holding all storage plus a freely replaceable implementation holding all logic. Credits are **ERC-1155**, one token ID per Credit Class, so attribution is a property of the token rather than bookkeeping alongside it. Provenance is anchored as **one Merkle root per 10-minute epoch platform-wide**, so on-chain cost doesn't scale with customer count and every historical root stays permanently verifiable.

Governance sits behind a 3-of-5 multisig routed through a timelock, with a separate hot guardian key that can pause but never unpause — and whose pause auto-expires after 72 hours so a compromised guardian causes a bounded outage rather than an indefinite one.

---

## Documentation

Read in this order. Each document assumes the ones above it.

| # | Document | What it covers |
|---|---|---|
| 1 | **[Product Requirements](./docs/CarbonCircuit_PRD.md)** | What the product is and why. Users, credit formulas and reference factor tables, Credit Classes, Provenance Score and Trust Tier formulas, registry verification, the marketplace, fraud rules, plans and pricing, multi-tenancy, deletion and export. |
| 2 | **[Backend High-Level Design](./docs/CarbonCircuit_Backend_HLD.md)** | Service inventory and boundaries, target chain and capacity model, synchronous vs. event communication, the security architecture, rate limits, latency SLOs, resilience, the four-layer idempotency strategy, and the full AI Agent Service design. |
| 3 | **[LLD — Global Conventions](./docs/CarbonCircuit_LLD.md)** | The cross-cutting standards every domain LLD inherits: API envelope and error taxonomy, layered service structure, database and indexing conventions, idempotency and inbox implementation, locking, the shared-Postgres architecture with RLS and PgBouncer, gRPC conventions, the chain transaction lifecycle, Kafka settings, and testing standards. |
| 4 | **[Smart Contract Design](./docs/CarbonCircuit_SmartContract_Design.md)** | Contract inventory, the Credit Class token model, interfaces, libraries, the role matrix, per-contract design detail, security architecture, the proxy and storage-layout discipline, cross-system flows, and the test and deployment strategy. |
| 5 | **[Frontend Architecture](./docs/CarbonCircuit_Frontend_Architecture.md)** | Component tiers, the server boundary and hydration strategy, state ownership, the accessible design system, security posture, wallet integration and authorization states, the full sitemap, and a page-by-page component breakdown. |

---

## Tech stack

| Layer | Stack |
|---|---|
| **Backend** | Go · Gin (API Gateway only) · gRPC + Protocol Buffers over mTLS · GORM |
| **Data** | PostgreSQL — schema and DB role per service, Row-Level Security · PgBouncer · Redis · pgvector |
| **Events** | Apache Kafka — transactional outbox on produce, transactional inbox on consume |
| **AI review** | Python · LangGraph · DSPy · Anthropic Claude (`claude-opus-5`) |
| **Smart contracts** | Solidity · Foundry · OpenZeppelin Upgradeable (UUPS) · ERC-1155 · EIP-712 |
| **Chain** | Base L2 · USDC settlement · Cloud KMS transaction signing |
| **Frontend** | Next.js (App Router) · TypeScript · Tailwind CSS · shadcn/ui |
| **Frontend state** | TanStack Query for server state · Zustand for client state |
| **Wallet** | wagmi · viem · RainbowKit · SIWE |
| **Platform** | Auth0 · Stripe · HashiCorp Vault |
| **Observability** | Prometheus + Grafana · Loki · OpenTelemetry + Tempo |
| **Tooling** | Docker · Docker Compose · k8s |
