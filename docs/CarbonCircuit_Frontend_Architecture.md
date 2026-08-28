# CarbonCircuit — Frontend Architecture & Design

**Scope:** Next.js, TypeScript, Tailwind CSS. Covers every customer-facing surface in the PRD (Manufacturer, Assembler, Logistics Partner, Verifier, Credit Buyer, and the public end-consumer QR page). The Admin/Ops Portal is a separate application built on a different internal Go framework and is out of scope here.

This document defines pages and routes, the component inventory per page, global state architecture, the design system, security posture, and wallet/contract integration. It is system design — no code — precise enough to build against directly.

---

## 1. Core Technical Decisions

| Decision | Choice | Why |
|---|---|---|
| Framework | Next.js, App Router | Server Components for data-heavy authenticated pages reduce client bundle size and simplify data fetching; file-based routing maps cleanly onto the sitemap below; ISR gives the public provenance page a CDN-cacheable render for free |
| Language | TypeScript, strict mode | Type safety across API contracts, especially important given the financial data flowing through the UI |
| Styling | Tailwind CSS | Utility-first, pairs naturally with a component-driven architecture and a constrained design token system (Section 5) |
| Component foundation | shadcn/ui (Radix primitives + Tailwind) | Every primitive ships as an owned, editable component file rather than an opaque dependency — which is what makes "everything must be a component" enforceable in practice, since there's no pressure to reach for inline markup when an accessible, styled primitive already exists |
| Icons | lucide-react | Consistent, modern, tree-shakeable |
| Charts | Recharts | Used sparingly, only where the data genuinely benefits from visualization |
| Global client state | Zustand | Lightweight, no provider ceremony; used strictly for client-only state (Section 4) |
| Server state | TanStack Query | Owns all data fetched from the backend — caching, background refetch, request deduplication, mutation invalidation. Server data never lives in Zustand. |
| Auth | Auth0 Next.js SDK | Matches the backend's Auth0 integration; session in secure HTTP-only cookies, never in client-side state, and the access token never reaches the browser (Section 3) |
| Wallet connection | wagmi + viem + RainbowKit | Standard modern EVM stack for wallet connection, transaction signing, and chain management |
| Forms | React Hook Form + Zod | Schema-driven validation, matching the backend's validate-at-the-edge philosophy |
| Decimal handling | dnum | Credit and USDC amounts are handled as exact decimals end to end. JavaScript numbers cannot represent a 6-decimal financial amount reliably, and a rounding artifact in a purchase preview is a trust failure, not a display bug. |

---

## 2. Component Architecture Rule

Every route's `page.tsx` is a thin composition layer: it fetches or receives data and renders a tree of components — never raw markup blocks beyond composing its children. This is enforced structurally through three component tiers:

| Tier | Location | Contains |
|---|---|---|
| **Primitives** | `components/ui/` | Design-system atoms: Button, Input, Select, Card, Badge, Dialog, Table, Tabs, Tooltip, Avatar, Skeleton, Toast — one component per file, shadcn/ui-based |
| **Shared feature components** | `components/shared/` | Cross-domain building blocks: `WalletConnectButton`, `AddressDisplay`, `CreditAmountDisplay`, `UsdcAmountDisplay`, `StatusPill`, `EmptyState`, `PageHeader`, `DataTable`, `FileDropzone`, `Timeline`, `ConfirmDialog`, `MetricCard`, `CopyButton` |
| **Domain feature components** | `components/features/{domain}/` | Domain-specific composed components (`features/batches/BatchCard.tsx`, `features/claims/ClaimStatusStepper.tsx`) — the only place domain-specific display logic lives, such as which badge variant maps to which claim status |

**Layout components** (`components/layout/`) — `AppShell`, `Sidebar`, `Topbar`, `MarketingShell`, `PublicLayout`, `AuthLayout` — are themselves composed from primitives and shared components, following the same rule.

No tier reaches upward: a primitive never imports a domain component, a shared component never imports a domain-specific one. This keeps the dependency graph one-directional and every component independently testable and reusable.

**Two shared components are worth calling out as deliberately narrow.** `CreditAmountDisplay` and `UsdcAmountDisplay` are the only components permitted to render a financial figure anywhere in the application. They take an exact decimal string from the API, never a JavaScript number, and they own the formatting, the unit suffix, and the precision. Every financial figure rendering through one component means a formatting or rounding fix happens once rather than in forty places, and it makes "did we ever render a float here" a question with a definitive answer.

---

## 3. Data Fetching, Auth, and the Server Boundary

This section comes early because it is the decision the rest of the application is built on top of, and the one most likely to be got wrong in an App Router codebase.

### 3.1 The Access Token Never Reaches the Browser

The Auth0 session lives in an HTTP-only, `SameSite=Lax`, `Secure` cookie that client-side JavaScript cannot read. The browser never holds an access token for the CarbonCircuit API.

