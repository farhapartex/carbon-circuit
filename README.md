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

What actually runs today, as distinct from what is designed. The design describes fifteen backend services, four smart contracts, and the full customer-facing product; the table below is the subset that is built.

### Frontend — `frontend/`

Next.js 16 (App Router), TypeScript in strict mode, Tailwind v4. **41 routes**, all rendering against typed fixtures — there is no live API yet, and every component is typed against view models that the real API will satisfy without the components changing.

| Area | State |
|---|---|
| Design system | **Done.** All 24 palette tokens, semantic type scale, shadcn/ui on Radix. Tailwind's default colour namespace is cleared so a stray `text-emerald-500` cannot bypass the palette. |
| Component tiers | **Done.** 14 primitives, 24 shared components, 7 layout shells, 6 feature domains. ESLint enforces the one-directional tier graph rather than leaving it to review. |
| Public surfaces | **Done.** Home, both solution pages, pricing, the QR-scan provenance page, and the public retirement log. |
| Onboarding | **Done.** Organization creation, registry verification (all three outcomes), plan selection, Treasury Address designation. |
| Settings | **Done.** Profile and MFA, organization details, members with invite and revoke, billing with usage and cancellation, API keys, Treasury Address. |
| Navigation | **Done.** Role-driven sidebar with capability gating for unverified and restricted organizations. |
| Wallet | **Partial.** wagmi, viem, and RainbowKit connect and switch networks. Proving an address by signature lands with the backend. |
| Remaining product surfaces | **Not started.** Batches, checkpoints, claims, the verifier workflow, credits, marketplace, dashboard, notifications. |
| Data layer | **Not started.** Auth0 session, the backend-for-frontend proxy, the typed API client, TanStack Query, and route protection all wait on the backend. |

### Backend — `backend/`

Go 1.26, single module. Gin at the edge only; everything behind it speaks gRPC, per the low-level design.

| Component | State |
|---|---|
| `api-gateway` | **Running.** The only public entry point. Gin REST surface, translates to gRPC, panic recovery, correlation IDs, standard response envelope. |
| `identity-service` | **Running.** gRPC only and not published to the host. Serves the gRPC health protocol driven by real database reachability. |
| Shared module | **Done.** Config loading that fails loudly at startup, structured logging, the full 22-code error taxonomy, gRPC server and interceptors, GORM over PgBouncer. |
| Local infrastructure | **Running.** Postgres, PgBouncer in transaction pooling mode, Redis, Kafka in KRaft mode. One `make up`. |
| Database schema | **Not started.** No models or migrations yet — the next piece of work. |
| Remaining 13 services | **Not started.** Billing, provenance, evidence, sustainability, credits, marketplace, chain writer and observer, AI agent, fraud detection, notification, admin. |

Two properties worth calling out, because both are verified rather than assumed. Each service holds its **own Postgres role and its own PgBouncer pool**, so "no service reaches into another service's tables" is enforced by the database permission system — `identity_service` is denied on the `billing` schema and vice versa. And a **correlation ID set at the gateway reaches the identity service log** across the gRPC hop, which is the mechanism the whole tracing story depends on.

### Smart contracts — `smart-contract/`

**Not started.** The design is complete in [Smart Contract Design](./docs/CarbonCircuit_SmartContract_Design.md). Credit Ledger, Marketplace, Chain Writer, and Chain Observer all depend on the contracts being deployable to a local Anvil node, which puts the Solidity work on the critical path for those services.

### Known gaps in what is built

Stated plainly rather than discovered later:

- **gRPC between services is plaintext, not mTLS.** The design requires mutual TLS. Traffic is currently confined to one Docker network, and the transport credentials are a single call site in each of the client and server.
- **No authentication yet.** Auth0 is configured but not wired. Every surface renders fixture data.
- **`identity-service` has no container healthcheck.** Its gRPC health service is registered and reports real database reachability, but the distroless image carries no shell or probe binary, so Compose cannot exercise it directly. The gateway's `/readyz` does.

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