Every request to the backend therefore goes through the Next.js server, which acts as a **backend-for-frontend**:

- **Server Components** call the API directly from the server, attaching the access token retrieved from the session.
- **Client Components** call Next.js Route Handlers under `/api/proxy/*`, which validate the session, attach the access token, forward to the API Gateway, and return the response unchanged.

The proxy is a pass-through by design — it adds credentials and nothing else. It does not reshape responses, does not implement business logic, and does not become a second place where API semantics live. It also forwards the `Idempotency-Key` header from the client rather than generating one, so a client retry genuinely retries the same logical operation rather than becoming a new one.

An XSS vulnerability in the frontend is contained by this: the attacker can make requests as the user through the proxy, but cannot exfiltrate a bearer token to use elsewhere, and cannot use it after the session cookie is invalidated.

### 3.2 Server Prefetch, Client Hydration

Authenticated pages prefetch on the server and hydrate on the client:

1. The Server Component creates a `QueryClient`, prefetches the queries the page needs, and renders the tree inside a `HydrationBoundary` with the dehydrated state.
2. Client Components use `useQuery` with the identical query key and receive the prefetched data instantly, with no loading state and no second request.
3. Background refetch, invalidation on mutation, and polling then work normally on the client.

Without this, a Server Component fetch and a Client Component `useQuery` become two separate requests for the same data — the page renders, then immediately flashes a loading state and refetches everything. Query keys are therefore defined in one shared module and imported by both sides, since a key mismatch produces exactly that bug silently.

### 3.3 Query Conventions

| Concern | Convention |
|---|---|
| Key structure | `['domain', 'resource', params]` — e.g. `['claims', 'list', { status, page }]` |
| Default `staleTime` | 30 seconds |
| Financial reads (balances, listings, quotes) | `staleTime: 0`, refetch on window focus |
| Pending on-chain transactions | Polled every 5 seconds until resolved, then invalidated |
| Mutation success | Explicit invalidation of affected keys, never a blanket cache clear |
| Error handling | `error.code` from the standard envelope drives the UI; `request_id` is always surfaced in the error state for support |
| Retry | Automatic on `DEPENDENCY_UNAVAILABLE` and network errors only; never on a 4xx |

---

## 4. State Management Architecture

State is split strictly by ownership, and this split is a hard rule rather than a preference — mixing the two responsibilities is exactly what causes stale-cache bugs.

### 4.1 Server State — TanStack Query

Everything originating from the backend — batches, claims, credit balances, marketplace listings, notifications, organization and facility data, plan and usage data — is fetched, cached, and mutated exclusively through TanStack Query.

### 4.2 Wallet State — wagmi Directly

Wallet connection status, connected address, and chain ID are read from wagmi's own hooks at the point of use. They are **not** mirrored into a global store.

Duplicating reactive external state into a second store creates two sources of truth that drift, and drifted wallet state is not a cosmetic problem — it is a user signing a transaction from an address the UI is no longer showing. Components that need wallet state call `useAccount` and `useChainId`. A thin `useWalletContext` hook composes wagmi state with the Organization's Treasury Address from the API to answer the one question the application actually asks repeatedly: *is the connected wallet authorized to act for this Organization?*

### 4.3 Client State — Zustand

Reserved strictly for state with no server-side source of truth and no external reactive source:

- **UI context store** — active Organization for users belonging to several, sidebar collapsed state, active table density
- **Toast queue store** — transient UI notifications, distinct from the persisted backend-driven notification centre
- **Multi-step form draft store** — in-progress, not-yet-submitted form state for batch creation and claim submission, so a user can navigate away and back without losing progress. Persisted to `sessionStorage` and cleared on successful submit. Uploaded evidence is referenced by the ID returned from the upload, never held as file content in the store.

### 4.4 What Never Goes in Zustand

Auth tokens (there are none client-side), wallet state (Section 4.2), and any data that has a backend record. If it can go stale relative to the server, it belongs in TanStack Query with an invalidation strategy.

---

## 5. Design System

### 5.1 Color Palette

A light palette built around a teal/emerald primary — a sustainability association without the literal recycling-green cliché — neutral warm grays for structure, and semantic colors for status. All backgrounds are light; there is no dark theme in this version.

The palette carries **two weights per semantic color for two different jobs**, and using the wrong one is the most common way a palette like this fails accessibility:

| Token | Hex | Contrast on `neutral-50` | Permitted use |
|---|---|---|---|
| `primary-50` | `#ECFDF5` | — | Tinted backgrounds: selected nav item, info banners |
| `primary-500` | `#10B981` | 2.43 | Decorative fills and illustrations only — never text, never a meaning-bearing indicator |
| `primary-600` | `#059669` | 3.61 | Borders, focus rings, status dots, icons |
| `primary-700` | `#047857` | 5.25 | **Text, links, and primary button fill** (white text on it reaches 5.48) |
| `primary-800` | `#065F46` | 7.36 | Button hover/pressed, text on `primary-50` |
| `neutral-50` | `#FAFAF9` | — | Page background |
| `neutral-100` | `#F5F5F4` | — | Subtle section background |
| `neutral-200` | `#E7E5E4` | — | Borders, dividers |
| `neutral-400` | `#A8A29E` | 2.41 | Decorative dividers only |
| `neutral-600` | `#57534E` | 7.30 | Secondary and muted text |
| `neutral-900` | `#1C1917` | 16.74 | Primary text |
| `success-50` | `#F0FDF4` | — | Tinted background |
| `success-600` | `#16A34A` | 3.16 | Icons, dots, borders |
| `success-700` | `#15803D` | 4.80 | Success text |
| `warning-50` | `#FFFBEB` | — | Tinted background |
| `warning-600` | `#D97706` | 3.05 | Icons, dots, borders |
| `warning-700` | `#B45309` | 4.81 | Warning text |
| `danger-50` | `#FEF2F2` | — | Tinted background |
| `danger-600` | `#DC2626` | 4.62 | Icons, dots, destructive button fill (white text reaches 4.83) |
| `danger-700` | `#B91C1C` | 6.19 | Error text, destructive button hover |
| `info-50` | `#EFF6FF` | — | Tinted background |
| `info-600` | `#2563EB` | 4.95 | Icons, dots, borders |
| `info-700` | `#1D4ED8` | 6.42 | Informational text, secondary links |

**The rule that follows from the table: 700-weights carry text, 600-weights carry non-text indicators, 500-weights are decorative.** A status pill uses a `-50` background with `-700` text and a `-600` dot; every one of those pairings clears its threshold. A green `-500` link on a white page is 2.43:1 and unreadable for a substantial number of users, which is why `primary-500` appears in the table with an explicit prohibition rather than being left as a judgment call.

**Status mappings**, applied consistently everywhere a status appears via `StatusPill`:

| Domain | Mapping |
|---|---|
| Claim status | Draft → `neutral` · Submitted / Under AI Review / Under Human Review → `warning` · Approved → `success` · Rejected → `danger` · More Info Requested → `info` |
| Listing status | Active → `success` · Partially filled → `info` · Filled → `neutral` · Cancelled / Expired → `neutral` |
| Transaction status | Awaiting signature → `info` · Pending → `warning` · Confirmed → `success` · Failed → `danger` |
| Trust Tier | New → `neutral` · Verified → `info` · Trusted → `primary` |
| Verification status | Verified → `success` · Unverified → `warning` · Rejected → `danger` |
| Provenance band | Complete → `success` · Strong → `primary` · Partial → `warning` · Limited → `neutral` |

Trust Tier deliberately uses a distinct three-step scale so it is never visually confused with claim status, and Provenance band uses `neutral` rather than `danger` for its lowest band — a batch with few checkpoints is incomplete, not fraudulent, and coloring it red would misrepresent what the score means.

### 5.2 Design Principles Applied

- **Generous whitespace** — the default 4px-based spacing scale, no cramped layouts
- **Soft elevation, not heavy borders** — cards use a `neutral-200` border plus a soft `shadow-sm` rather than harsh dark borders
- **Consistent corner radius** — `rounded-lg` (8px) as the default across cards, buttons, inputs, badges; `rounded-full` reserved for pills, avatars, and status dots
- **Typography hierarchy** — Inter with a constrained scale (page title, section heading, body, caption), never more than two weights on a single screen. Financial figures and addresses use a tabular-numeral variant so columns align and digits don't shift on update.
- **Micro-interactions** — 150–200ms transitions on hover and focus, skeleton loaders matching content shape rather than spinners
- **Focus visibility** — a 2px `primary-600` focus ring with a 2px offset on every interactive element, never removed. Keyboard navigation is a supported way to use this application, not an afterthought.
- **Motion respect** — every transition is disabled under `prefers-reduced-motion`

### 5.3 Accessibility as a Verified Property

Every palette pairing in Section 5.1 is asserted by an automated contrast test in CI against the WCAG AA thresholds, and the ratios in that table are the test's expected values. This is stated as a mechanism rather than an intention because a palette claim that isn't tested drifts the first time someone adds a token.

Beyond contrast: every interactive primitive inherits Radix's keyboard navigation and ARIA behaviour through shadcn/ui, so accessibility is a property of the primitive tier rather than something bolted on per page. Automated axe checks run against every route in CI, and every form field has a programmatically associated label — a placeholder is never a label.

---

## 6. Security Posture

### 6.1 Content Security Policy

A strict CSP is served on every response, with a per-request nonce for inline scripts. Wallet connectivity requires explicit allowances — `connect-src` for the RPC endpoint and WalletConnect relay, `img-src` for wallet icon assets — and those allowances are enumerated explicitly rather than opened with a wildcard, because a wallet integration is exactly the kind of feature that quietly turns into `script-src *` if nobody writes down what it actually needs.

`frame-ancestors 'none'` prevents the application being framed, which matters more than usual here: a framed wallet-signing flow is the standard shape of a transaction-approval phishing attack.

### 6.2 Other Headers and Handling

`Strict-Transport-Security` with preload, `X-Content-Type-Options: nosniff`, `Referrer-Policy: strict-origin-when-cross-origin`, and a `Permissions-Policy` denying camera, microphone, and geolocation except on the checkpoint-logging route, which requests geolocation explicitly.

**No user-supplied content is ever rendered as HTML.** Evidence documents are displayed through a PDF viewer or as images, never injected into the DOM; rejection reasons, retirement purposes, and facility descriptions render as plain text. Evidence and metadata are fetched through the Next.js server rather than by the browser directly, so no customer-controlled URL is ever dereferenced by the client and object-storage URLs are never exposed.

---

## 7. Wallet & Contract Integration

### 7.1 Connection and Authorization

`WalletConnectButton` (always present in the Topbar) opens a RainbowKit modal supporting MetaMask, WalletConnect, and Coinbase Wallet. Once connected, the address is proven through a **Sign-In With Ethereum** signature against a server-issued, single-use nonce bound to the application's domain — the nonce is requested from the backend, consumed on verification, and cannot be replayed.

**Connecting a wallet is not the same as being authorized to act for the Organization.** Credits are Organization assets held at the Organization's Treasury Address (PRD Section 3.8), and an individual employee's personal wallet is never their custodian. The UI distinguishes three states clearly, because a user who does not understand which one they are in will send a transaction that reverts and will not know why:

| State | What the UI shows |
|---|---|
| No wallet connected | Connect prompt on any action requiring a signature |
| Connected, but not the Treasury Address or an approved operator | An explanation naming the Treasury Address and what is needed, with the action disabled — never a generic failure |
| Connected and authorized | The action is available |

### 7.2 What Requires a Connected Wallet

| Action | Wallet required? | Notes |
|---|---|---|
| Browsing marketplace listings | No | Read-only, works unauthenticated |
| Viewing credit balance | No | Read from the backend's off-chain mirror, not a live chain read |
| Buying credits | Yes | User signs the purchase transaction directly |
| Listing credits for sale | Yes | Escrow transfer signed by the Treasury Address or an approved operator |
| Retiring credits | Yes | Burn signed by the holder |
| Submitting a batch or sustainability claim | No | Backend-orchestrated; the submitting user never signs |
| Setting the Treasury Address | Yes | A signature from the new address proves ownership |

### 7.3 Transaction Feedback and Recovery

Every wallet-signed action follows a consistent state sequence: **Awaiting signature** → **Pending** (with a block-explorer link) → **Confirmed** or **Failed**, driven by wagmi's transaction-watching hooks rather than manual polling.

**A transaction the browser stops watching still resolves.** Before opening the wallet, the intent is registered with the backend; once a hash exists it is sent to the backend, which tracks the transaction through Chain Observer Service independently of the browser. If the user closes the tab mid-flow, the transaction still completes, the backend still records it, and the user still receives their notification. On returning, a `PendingTransactionBanner` shows any operations still in flight for their Organization.

This matters because the alternative failure is genuinely bad: a user buys credits, closes the tab before confirmation, and finds nothing in the UI acknowledging that their USDC left their wallet. Making the browser session non-load-bearing is the difference between a transient network problem and an apparent loss of funds.

**Pre-flight checks** run before the wallet is opened, since every one of these produces a confusing wallet-level revert if left to the chain: wrong network, insufficient USDC, insufficient allowance, listing already filled or expired, quantity below minimum, remainder rule violated, account frozen. `PurchaseForm` fetches a fresh `quote` from the contract immediately before signing and shows the exact cost, fee, and total, so the figures the user approves are the figures that settle.

### 7.4 Network Handling

`NetworkSwitcher` detects a wrong-chain connection and prompts a one-click switch before any signed action is possible. The expected chain comes from configuration, never hardcoded, so the same build runs against a local node, testnet, and production.

---

## 8. Sitemap & Routing

### 8.1 Public Routes (No Authentication)

| Route | Rendering | Purpose |
|---|---|---|
| `/` | Static | Home landing page — introduces both product pillars with a persona-based split |
| `/solutions/traceability` | Static | Landing page for Manufacturers, Assemblers, Logistics Partners |
| `/solutions/carbon-credits` | Static | Landing page for Credit Buyers and sustainability teams |
| `/pricing` | ISR, 5 min | Public plan comparison, rendered from live backend plan data |
| `/track/[publicRef]` | ISR, 60s + CDN | Public provenance timeline — the QR-scan destination |
| `/marketplace/retirements` | ISR, 60s | Public retirement transparency log |
| `/login` | Dynamic | Auth0 login redirect |
| `/signup` | Dynamic | Organization registration and onboarding |

`/track/[publicRef]` is the one route with a genuinely unpredictable traffic profile, and it is built for that: it renders statically with a 60-second revalidation window and is served from the CDN with `stale-while-revalidate`, so a batch going viral is absorbed at the edge and never reaches origin at scale. The route parameter is the **public batch reference** from the backend — a random opaque identifier — never an internal ID, so the URL space cannot be enumerated to map platform activity.

### 8.2 Authenticated Routes

| Route | Purpose |
|---|---|
| `/dashboard` | Role-aware overview |
| `/facilities`, `/facilities/new`, `/facilities/[facilityId]` | Facility management |
| `/batches`, `/batches/new`, `/batches/[batchId]` | Batch list, creation, detail with checkpoint timeline |
| `/batches/[batchId]/checkpoints/new` | Log a checkpoint |
| `/claims`, `/claims/new`, `/claims/[claimId]` | Sustainability claims |
| `/verifier/queue`, `/verifier/queue/[claimId]`, `/verifier/history` | Verifier workflow (role-gated) |
| `/credits` | Credit portfolio by Credit Class |
| `/marketplace`, `/marketplace/listings/[listingId]` | Browse and purchase |
| `/marketplace/my-listings`, `/my-listings/new`, `/my-retirements` | Selling and retirement history |
| `/billing`, `/billing/plans` | Current plan, usage, upgrade |
| `/organization/users` | Team members and roles |
| `/organization/api-keys` | Partner API key management |
| `/organization/wallet` | Treasury Address management |
| `/organization/settings` | Organization profile and product categories |
| `/notifications` | Notification centre |
| `/account` | User profile, personal wallet linking, notification preferences |

Verifier accounts are internal, not tied to a customer Organization, and land on `/verifier/queue` after login with a distinct minimal sidebar — they never see Organization, Billing, or Facilities navigation.

**Route protection is enforced in middleware**, before a page renders, using the session and the role claims within it. A client-side role check is a presentation detail; it is never the access control. Every route is additionally protected by the backend's own authorization, so a bypassed frontend check gains nothing.

---

## 9. Sidebar Navigation

A persistent left sidebar, collapsible to an icon rail on smaller viewports, grouped into labeled sections with the active route highlighted using a `primary-50` background and `primary-800` text. Composition is role- and org-type-aware — no user ever sees a menu item for an action their role or organization type cannot perform.

**Manufacturer / Assembler:** Overview (Dashboard) · Provenance (Batches, Facilities) · Sustainability (Claims, Credits) · Trade (Marketplace, My Listings, My Retirements) · Organization (Team, API Keys, Wallet, Settings, Billing) · Notifications

**Logistics Partner:** Overview (Dashboard) · Provenance (Batches — checkpoint logging focus, no batch creation) · Organization (Team, API Keys, Settings, Billing) · Notifications

**Credit Buyer:** Overview (Dashboard) · Trade (Marketplace, My Retirements) · Portfolio (Credits) · Organization (Team, Wallet, Settings) · Notifications
*(No Batches, Facilities, or Claims — Credit Buyers don't produce anything. No Billing beyond payment method, since the Buyer plan is free.)*

**Verifier (internal):** Review (Queue, History) · Notifications

The menu is built from a role-driven configuration rather than hardcoded per page, so adding a capability to a role later — Logistics Partners gaining facilities, for instance — is a configuration change rather than a restructure.

**Unverified and restricted Organizations** see their full menu with gated items visibly disabled and a tooltip naming what is required, rather than having navigation silently disappear. A user whose menu shrinks without explanation assumes the product is broken; a user who sees a disabled item with a reason knows what to do.

---

## 10. Page-by-Page Component Breakdown

### 10.1 Landing Pages — `/`, `/solutions/traceability`, `/solutions/carbon-credits`

These share a marketing component set (`components/features/marketing/`) rather than each inventing its own layout, so adding a future vertical landing page is a matter of composing existing components with new copy.

**Shared marketing components:** `MarketingNav`, `HeroSection`, `FeatureGrid` (composing `FeatureCard`), `HowItWorksSteps`, `StatsBand`, `TestimonialCard`, `CTABanner`, `MarketingFooter`

**`/` adds:** `PersonaSplitSection` (a two-column "track my supply chain" versus "buy verified carbon credits" choice linking to the two solution pages), `FeatureGrid` covering both pillars

**`/solutions/traceability` adds:** traceability-specific `HeroSection`, `FeatureGrid` (batch tracking, checkpoint history, Provenance Score, multi-industry support), `HowItWorksSteps`, `IndustryLogosBand`, `CTABanner`

**`/solutions/carbon-credits` adds:** credit-specific `HeroSection`, `FeatureGrid` (AI-assisted plus human verification, Credit Class attribution, transparent retirement), `HowItWorksSteps`, `TrustBand` (a short explainer on what double-counting prevention does and does not cover, per PRD Section 1.3, linking to the public retirement log as proof), two distinct `CTABanner` targets given this page serves both sellers and buyers

### 10.2 `/dashboard`

Role-aware summary: recent batch activity, claim status, credit balance snapshot and plan usage for a Manufacturer/Assembler; portfolio and recent marketplace activity for a Credit Buyer; queue depth and recent decisions for a Verifier.

**Components:** `PageHeader`, `MetricCard` (×3–4), `RecentActivityList` (composing `Timeline`), `PlanUsageWidget`, `QuickActionButtons`, `PendingTransactionBanner`, `VerificationStatusBanner` (for unverified or restricted Organizations)

### 10.3 `/facilities` and `/facilities/[facilityId]`

**List:** `PageHeader` with an add action, `DataTable` (name, type, grid region, verification status, Trust Tier, batch count), `TrustTierBadge`, `EmptyState`

**Detail:** `PageHeader`, `FacilityProfileCard`, `TrustTierBadge`, `VerificationStatusBadge`, `CapacityCard` (attested versus declared capacity, and the resulting ceiling discount factor — shown explicitly, since a facility owner needs to understand why their ceiling is what it is), `Tabs` (Overview / Batches / Claims / Credits), `MetricCard`, nested `DataTable` per tab

### 10.4 `/batches` and `/batches/[batchId]`

**List:** `PageHeader`, `DataTable` (batch ID, product category, quantity, status, Provenance Score), `ProductCategoryBadge`, `ProvenanceScoreBadge`, `FilterBar`, `Pagination`

**Detail:** `PageHeader`, `BatchSummaryCard`, `ProvenanceScoreBadge` with a `ProvenanceScoreBreakdown` popover showing each component's contribution, `Timeline` (chronological checkpoints), `CheckpointCard`, `CorrectionIndicator` (superseded checkpoints shown struck through with their correction linked), `AnchorStatusIndicator` (whether each checkpoint is included in a confirmed on-chain epoch, with a proof link), `ParentBatchLinks` (one level up only, per PRD Section 9.2), `QRCodeDisplay` with copy and regenerate actions, `AddCheckpointButton`

The score breakdown exists because an unexplained number invites distrust: a manufacturer seeing "Partial, 58" needs to know that logging two missing checkpoint types would move it, not be left guessing.

### 10.5 `/batches/new`

**Components:** `PageHeader`, `MultiStepForm` (using the Zustand draft store), `ProductCategorySelector`, `BatchDetailsFormStep`, `EvidenceUploadStep` (via `FileDropzone`), `ReviewSummaryStep`, `FormNavigationFooter`

The Product Category selector warns explicitly that the choice is permanent for the batch's lifetime before the step can be completed.

### 10.6 `/claims` and `/claims/[claimId]`

**List:** `PageHeader`, `DataTable` (activity type, period, submitted date, status, computed ceiling, requested amount), `ClaimStatusPill`, `FilterBar`

**Detail:** `PageHeader`, `ClaimStatusStepper` (Submitted → AI Review → Human Review → Decided), `ClaimDetailsCard`, `CeilingComparisonCard` (requested versus computed ceiling, with the discount factor and its reason), `EvidenceViewer`, `DecisionOutcomeCard` (approval or rejection reason and issued amount, once decided), `ResubmitButton` (when more information was requested)

**The AI Agent's assessment and reasoning are never rendered on this page.** Showing a claimant exactly which discrepancies the AI detects teaches them precisely what to adjust next time, which converts a fraud control into a fraud tutorial. The claimant sees status and the human decision; the reasoning is internal to the Verifier view.

### 10.7 `/claims/new`

**Components:** `PageHeader`, `MultiStepForm`, `ActivityTypeSelector`, `ClaimFiguresFormStep` (activity-specific fields rendered dynamically), `EvidenceUploadStep`, `CalculatedCeilingPreview`, `ExclusivityAttestationStep`, `ReviewSummaryStep`

`CalculatedCeilingPreview` calls the backend for the figure rather than computing it client-side. A client-side estimate that disagrees with the backend's authoritative number is worse than no preview, and the formula involves reference-table versions the client has no business holding.

The exclusivity attestation (PRD Section 3.3) is its own step with its own confirmation, not a checkbox appended to a form — it is a legal representation, and its prominence should match that.

### 10.8 `/verifier/queue` and `/verifier/queue/[claimId]`

**Queue:** `PageHeader`, `DataTable` (facility, activity type, submitted date, requested amount, priority), `PriorityBadge`, `QueueDepthMetrics`, `FilterBar`

**Review detail:** `PageHeader`, `ClaimDetailsCard`, `CeilingComparisonCard`, `EvidenceViewer` with page-level navigation, `AIAssessmentPanel`, `SecondApprovalBanner` (when this claim already carries one approval, showing the first Verifier's decision and reasoning), `DecisionActionBar` (Approve / Reject / Request More Info), `ConfirmDialog`

`AIAssessmentPanel` shows the extracted figures with **clickable citations that jump the evidence viewer to the cited page**, the discrepancy analysis, the confidence score, the flags raised, and — where relevant — an explicit warning that instruction-like content was detected in a document. The citation link is the panel's most important feature: it turns the AI from something a Verifier must decide whether to trust into something they can check in two seconds. A recommendation without a citation is an opinion; a recommendation with one is a shortcut to the evidence.

The approve control is **hard-capped at the computed ceiling** — the input cannot be set higher, rather than accepting a higher value and warning about it. Reject requires a reason of at least 40 characters.

### 10.9 `/credits`

**Components:** `PageHeader`, `MetricCard` (total balance), `CreditClassTable` (originating facility, vintage, activity type, quantity, status), `CreditClassBreakdownChart` (Recharts stacked bar), `SellButton`, `RetireButton`

Balance is always presented broken down by Credit Class and never as a single pooled figure, matching PRD Section 2.3. A total is shown as a summary line, clearly labeled as an aggregate across classes rather than as a spendable amount, since it is not one — every sale and every retirement operates on a specific class.

### 10.10 `/marketplace` and `/marketplace/listings/[listingId]`

**Browse:** `PageHeader`, `FilterBar` (activity type, vintage, region, originating facility, Trust Tier, price range), `ListingCard`, `TrustTierBadge`, `Pagination`, `EmptyState`

**Detail:** `PageHeader`, `ListingDetailCard`, `SellerFacilityCard` (originating facility disclosed, not anonymized), `CreditClassSummary`, `DoubleCountingDisclosure` (the PRD Section 1.3 boundary, shown on the page rather than buried in terms), `PurchaseForm` (quantity respecting the minimum and the remainder rule, with a live `quote` breakdown of cost, fee, and total), `WalletAuthorizationState`, `TransactionStatusToast`

### 10.11 `/marketplace/my-listings`, `/my-listings/new`, `/my-retirements`

**My listings:** `PageHeader`, `DataTable` (Credit Class, quantity remaining, price, status, expiry), `CancelListingButton`, `ConfirmDialog`, `ExpiryWarningBadge` (listings within 7 days of expiry)

**New listing:** `PageHeader`, `ListingForm` (Credit Class selection from balance, quantity, price per tCO2e, minimum purchase quantity, expiry), `PricePreview` (gross, platform fee at the Organization's plan rate, net proceeds), `EscrowExplainer`

`EscrowExplainer` states plainly that listed credits leave the spendable balance immediately. This is the single most surprising behaviour in the marketplace and explaining it up front costs one paragraph, where not explaining it costs a support ticket per seller.

**Retirements:** `PageHeader`, `DataTable` (Credit Class, quantity, reason, date, transaction link), `RetireCreditsButton` → `RetirementForm` in a `Dialog`, with an explicit irreversibility confirmation

### 10.12 `/billing` and `/billing/plans`

**Current plan:** `PageHeader`, `PlanSummaryCard`, `PlanUsageWidget` (per metered dimension, each a progress bar with hard-limit or overage-rate messaging depending on plan), `OverageEstimateCard`, `InvoiceHistoryTable`, `PaymentMethodCard`, `PaymentFailureBanner` (during a grace period, with days remaining and exactly what happens at the end)

**Plan comparison:** `PageHeader`, `PlanComparisonTable` rendered dynamically from backend plan data — never hardcoded, since Admins change plan details — `UpgradeButton`, `DowngradeImpactDialog` (what an Organization would lose and what happens to over-limit resources), `ConfirmDialog`

### 10.13 `/organization/users`, `/api-keys`, `/wallet`, `/settings`

**Users:** `PageHeader`, `DataTable` (name, email, role, MFA status), `InviteUserButton` → `InviteUserForm` in a `Dialog`, `RoleSelector`, `MFARequiredBadge` (on roles requiring it)

**API keys:** `PageHeader`, `DataTable` (name, prefix, created, last used, status), `CreateAPIKeyButton` → `Dialog` showing the key **once** with a copy affordance and an explicit "this will not be shown again" warning, `RevokeKeyButton` with `ConfirmDialog`, `KeyUsageChart`

**Wallet:** `PageHeader`, `TreasuryAddressCard` (current address, credits held, active listings), `ChangeTreasuryAddressFlow` (a multi-step flow: connect the new wallet, sign to prove ownership, MFA re-authentication, confirm — followed by a 72-hour pending state), `PendingChangeBanner` (countdown, who initiated it, and a cancel action available to any Owner), `TreasuryChangeHistory`

The pending-change banner is the security control made visible. The 72-hour delay only protects anyone if the people who could cancel it actually see it, so it appears on the dashboard as well as this page, for every Owner and Admin.

**Settings:** `PageHeader`, `OrganizationProfileForm`, `VerificationStatusCard` (status, what it gates, and how to resolve it), `ProductCategorySelector`

### 10.14 `/notifications`

**Components:** `PageHeader`, `NotificationList` (composing `NotificationItem`, grouped read/unread), `MarkAllReadButton`, `NotificationFilterTabs`, `DigestGroupItem` (collapsed groups from bulk activity, expandable)

### 10.15 `/account`

**Components:** `PageHeader`, `UserProfileCard`, `MFASettingsCard`, `PersonalWalletCard` (linked address via `AddressDisplay`, with an explicit note that a personal wallet is for signing on the Organization's behalf and never holds Organization credits), `NotificationPreferencesForm`, `ActiveSessionsList` with a revoke action

### 10.16 `/track/[publicRef]` (Public)

**Components:** `PublicLayout` (a lightweight utility header and footer, distinct from the marketing chrome — this is a functional destination reached by QR scan, not a marketing page), `BatchPublicSummary`, `Timeline` (the same component reused from the authenticated batch detail page), `ProvenanceScoreBadge` with its plain-language band explanation, `SustainabilityHighlightCard` (the originating facility's approved claims, if any), `VerifyOnChainLink` (an explorer link plus the inclusion proof, for a visitor who wants to verify rather than trust), `ShareButton`

The page shows checkpoint types, locations, timestamps, the originating facility's name and country, and the Provenance Score. It shows no quantities, no prices, no counterparty names, and no evidence — the consumer-facing story does not require any of them, and each would leak commercially sensitive data to an unauthenticated visitor.

### 10.17 `/marketplace/retirements` (Public)

**Components:** `PublicLayout`, `PageHeader`, `DataTable` (retiring organization, originating facility, vintage, activity type, quantity, reason, date, transaction link), `Pagination`, `FilterBar` (facility, vintage, activity type, date range — public fields only)

### 10.18 `/pricing` (Public)

**Components:** `MarketingNav`, `MarketingFooter`, `PlanComparisonTable` (the same component as `/billing/plans` — one component, two contexts), `SignupCTAButton`, `BuyerPlanCallout` (making it obvious that buying and retiring credits is free, since a pricing page that leads with $49 will lose buyers before they read the table)

---

## 11. Cross-Cutting Concerns

- **Loading states:** every data-fetching component renders a `Skeleton` matching its final content's shape, never a generic spinner.
- **Empty states:** every list page defines an `EmptyState` with a clear next action, never a bare empty table.
- **Error boundaries:** each route segment has a scoped error boundary rendering `PageErrorState` from the backend's standard envelope — the generic message plus the `request_id`, with a copy affordance so a user can quote it to support. The client never attempts to parse or display anything beyond that.
- **Optimistic updates are permitted for presentation only, never for value.** A filter change, a notification marked read, a draft saved — these update optimistically. **Credit balances, purchase outcomes, listing fills, retirement status, and issued amounts never do**, without exception. A user shown "retired" that later reverts has been given a false compliance artifact, and the trust cost of that is out of all proportion to the responsiveness gained.
- **Rate limit handling:** a `429` surfaces a clear message with the `Retry-After` interval rather than a generic error, and automatic retry respects it. Bulk operations in the UI are client-side throttled to stay within the Organization's plan limit rather than firing and failing.
- **Responsive behavior:** the sidebar collapses to an icon rail below a defined breakpoint, and `DataTable` degrades to a stacked card layout on small viewports rather than horizontal scrolling — defined once in the shared component, never re-implemented per page.
- **Financial formatting:** credit amounts render to 6 decimal places with trailing zeros trimmed and a `tCO2e` suffix; USDC to 2 decimal places with the symbol. Both come from exact decimal strings through `CreditAmountDisplay` and `UsdcAmountDisplay` — never a JavaScript number, never a locally computed value.
- **Address display:** `AddressDisplay` truncates to first six and last four characters with a copy affordance and, where the address is known to the platform, the Organization or facility name alongside it. A user comparing raw hex to confirm they are transacting with the right party is a user about to make a mistake.
- **Timestamps:** displayed in the viewer's local timezone with the UTC value in a tooltip. Supply chain events span time zones and an ambiguous timestamp is a real source of dispute.

---

## 12. Testing

- **Component tests** (Vitest + Testing Library) for every shared and domain component, driven by behaviour rather than implementation detail.
- **Contract tests** against the backend's generated types, so an API change that breaks the frontend fails CI rather than production.
- **End-to-end tests** (Playwright) covering the flows where a failure is expensive: registration and verification, batch creation, claim submission, the full Verifier decision path including dual approval, marketplace purchase with a mocked wallet, retirement, and the Treasury Address change flow including cancellation.
- **Accessibility tests** — automated axe checks on every route, plus the palette contrast assertions from Section 5.3.
- **Visual regression** on the primitive and shared component tiers, since those are where an unintended change propagates furthest.
