# Product Requirements Document: CarbonCircuit

**A Multi-Industry Provenance & Carbon Credit Platform — Launching with Electronics Components**

---

## 0. Why This Document Exists

This PRD is the source of truth for CarbonCircuit. It describes *what* we are building and *why*, not *how* it will be technically implemented. Any developer joining this project should be able to read this document and understand the product, the users, the data flowing through the system, and the exact conditions under which things happen — before ever opening an architecture diagram.

**A note on why this project exists:** I'm building CarbonCircuit as a personal project to grow my hands-on skills in building real-life, production-grade Web3 applications — specifically around combining blockchain (smart contracts) with modern backend patterns (event-driven microservices), and applied AI (an AI agent embedded in a real workflow, not a chatbot bolted onto the side).

**A note on scope:** CarbonCircuit is designed as a **commodity-agnostic platform** — provenance tracking and carbon credit issuance are relevant to many industries (agriculture, pharma, textiles, electronics, and more), not just one. Electronics manufacturing was chosen as the **first vertical** to build and demo against, deliberately, because it's a real, well-documented industry pain point (conflict minerals, e-waste, opaque multi-tier supply chains). Every feature below should be understood as "electronics is our launch example," not "electronics is the only thing this system can ever support."

---

## 1. Product Overview

### 1.1 What is CarbonCircuit?

CarbonCircuit is a platform that does two connected things for physical supply chains, built to work across multiple industries (see Section 2.2 on Product Categories), launching first with electronics components:

1. **Provenance Tracking** — it records the journey of a physical product or batch (e.g., a batch of PCBs, semiconductor chips, or battery cells for our launch vertical) from raw material sourcing through manufacturing, assembly, and final shipment — creating a tamper-evident, verifiable history.

2. **Carbon Credit Issuance & Marketplace** — it allows manufacturers and facilities in that supply chain to have their sustainable practices (renewable energy use, verified emission reductions, responsible material sourcing) independently reviewed and converted into tradeable carbon credits, which other companies can buy to offset their own emissions.

The two systems are connected but distinct: **provenance answers "where did this come from and is it authentic?"**, while **carbon credits answer "how much verified environmental benefit did this facility generate, and can it be bought/sold/claimed?"**

### 1.2 The Problem We're Solving

Electronics supply chains today are:
- **Opaque** — a finished device may pass through 5-8 different facilities/vendors across multiple countries, and most brands (let alone consumers) have no verifiable way to confirm origin claims (e.g., "conflict-free" minerals, "responsibly sourced" materials).
- **Prone to greenwashing** — sustainability claims from suppliers are often self-reported with no independent verification, and carbon credit markets globally have faced credibility scandals from unverifiable or double-counted credits.
- **Fragmented across manual paperwork** — checkpoint data (customs, warehouse transfers, quality checks) is often tracked in disconnected spreadsheets or siloed vendor systems, making end-to-end traceability nearly impossible to reconstruct after the fact.

### 1.3 What CarbonCircuit Does and Does Not Guarantee

This distinction is stated up front because the entire credibility of the product rests on being precise about it, and because a developer who misunderstands it will build the wrong safeguards.

**What the platform guarantees:** a credit issued on CarbonCircuit is issued exactly once against exactly one approved Sustainability Claim, can be transferred only through recorded transactions, and once retired can never be resold, re-transferred, or retired again by anyone. Every one of those properties is enforced on an immutable ledger, not by a database row an administrator could edit.

**What the platform does not guarantee:** that the underlying real-world emissions reduction has not *also* been registered with an external registry (Verra, Gold Standard, a national scheme). CarbonCircuit prevents double-*retirement of a CarbonCircuit credit*; it cannot, on its own, prevent double-*issuance of the underlying reduction* across unconnected registries. The mitigations we do apply are: a mandatory exclusivity attestation at claim submission, a duplicate-evidence check across the whole platform (Section 3.6), and the public retirement log that lets any third party audit what was claimed. This limitation is disclosed to buyers on every listing page rather than buried, because a platform that overstates this is exactly the kind of platform this product exists to be an alternative to.

### 1.4 Who This Is For

| User Type | Description |
|---|---|
| **Component Manufacturer / Facility** | Produces raw components (chips, PCBs, batteries, connectors) and wants to prove authenticity and claim sustainability credit for their practices |
| **Assembler / OEM** | Buys components from manufacturers, assembles finished goods, wants to show end customers a verified supply chain |
| **Logistics Partner** | Moves goods between facilities, responsible for scanning/reporting checkpoint events |
| **Verifier / Auditor** | An independent reviewer (internal team member, in this product's scope) who reviews sustainability evidence and approves or rejects carbon credit issuance |
| **Credit Buyer** | Any company that wants to purchase carbon credits to offset their own reported emissions |
| **End Consumer** | Someone who scans a QR code on a finished product to see its origin and sustainability story |
| **Platform Admin / Ops Team** | Internal team responsible for platform health, fraud review, and dispute resolution |

---

## 2. Key Concepts & Definitions

These terms are used throughout this document and the product. Every developer should understand them before building any feature.

| Term | Definition |
|---|---|
| **Batch** | A defined quantity of a specific component produced together (e.g., "10,000 units of Lithium-ion cell Model X, produced March 2026 at Facility A"). This is the core unit of provenance tracking. |
| **Checkpoint** | A recorded event marking the batch's movement or state change — e.g., "left factory," "cleared customs," "received at assembly plant." Each checkpoint has a location, timestamp, and reporting party. |
| **Facility** | A physical manufacturing or processing site, owned/operated by a registered Manufacturer or Assembler account. Sustainability claims and carbon credits are tied to facilities, not to individual batches. |
| **Sustainability Claim** | A formal submission by a facility describing a specific sustainability practice over a defined time period (e.g., "Q1 2026: 40% of energy from verified solar source"), submitted with supporting evidence. This is what gets reviewed and, if approved, converted into carbon credits. |
| **Carbon Credit** | A tradeable unit representing one verified metric ton of CO2-equivalent emissions avoided or reduced, issued only against an approved Sustainability Claim. Every credit permanently carries the identity of the **facility that earned it**, its **vintage**, and its **activity type** — see Section 2.3. |
| **Credit Class** | The permanent, inseparable combination of (originating Facility, Vintage, Activity Type) that defines what a credit actually represents. Credits of the same Credit Class are interchangeable with each other; credits of different Credit Classes are not, and are never pooled. See Section 2.3. |
| **Vintage** | The calendar year during which the emissions reduction represented by a carbon credit actually occurred. Credits from different vintages are not treated as identical — a 2024 credit and a 2026 credit represent different claims and are distinguishable everywhere in the marketplace. |
| **Retirement** | The act of a credit buyer permanently "using" a carbon credit to claim an offset. Once retired, a credit can never be sold or claimed again. This is the mechanism that prevents the same environmental benefit from being claimed twice by different parties. |
| **Provenance Score** | A 0–100 consumer-facing rating reflecting how complete, timely, and independently anchored a batch's checkpoint history is, plus whether its originating facility has approved sustainability claims. The exact formula is in Section 2.4. |
| **Trust Tier** | A facility-level classification (New / Verified / Trusted) computed from submission history, approval rate, registry verification status, and fraud history. The exact formula is in Section 2.5. |
| **Product Category** | The industry/commodity vertical a Batch belongs to (e.g., Electronics, Agriculture, Pharma, Textiles). Determines which attributes, checkpoint expectations, and sustainability claim types are relevant for that batch. See Section 2.2. |
| **Tenant / Organization** | Every Organization on the platform is treated as an isolated tenant — its data is logically separated from every other Organization's, even though all tenants share the same underlying platform. See Section 9. |
| **Plan** | The subscription tier an Organization is on (Buyer / Starter / Growth / Enterprise), which determines usage limits, included AI-reviewed claims, API access, and rate limits. See Section 7. |
| **Organization Treasury Address** | The single blockchain address that receives, holds, and spends an Organization's carbon credits. Distinct from any individual user's personal wallet. See Section 3.8. |
| **Verification Status** | Whether an Organization and its Facilities matched the business registry reference dataset at registration. Gates the ability to submit claims and receive credits. See Section 3.1. |

### 2.1 How a Carbon Credit Amount Is Calculated (Formula Definitions)

This is one of the most important sections in this document — every developer building the sustainability claim or credit issuance features must understand this, since it defines what "correct" looks like.

Carbon credits are **never calculated per batch**. They are calculated **per facility, per claim, per time period**, based on the specific type of sustainability activity being claimed. Every credit amount is expressed in **metric tons of CO2-equivalent (tCO2e)**, carried to **6 decimal places**, and every intermediate value is computed in exact decimal arithmetic — never floating point.

Each activity type has its own defined formula, and each formula draws its multiplier from a **maintained reference table** (Section 2.6), never from a value invented per claim and never from a figure the submitting facility supplies.

**Renewable Energy Usage**
```
Credits (tCO2e) = (Verified kWh from renewable source) × (Regional grid emission factor in kgCO2e/kWh) ÷ 1000
```
The regional grid emission factor is a published, standard reference value that varies by grid region, looked up from the reference table in Section 2.6.1 using the facility's registered grid region and the claim's vintage year.

**Reduced-Emission Logistics**
```
Credits (tCO2e) = [ (Baseline factor − Actual factor) × Tonne-kilometres shipped ] ÷ 1000
```
Both factors are in kgCO2e per tonne-kilometre. The baseline factor comes from the reference table in Section 2.6.2 for the standard method on that route; the actual factor is derived from submitted evidence (carrier fuel records, verified carrier sustainability reports). A claim where the actual factor is greater than or equal to the baseline yields zero credits and is auto-rejected at submission rather than entering the review queue.

**Responsible Material Sourcing / Waste Reduction**
```
Credits (tCO2e) = (Verified quantity of recycled/responsibly-sourced material) × (Emissions-avoided factor for that material)
```
The emissions-avoided factor is a published reference value per material type (Section 2.6.3), expressed per tonne or per kilogram depending on the material.

#### The Credit Ceiling

Every Sustainability Claim carries a **maximum possible credit ceiling**, computed automatically at submission time. A claim can never issue more than its ceiling, regardless of what the facility requested or what a Verifier approves.

```
Ceiling = (Formula result using the facility's attested operating capacity for the claim period)
          × Verification discount factor
```

The **verification discount factor** reflects how much independent corroboration exists for the facility's declared scale:

| Facility verification status | Discount factor | Rationale |
|---|---|---|
| Matched in the facility registry dataset, attested capacity used | 1.00 | Capacity figure is independently sourced |
| Registry-matched organization, facility not individually matched | 0.75 | Organization is real, facility scale is self-declared |
| No registry match (self-declared only) | 0.50 | Nothing about the declared scale is corroborated |

A facility with no registry match therefore cannot claim more than half of what its own self-declared capacity would theoretically support. This is the single most important guard against the most obvious attack on the platform: registering a fictional facility with an enormous declared capacity and minting credits against it.

Verifiers see three numbers side by side when reviewing: the amount requested, the computed ceiling, and the AI-extracted figure from the evidence. **A claim requesting more than its ceiling is flagged at priority `high` and cannot be approved for more than the ceiling** — the approval control is hard-capped in the UI, not merely warned about.

### 2.2 Product Categories (Multi-Industry Support)

CarbonCircuit is not built exclusively for electronics — it's built to support many industries, with electronics as the first one we launch and demo. This is a deliberate, foundational product decision, not an afterthought, and it affects several features described below.

- Every **Batch** belongs to exactly **one Product Category** (e.g., Electronics, Agriculture, Pharma, Textiles), chosen at the time the batch is created. **A batch's Product Category cannot be changed after creation** — this keeps everything downstream (claim types, checkpoint expectations, provenance presentation) unambiguous for that batch's entire lifecycle.
- Each Product Category defines its own **relevant attributes** for a batch (e.g., electronics batches track "component type / lot number"; agriculture batches would track "crop type / harvest region"). A developer adding a new Product Category later should be able to do so by defining a new attribute set, not by changing the core batch model.
- Each Product Category also determines which **Sustainability Claim types** are available to facilities operating in it — electronics facilities can claim things like responsible material sourcing, while an agriculture facility (in a future version) might claim regenerative farming practices instead. Not every claim type applies to every category.
- Each Product Category defines its **expected checkpoint sequence**, which is what the Provenance Score's completeness component (Section 2.4) is measured against. For Electronics the expected sequence is: `production_complete` → `departed_origin` → `customs_export` → `customs_import` → `arrived_destination`.
- **Batches can reference parent batches across different categories.** A single batch is locked to one category, but a finished product's component chain can span categories — e.g., a finished electronic device batch (Electronics) might reference a separate sustainable-packaging batch (a different category). This keeps each batch's own rules clean while still allowing realistic supply chains to be represented.
- At Organization setup, a company declares which Product Categories are relevant to their business, which determines what they see by default in their dashboard (an electronics-only manufacturer shouldn't be shown agriculture-specific claim types).

### 2.3 Credit Classes — Why a Credit Is Never Just a Number

A carbon credit on CarbonCircuit is **not** an undifferentiated balance. Every credit permanently carries three attributes that together form its **Credit Class**:

1. **Originating Facility** — which specific facility's verified practice produced this credit
2. **Vintage** — the calendar year in which the reduction occurred
3. **Activity Type** — renewable energy / reduced-emission logistics / responsible sourcing

These three attributes travel with the credit through every transfer and are still attached at the moment it is retired. A buyer purchasing credits is always purchasing a specific Credit Class, not an anonymous quantity, and the public retirement log names all three.

**Why this is a product requirement and not an implementation detail:** the core value proposition to a Credit Buyer (PRD Section 3.5) is that they can point at a specific facility's specific verified practice in their own ESG report. A credit that has been pooled with other facilities' credits cannot support that claim, and a buyer who discovers after purchase that "which facility earned this" is unanswerable has been sold exactly the kind of unverifiable instrument this platform exists to replace. Credits of different Credit Classes are therefore never merged, never averaged, and never presented as a single fungible balance anywhere in the product.

Practically, this means: balances are always shown broken down by Credit Class; a marketplace listing is always for exactly one Credit Class; and a retirement record always names the originating facility.

### 2.4 Provenance Score — Exact Formula

The Provenance Score is a 0–100 integer, recomputed whenever a batch's underlying data changes, and displayed publicly on the QR-scan page. It is composed of five weighted components:

| Component | Points | How it's measured |
|---|---|---|
| **Checkpoint completeness** | 40 | `(distinct expected checkpoint types present ÷ expected checkpoint types for this Product Category) × 40` |
| **On-chain anchoring** | 20 | `(checkpoints included in a confirmed on-chain anchor ÷ total checkpoints) × 20` |
| **Chain depth resolution** | 15 | Full 15 if every declared parent batch resolves to a registered batch with a known originating facility; `(resolved ÷ declared) × 15` otherwise; full 15 if the batch declares no parents |
| **Reporting timeliness** | 15 | Full 15 if the median lag between a checkpoint's stated event time and its submission time is under 24 hours; scales linearly to 0 at a 7-day median lag |
| **Facility sustainability record** | 10 | Full 10 if the originating facility has at least one approved Sustainability Claim with a vintage within the last 4 years; 5 if it has an approved claim older than that; 0 otherwise |

**Presentation bands** (what the consumer actually sees, since a bare number invites false precision):

| Score | Band |
|---|---|
| 90–100 | Complete |
| 70–89 | Strong |
| 40–69 | Partial |
| 0–39 | Limited |

A batch with zero checkpoints scores 0 and displays as *Limited* with an explicit "this batch has no recorded journey yet" message rather than an unexplained low number.

### 2.5 Trust Tier — Exact Formula

Trust Tier is computed per Facility and recomputed on every claim decision and every fraud-flag state change. It is displayed on marketplace listings and is a filter buyers can sort by, so it needs a definition that is defensible rather than impressionistic.

| Tier | All of these must hold |
|---|---|
| **New** | Default for every new facility. Any facility not meeting Verified criteria. |
| **Verified** | ≥ 3 decided claims · approval rate ≥ 70% · organization verification status is `verified` · zero escalated fraud flags in the last 180 days |
| **Trusted** | ≥ 10 decided claims · approval rate ≥ 90% · organization verification status is `verified` · facility individually matched in the facility registry dataset · claims spanning ≥ 2 distinct activity types · zero escalated fraud flags in the last 365 days |

**Approval rate** counts only claims that reached a final Approved or Rejected state; claims withdrawn or still pending are excluded from both numerator and denominator.

**Demotion is immediate and asymmetric.** Any fraud flag reaching `escalated` status drops the facility to **New** immediately, and it cannot rise above New for 90 days after that flag is resolved. Tier is easy to lose and slow to regain, deliberately — the tier is a signal buyers rely on when spending money, and a slow-to-demote tier would make it worthless.

### 2.6 Reference Factor Tables (Seeded Values)

These tables are seeded into the database at deployment and maintained by Admins through the Admin Portal. Every value is versioned with an effective date range, so a claim for vintage 2025 uses the factor that was in effect for 2025 even if the table is later updated. **A claim's computation always pins the reference-table row version it used**, and that version is stored with the claim forever — a recomputation years later must produce the same answer.

#### 2.6.1 Grid Emission Factors (kgCO2e per kWh)

| Grid region | Factor |
|---|---|
| `US-CAISO` (California) | 0.237 |
| `US-ERCOT` (Texas) | 0.396 |
| `US-PJM` (Mid-Atlantic) | 0.324 |
| `US-MISO` (Midwest) | 0.438 |
| `EU-DE` (Germany) | 0.381 |
| `EU-FR` (France) | 0.056 |
| `EU-PL` (Poland) | 0.662 |
| `UK` | 0.207 |
| `CN-East` | 0.581 |
| `CN-South` | 0.512 |
| `IN-North` | 0.713 |
| `JP` | 0.462 |
| `KR` | 0.436 |
| `TW` | 0.494 |
| `VN` | 0.521 |
| `MY` | 0.585 |
| `SG` | 0.412 |
| `TH` | 0.499 |

#### 2.6.2 Logistics Baseline Factors (kgCO2e per tonne-kilometre)

| Shipping method | Baseline factor |
|---|---|
| Air freight, short-haul (< 1,500 km) | 1.316 |
| Air freight, long-haul (≥ 1,500 km) | 0.602 |
| Sea freight, container | 0.016 |
| Sea freight, bulk | 0.008 |
| Rail, electric | 0.024 |
| Rail, diesel | 0.032 |
| Road, heavy goods vehicle (> 32t) | 0.086 |
| Road, light goods vehicle | 0.231 |
| Inland waterway | 0.031 |

#### 2.6.3 Material Emissions-Avoided Factors

| Material (recycled/recovered, vs. virgin) | Unit | Factor |
|---|---|---|
| Aluminium | tCO2e / tonne | 8.10 |
| Copper | tCO2e / tonne | 2.60 |
| Steel | tCO2e / tonne | 1.09 |
| Tin | tCO2e / tonne | 3.40 |
| Gold | tCO2e / kg | 15.80 |
| Tantalum | tCO2e / kg | 0.94 |
| Plastics (post-consumer ABS) | tCO2e / tonne | 1.85 |
| Plastics (post-consumer PET) | tCO2e / tonne | 1.53 |
| Rare-earth magnets (NdFeB, recovered) | tCO2e / tonne | 22.60 |
| Lithium-ion cell black mass | tCO2e / tonne | 4.70 |

---

## 3. Features (Section by Section)

### 3.1 Organization Registration, Verification & Facility Management

**What it does:** Allows Manufacturers, Assemblers, Logistics Partners, and Credit Buyers to register an organization account, be checked against a business registry reference dataset, and (where relevant) add one or more physical Facilities.

**What a user does:**
- Registers an organization: company name, organization type (Manufacturer / Assembler / Logistics / Credit Buyer), country of incorporation, and **business registration number**
- Adds one or more Facilities, each with: name, physical address, grid region (from the Section 2.6.1 list), facility type (raw material processing / component fabrication / assembly / distribution), declared annual production capacity, and declared annual energy consumption
- Views their verification status, Trust Tier, and submission history

#### Registry Verification

The submitted business registration number is checked against a **seeded business registry reference dataset** — approximately 500 rows covering real-shaped registration records across the launch markets, keyed on `(country_code, registration_number)`. Each row carries: legal entity name, registered address, incorporation date, entity status (`active` / `dissolved`), industry classification codes, and a sanctions/restricted-party flag.

The check produces one of three outcomes, and the outcome determines what the Organization is allowed to do:

| Outcome | Condition | What the Organization can do |
|---|---|---|
| **`verified`** | Exact match on `(country, registration_number)`, entity status `active`, name similarity to the registered legal name ≥ 85%, no sanctions flag | Everything their Plan allows |
| **`unverified`** | No match found in the dataset | Log in, create Facilities, create Batches, log Checkpoints, browse the marketplace, **buy** credits. **Cannot** submit Sustainability Claims, **cannot** receive credit issuance, **cannot** list credits for sale. |
| **`rejected`** | Match found but entity status is `dissolved`, or the sanctions flag is set, or name similarity < 85% | Account is created but immediately suspended pending manual Admin review. No platform actions available. |

An `unverified` Organization sees a persistent, dismissible-per-session banner explaining exactly which capabilities are gated and how to resolve it (correct the registration number, or contact support for manual verification). An Admin can manually promote an Organization to `verified` from the Admin Portal, and that manual promotion is recorded with the Admin's identity and a required justification.

Facilities are checked separately against a **facility registry reference dataset** keyed on `(organization_registration_number, facility_reference)`, which carries an independently attested annual production capacity and energy consumption. Whether a facility matches determines its verification discount factor in the credit ceiling formula (Section 2.1) — this is the mechanism by which self-declared scale is treated with appropriate suspicion rather than taken at face value.

**A note on this approach:** using a seeded reference dataset rather than a live third-party registry API is a deliberate scoping decision. The verification *logic*, the *gating* it drives, and the *discount factors* it feeds are all real and fully implemented; only the data source is a fixture. Swapping the fixture for a live registry API later is a change to one lookup implementation behind an interface, not a change to any of the product rules above.

---

### 3.2 Batch Registration & Provenance Tracking

**What it does:** Lets a Manufacturer or Assembler create a Batch record representing a produced quantity of a component, and lets any authorized party in the chain log Checkpoints as that batch physically moves.

**What a user does:**
- **Manufacturer:** Creates a new Batch — selects the **Product Category** (permanent for this batch, see Section 2.2), then enters component type, quantity, production date, originating Facility, and optionally links supporting documents/certifications (e.g., material origin certificates) and photos.
- **Logistics Partner / Receiving Facility:** Logs a Checkpoint whenever the batch changes hands or location — e.g., "departed Facility A," "arrived at customs," "received at Assembler's warehouse." Each checkpoint requires a checkpoint type, location, event timestamp, and is tied to the account that reported it.
- **Assembler:** When incorporating a component batch into a finished product, links the component batch(es) to the finished product's own Batch record — creating a chain of custody from raw component to finished good.
- **End Consumer:** Scans a QR code on a finished product (no login required) and sees the full checkpoint history, originating facility information, and Provenance Score, presented as a simple visual timeline.

**Data note for developers:** A Batch can have a parent-child relationship with other Batches (a finished product batch references the component batches used within it). This needs to support multiple levels — a finished laptop batch might reference a PCB batch, which itself references a raw semiconductor batch. Chain depth is capped at **10 levels** and a cycle in the parent graph is rejected at write time.

#### Public Batch References

The URL a consumer reaches by QR scan uses a **public batch reference** — a 22-character, randomly generated, non-sequential identifier — and never the batch's internal identifier.

This matters more than it looks. Internal identifiers are time-ordered, which means a party holding a handful of them can infer creation ordering and volume across the whole platform. Batch volumes and shipment cadence are commercially sensitive, competing manufacturers share this platform, and tenant isolation is a foundational requirement (Section 9). A public identifier must therefore be independently random, and must be regenerable by the owning Organization if a QR code is compromised — regenerating it invalidates previously printed codes, so the UI states that consequence plainly before confirming.

#### Checkpoint Correction

Checkpoints are append-only; a logged checkpoint is never edited or deleted. A party who logged an incorrect checkpoint files a **correction**, which creates a new checkpoint record marked as superseding the original and requires a reason. Both records remain visible in the batch's history — the original shown as struck through with its correction linked — because silently rewriting a supply chain record is precisely the failure mode this product exists to prevent. Only the Organization that logged the original checkpoint may correct it, and only within **7 days** of logging; after that, correction requires an Admin action.

---

### 3.3 Sustainability Claim Submission

**What it does:** Allows a verified Facility to submit a formal claim describing a sustainability practice over a defined period, along with supporting evidence, for review.

**What a user does:**
- Selects the type of sustainability activity being claimed (renewable energy, reduced-emission logistics, responsible sourcing — see Section 2.1)
- Enters the relevant declared figures for that activity type (e.g., kWh of renewable energy used, quantity of recycled material) and the claim period
- Uploads supporting evidence (utility bills, third-party audit reports, sensor/meter data exports, supplier certificates)
- Affirms the **exclusivity attestation** — an explicit confirmation that the reduction being claimed has not been, and will not be, registered with any other carbon registry. This is a recorded, timestamped affirmation tied to the submitting user, not a checkbox that disappears after submission.
- Submits the claim, which enters a review queue
- Can view claim status at any time: **Submitted → Under AI Review → Under Human Review → Approved / Rejected / More Information Requested**

**Submission limits:** at most **25 evidence documents** per claim, **25 MB** per file, **100 pages** per document. Accepted formats: PDF, PNG, JPEG, CSV, XLSX. These limits exist because evidence volume directly drives AI review cost (Section 7.2), and an unbounded upload is an unbounded bill.

**What happens behind the scenes (described at a product level):** Every submitted claim is first reviewed by an AI Agent that reads the evidence, extracts the figures it can find, cross-references them against the declared figures and against the reference ranges for that activity and region, checks for evidence previously submitted elsewhere on the platform, and produces a structured assessment with a confidence level and a set of specific flags. A human Verifier then reviews the claim — always seeing the AI's assessment and its citations alongside the original evidence, but the human always makes the final approval or rejection decision. **The AI never auto-approves a claim, and the AI never computes the credit amount** — the credit amount always comes from the deterministic formula in Section 2.1.

---

### 3.4 Verification & Credit Issuance (Verifier Workflow)

**What it does:** Gives Verifiers (internal team members with this role) a queue of pending Sustainability Claims to review, along with the AI assessment, and lets them approve, reject, or request more information.

**What a user (Verifier) does:**
- Views the queue of pending claims, sorted by priority then submission date
- Reviews the claim details, evidence, and the AI Agent's assessment, extracted figures, and citations
- Approves the claim (which triggers carbon credit issuance to the Organization's Treasury Address), rejects it (with a required reason of at least 40 characters), or requests more information from the submitting facility (pausing the claim and notifying the facility)
- Views a history of their own past review decisions

**Queue priority levels:**

| Priority | Trigger |
|---|---|
| `critical` | Claim would issue > 5,000 tCO2e, or the AI flagged a suspected duplicate evidence match |
| `high` | Requested amount exceeds the computed ceiling, or the facility is `unverified`, or the AI confidence is below 0.60 |
| `normal` | Everything else |

Within a priority level, Enterprise-plan claims sort ahead of others (reflecting the expedited-review commitment in Section 7.3), then oldest first.

**Product rules:**
- The exact credit amount issued is calculated using the formula for that activity type (Section 2.1), capped at whichever is lower: the computed ceiling, or the amount the Verifier explicitly approves. A Verifier can approve for a lower amount than requested, but the approval control physically cannot be set above the ceiling.
- **Claims issuing more than 5,000 tCO2e require two distinct Verifier approvals** before issuance. The second Verifier sees the first Verifier's decision and reasoning, and cannot be the same person. This threshold exists because a single compromised or mistaken Verifier account is otherwise the shortest path to fraudulent issuance in the entire system.
- A Verifier cannot review a claim from an Organization they have been flagged as related to; the relationship flag is set by an Admin.
- Once a decision is recorded it is final and cannot be edited. A wrong decision is corrected by an Admin-initiated reversal, which is a separate, logged, dual-authorized action — not an edit.

---

### 3.5 Carbon Credit Marketplace

**What it does:** Allows facilities holding carbon credits to list them for sale, and allows any registered Organization to browse, purchase, and retire credits.

**What a user does:**
- **Seller (organization holding credits):** Views their credit balance broken down by Credit Class (originating facility, vintage, activity type), and lists a quantity from **one Credit Class** for sale at a chosen price per tCO2e in USDC. Listing a credit moves it into escrow — it leaves the seller's spendable balance the moment the listing goes live, which is what makes overselling structurally impossible rather than merely checked for.
- **Buyer:** Browses available listings, filterable by activity type, vintage, region, originating facility, and seller Trust Tier. Each listing shows the **originating Facility** (not anonymized — buyers can see exactly which facility's verified sustainability practice a credit represents, which supports their own ESG reporting credibility) alongside the disclosure from Section 1.3. Purchases are settled in **USDC**, subject to a **minimum purchase quantity** set per listing by the seller.
- **Buyer:** At any point, can **retire** credits they hold — permanently marking them as "used" to claim an offset, with a required reason/purpose text of at least 20 characters (e.g., "2026 annual ESG report offset")
- **Anyone (public):** Can view a public retirement log — showing the retiring organization, the originating facility, quantity, vintage, activity type, reason, and date

**Marketplace rules with exact values:**

| Rule | Value |
|---|---|
| Minimum listing quantity | 1.000000 tCO2e |
| Minimum purchase quantity (per listing) | Set by seller, at least 0.100000 tCO2e |
| Minimum transaction notional | 1.00 USDC — a purchase computing to less is rejected |
| Price bounds | 0.50 – 5,000.00 USDC per tCO2e |
| Listing expiry | 90 days from creation, after which credits automatically return to the seller's balance |
| Remainder rule | A partial purchase must either leave a remainder of at least the listing's minimum purchase quantity, or consume the listing entirely |
| Platform fee | Charged to the **seller** at their Organization's plan rate (Section 7.3). Buyers pay no platform fee. |

The **remainder rule** exists because without it a listing can be left holding a quantity smaller than its own stated minimum, which nobody is permitted to buy and which therefore sits dead until it expires. This is the kind of rule that is obvious in hindsight and invisible until a listing gets stuck.

**Why retirement matters (for developer understanding):** Once retired, a credit is permanently removed from circulation and can never be resold, retraded, or retired again by anyone else. This is the core mechanism that prevents the same unit of environmental benefit from being claimed as an offset by two different companies. Read Section 1.3 for the precise boundary of what this does and does not guarantee.

---

### 3.6 Fraud & Anomaly Monitoring

**What it does:** Continuously monitors platform activity for suspicious patterns and surfaces them to the Admin/Ops team for review.

**Detection rules with exact thresholds:**

| Rule | Trigger condition | Initial severity |
|---|---|---|
| Impossible travel | Implied speed between two consecutive checkpoints exceeds 900 km/h (air freight ceiling), or exceeds 120 km/h where the declared mode is road/rail | `medium` |
| Backdated checkpoint | Event timestamp is more than 30 days before submission time | `low` |
| Future-dated checkpoint | Event timestamp is more than 4 hours after submission time | `medium` |
| Ceiling overreach, repeated | Facility submits 3 or more claims exceeding their computed ceiling within 90 days | `high` |
| Exact duplicate evidence | A file with an identical content hash was submitted on a different claim | `high` |
| Near-duplicate evidence | An evidence document scores ≥ 0.95 semantic similarity against a document from a different claim | `high` |
| Wash trading | The same two Organizations trade the same Credit Class in both directions within 7 days | `high` |
| Velocity anomaly | An Organization's checkpoint submission rate exceeds 10× its trailing 30-day daily median | `medium` |
| Claim clustering | 5 or more claims from the same Organization submitted within 1 hour | `low` |
| Instruction-like content in evidence | The AI review pipeline detects imperative text inside an evidence document addressed at the reviewer | `critical` |

The near-duplicate rule is worth highlighting for developers: exact content hashing only catches a byte-identical file. The same utility bill re-exported from a different PDF tool, or rescanned, produces a completely different hash while being the same document. Semantic similarity is what actually catches evidence reuse, which is why it exists as a separate rule rather than being folded into the hash check.

**What a user (Admin/Ops) does:**
- Views a fraud/anomaly queue, each entry showing what was flagged, why, the computed evidence, and a link to the relevant batch/claim/account
- Marks a flag as reviewed (false positive) or escalates it

**Escalation consequences are automatic and immediate.** When a flag is escalated, the affected Organization is placed in a `restricted` state: it can no longer submit Sustainability Claims, create Marketplace listings, or receive credit issuance. Existing listings are cancelled and escrowed credits returned. The Organization can still log in, view its data, log checkpoints, and export its data — restriction is not a lockout. Every one of these consequences is enforced at the point of action in every service, and separately at the ledger layer, rather than only being hidden in the UI.

---

### 3.7 Admin & Operations Portal (Brief Overview)

**What it does:** An internal-only tool for the Ops/Dev team to monitor platform health and take corrective action. This is intentionally minimal in scope for this version of the product.

**What an Admin user does:**
- Views the fraud/anomaly queue (Section 3.6) and escalates or dismisses flags
- Views system health indicators (checkpoint processing rate, review queue depth, event consumer lag, on-chain transaction backlog)
- Manages Verifier accounts and permissions, and sets Verifier–Organization relationship flags
- Manually sets an Organization's verification status, with a required justification
- Manages the reference factor tables (Section 2.6), with every change versioned by effective date
- In an emergency, can pause specific platform functions (credit issuance, marketplace trading, checkpoint anchoring) independently of each other
- **Manages subscription Plans** — pricing, quotas, rate limits, and feature availability per plan, adjustable without a code deployment, including per-Organization overrides (trial extensions, comped review quota, custom Enterprise terms)

This portal is for internal team use only and is not customer-facing. It is treated as a separate, smaller product surface from the rest of CarbonCircuit.

---

### 3.8 Organization Treasury Address

**What it does:** Establishes the single blockchain address that holds an Organization's carbon credits.

This is a distinct concept from an individual user's personal wallet, and conflating the two is the most consequential mistake this product could make. Credits belong to the **Organization**, not to whichever employee happened to connect a wallet first. If credits were minted to a personal address, an employee leaving the company would take the company's assets with them, and there would be no mechanism to recover them.

**How it works:**
- During onboarding, an Organization designates a Treasury Address. Ownership is proven by a signature from that address.
- All credit issuance is delivered to the Treasury Address. All marketplace listings sell from it. All retirements burn from it.
- Changing the Treasury Address requires: an Organization Owner role, a fresh signature from the new address, multi-factor re-authentication, and a **72-hour delay** during which the change is pending and every Owner and Admin on the Organization is notified by email. The change can be cancelled by any Owner during that window.
- An Organization with credits in escrow on active listings cannot change its Treasury Address until those listings are cancelled or filled.

The 72-hour delay exists for exactly one reason: an attacker who compromises a single user session should not be able to redirect an Organization's entire credit holding before anyone notices. A delay plus a broadcast notification turns a silent theft into a visible, cancellable event.

Individual users may still link a personal wallet to their own account — this is used only for signing actions on the Organization's behalf where they hold the required role, never for custody.

---

## 4. User Data Flow (For Developer Schema Understanding)

This section walks through how data is created and connected across the product lifecycle, so a developer can reason about what entities exist and how they relate to each other — without this document prescribing exact database schemas or technical implementation.

### 4.1 Core Entities and Their Relationships

- **Organization** — the top-level account (a Manufacturer, Assembler, Logistics Partner, or Credit Buyer company)
  - has a **verification status** and a **restriction state**
  - has exactly one **Treasury Address**
  - has one or more **Facilities** (only relevant for Manufacturers/Assemblers)
  - has one or more **Users** (individual people who log in on behalf of the org, each with a role)

- **Facility** — belongs to an Organization
  - has a **Trust Tier** and a **facility verification status**
  - has a **grid region** and **attested/declared capacity** figures
  - has zero or more **Sustainability Claims**
  - originates zero or more **Batches**

- **Batch** — belongs to a Facility (the originating one)
  - has a **public batch reference** distinct from its internal identifier
  - has one or more **Checkpoints**, each logged by an Organization involved in moving/handling it
  - may reference one or more **parent Batches** (component batches used to create it) — supports multi-level chains up to 10 deep
  - has a derived **Provenance Score** (Section 2.4), maintained as data changes rather than recomputed on every view

- **Sustainability Claim** — belongs to a Facility
  - has an **activity type**, a **claim period**, and **declared figures**
  - has a **computed ceiling** with the reference-table version pinned
  - has submitted supporting **Evidence** (documents/files)
  - has an **AI Review Result** (structured assessment, extracted figures with citations, confidence, flags — generated once, retained permanently)
  - has one or two **Verifier Decisions** depending on the dual-approval threshold
  - if approved, results in a **Carbon Credit Issuance** record

- **Carbon Credit Issuance** — tied to exactly one approved Sustainability Claim
  - defines a **Credit Class** (originating facility + vintage + activity type)
  - has an **amount issued** delivered to the Organization's Treasury Address
  - credits from this issuance can later be listed, traded, and retired, always retaining their Credit Class

- **Marketplace Listing** — created by an Organization selling credits of exactly one Credit Class
  - has a **quantity**, **price per tCO2e**, **minimum purchase quantity**, **expiry**, and **status** (active / partially filled / filled / cancelled / expired)
  - holds its credits in escrow while active
  - results in a **Trade** record on each fill

- **Trade** — connects a buyer Organization to a quantity of a specific Credit Class from a specific listing, with the settled USDC amount and platform fee recorded

- **Retirement** — a permanent record tied to a specific Organization, Credit Class, quantity, and stated purpose. Publicly viewable.

- **Fraud Flag** — can reference a Batch, Checkpoint, Sustainability Claim, Evidence document, or Trade, with a rule identifier, severity, computed evidence, and status (open / reviewed / escalated / resolved)

### 4.2 A Concrete End-to-End Example (Electronics Component)

A semiconductor Facility in Taiwan registers, matches the business registry dataset, and is marked `verified`; its facility record also matches, so its ceiling calculations use a discount factor of 1.00. It produces a Batch of 5,000 chips. As that batch moves — leaves the fab, clears customs, arrives at an Assembler — three Checkpoints are logged, each by whichever Organization handled that leg, and each is included in the next periodic on-chain anchor.

Separately, and on its own timeline, that same Facility submits a Sustainability Claim for vintage 2026 stating it used 12,400,000 kWh of verified renewable energy. The ceiling computation looks up the `TW` grid factor (0.494 kgCO2e/kWh), multiplies, divides by 1,000, and applies the 1.00 discount factor: a ceiling of 6,125.6 tCO2e. Because that exceeds 5,000, the claim enters the queue at `critical` priority and will require two Verifier approvals.

The AI Agent reads the submitted utility statements, extracts a renewable-supply figure of 12,180,000 kWh with a page citation, notes a 1.8% discrepancy against the declared figure, finds no duplicate evidence, and returns a `corroborated_with_discrepancy` assessment at 0.87 confidence. Two Verifiers review and approve at the extracted figure rather than the declared one — 6,016.9 tCO2e — and credits are issued to the Organization's Treasury Address as Credit Class `(Facility TW-01, vintage 2026, renewable_energy)`.

The batch's public Provenance Score now reflects that its originating facility has an approved sustainability claim on record. The Facility later lists half its credits on the marketplace; a Credit Buyer purchases them in USDC, and eventually retires them when filing their annual ESG report — at which point the public retirement log records that this specific buyer retired credits earned by Facility TW-01 in 2026 through renewable energy use.

---

## 5. Notifications

This section defines exactly when a user should receive a notification, and through which channel (in-app, email, or both). This is a product requirement, not a suggestion — every feature above should be cross-checked against this list.

| Event | Recipient | In-App | Email |
|---|---|---|---|
| Registration verification succeeded | Registering Organization | Yes | Yes |
| Registration verification failed or requires review | Registering Organization | Yes | Yes |
| Checkpoint logged for a batch you originated or handle | Facility / Logistics Partner | Yes | No |
| Checkpoint correction filed on a batch you originated | Originating Facility | Yes | No |
| Sustainability Claim submitted successfully | Submitting Facility | Yes | Yes |
| AI review completed on your claim | Submitting Facility | Yes | No |
| Verifier requests more information on your claim | Submitting Facility | Yes | Yes |
| Claim approved | Submitting Facility | Yes | Yes |
| Claim rejected | Submitting Facility | Yes | Yes |
| Carbon credits issued to your Treasury Address | Organization Owners | Yes | Yes |
| Credit issuance deferred (chain congestion or rate limit) | Organization Owners | Yes | No |
| New claim entering the review queue | Verifier(s) | Yes | No |
| Claim entering the queue at `critical` priority | Verifier(s) | Yes | Yes |
| Claim awaiting your second approval | Verifier(s) other than the first | Yes | Yes |
| Your marketplace listing sold (fully or partially) | Selling Organization | Yes | Yes |
| Your marketplace listing expired | Selling Organization | Yes | Yes |
| Your marketplace purchase completed | Buying Organization | Yes | Yes |
| Your marketplace purchase failed on-chain | Buying Organization | Yes | Yes |
| You successfully retired credits | Buying Organization | Yes | Yes |
| Treasury Address change requested | All Owners and Admins on the Organization | Yes | Yes |
| Treasury Address change completed or cancelled | All Owners and Admins on the Organization | Yes | Yes |
| API key created | All Owners on the Organization | Yes | Yes |
| API key revoked | All Owners on the Organization | Yes | Yes |
| Plan quota reached 80% | Organization Owners | Yes | No |
| Plan quota exhausted | Organization Owners | Yes | Yes |
| Overage charge incurred | Organization Owners | Yes | Yes |
| Payment failed | Organization Owners | Yes | Yes |
| A fraud flag is raised involving your account/batch/claim | Relevant Organization | Yes | No (until escalated) |
| A fraud flag involving your account is escalated (account restricted) | Relevant Organization | Yes | Yes |
| Account restriction lifted | Relevant Organization | Yes | Yes |
| New fraud flag entering the queue | Admin/Ops | Yes | No |
| Fraud flag raised at `critical` severity | Admin/Ops | Yes | Yes |
| Fraud flag escalated | Admin/Ops | Yes | Yes |
| Platform function paused (emergency) | Admin/Ops | Yes | Yes |
| On-chain contract paused | Admin/Ops | Yes | Yes |

**Guiding rule for developers adding new features later:** anything that changes account status, involves money/credits changing hands, or requires a time-sensitive human decision should be both in-app and email. Anything that's simply informational/status-tracking and doesn't require action can be in-app only.

**Digest behaviour:** in-app notifications of the same type for the same Organization are collapsed into a single entry when more than 10 arrive within an hour, so a bulk API ingest does not bury everything else in a user's notification centre. Email is never digested for the money-and-status events above; those always send individually.

---

## 6. Data Submission Methods

CarbonCircuit supports two distinct ways that batch and checkpoint data enters the system, since different users have different needs:

**1. Manual Portal Entry**
A user logs into the CarbonCircuit dashboard directly and fills out forms to register a batch, log a checkpoint, or submit a sustainability claim. This is the expected path for smaller facilities or logistics partners without their own internal systems, and is also how every user interacts with the marketplace and admin/verifier workflows (those are always portal-only, not API-driven).

**2. API-Based Integration**
Larger manufacturers or logistics partners who already run their own internal systems (ERP, warehouse management, etc.) can integrate directly with CarbonCircuit using an API key issued to their Organization. Their existing systems can automatically submit batch creation and checkpoint events to CarbonCircuit as they happen in their own workflows, without a human manually re-entering data in our portal.

**Every API-submitted batch and checkpoint carries a required `external_id`** — the identifier that record has in the partner's own system. This is unique per Organization and is what makes replay safe: a partner whose ERP retries a day of checkpoints after an outage submits the same `external_id` values, and the platform recognises them as the same records rather than creating duplicates. This is a stronger guarantee than a request header alone, because it survives a retry that regenerates its request metadata, and it gives the partner a stable key to reconcile against on their side.

**Referenced facilities must already exist.** If an API submission references a facility the Organization has not registered, the request is rejected with an error naming the unknown reference — the facility is **not** auto-created. A silently auto-created facility would be a self-declared record with a self-declared capacity feeding directly into the credit ceiling formula, created by a typo, with nobody having reviewed it. Facilities are provisioned deliberately, through the portal or through the explicit facility-creation endpoint, and only then referenced by ingestion.

Both submission methods always result in the exact same underlying batch/checkpoint records; a batch created via API looks and behaves identically to one created manually in the portal.

---

## 7. Business Model & Subscription Plans

CarbonCircuit is designed as a real product with a real business model, not just a feature demo — this section defines how the platform makes money and exactly what each Organization gets.

### 7.1 How CarbonCircuit Makes Money

Revenue comes from two combined sources:

1. **Subscription fees** — Organizations that *produce* data on the platform (Manufacturers, Assemblers, Logistics Partners) pay a recurring fee based on their Plan tier, which covers platform access, usage limits, and included AI-assisted claim reviews.
2. **Marketplace transaction fee** — CarbonCircuit takes a percentage fee on every completed carbon credit trade, charged to the **seller** at the seller's plan rate.

**Credit Buyers are free.** An Organization that only buys and retires credits pays no subscription. This is deliberate: charging a company $49/month for the privilege of spending money on the marketplace is the fastest way to have no buyers, and a marketplace with no buyers is worthless to the sellers who *are* paying. Buyer-side revenue comes entirely from the seller-paid transaction fee on trades that would not have happened otherwise.

### 7.2 A Note on AI Review Cost

The AI Agent used during Sustainability Claim review (Section 3.3) is **not a feature the user chooses to turn on or pay for directly** — it's an internal part of how CarbonCircuit processes every claim, always. Because each AI-assisted review has a real, variable cost to us, it's treated as a cost baked into subscription pricing (**included claim reviews per month, per Plan**), not as a separate line-item charge to the user. Usage beyond a Plan's included quota is billed as a metered overage rather than offered for free.

If a submitted claim is sent back by a Verifier for "more information" and resubmitted, that resubmission does **not** count against the Organization's monthly quota — only genuinely new claim submissions do. This avoids penalizing an Organization for a Verifier's request rather than their own mistake.

### 7.3 The Plans

There is one free tier, available only to Organizations that do not produce data. These figures are the finalized starting values, seeded directly into the database; Admins can adjust them without a redeploy.

| | **Buyer** | **Starter** | **Growth** | **Enterprise** |
|---|---|---|---|---|
| **Who it's for** | Companies that only purchase and retire credits | Small facilities or logistics partners just getting started | Mid-sized manufacturers/assemblers with regular volume | Large manufacturers, multi-facility organizations, or companies with compliance requirements |
| **Price** | Free | $49 / month | $199 / month | Custom (from $999/month, contract-negotiated) |
| **Organization types allowed** | Credit Buyer only | Any | Any | Any |
| **Batches per month** | — | 50 | 1,000 | Unlimited (50,000/month fair-use ceiling) |
| **Checkpoints per month** | — | 500 | 20,000 | Unlimited (1,000,000/month fair-use ceiling) |
| **Data submission method** | — | Manual portal entry only | Manual portal + API integration | Manual portal + API integration |
| **Facilities per Organization** | — | 1 | 5 | Unlimited (200 fair-use ceiling) |
| **Users per Organization** | 5 | 5 | 25 | Unlimited (500 fair-use ceiling) |
| **Included AI-reviewed Sustainability Claims / month** | — | 5 | 50 | 500 (custom quotas negotiable) |
| **Additional claim reviews beyond quota** | — | Not available — must upgrade | $5 per additional claim | $3 per additional claim, or custom terms |
| **Evidence storage** | — | 5 GB | 50 GB | 500 GB |
| **Portal rate limit** | 300 req/min | 300 req/min | 300 req/min | 600 req/min |
| **API rate limit** | — | — | 600 req/min (10 req/s) | 6,000 req/min (100 req/s) |
| **Bulk ingest** | — | — | 500 items per request, 10 requests/min | 1,000 items per request, 60 requests/min |
| **API keys** | — | — | 5 active | 25 active |
| **Marketplace access** | Buy and retire | Buy, sell, retire | Buy, sell, retire | Buy, sell, retire, plus priority listing visibility |
| **Marketplace transaction fee (charged to seller)** | n/a — buyers pay no fee | 3.0% | 2.5% | 1.5% (negotiated) |
| **Verifier review turnaround** | — | Standard queue | Standard queue | Expedited — 24 business hours |
| **Human-only review option (no AI pre-screening)** | — | Not available | Not available | Available, for compliance-sensitive Organizations |
| **Fraud/anomaly visibility** | Own account only | Own account only | Own account only | Own account, plus account-level summary reporting |
| **Data export** | Available (Section 10) | Available (Section 10) | Available (Section 10) | Available (Section 10), plus scheduled automated exports |
| **Support** | Community | Community/email | Priority email | Dedicated support contact |

**Fair-use ceilings** on Enterprise exist so that "unlimited" has a defined operational meaning. Crossing one does not block the Organization; it triggers an Admin notification and a commercial conversation. This is stated in the contract rather than enforced as a hard stop, because silently throttling a paying Enterprise customer mid-shipment is worse than a phone call.

**Overage behaviour:** an Organization exceeding a hard limit (batches, checkpoints, facilities, users, storage) receives a `PLAN_LIMIT_EXCEEDED` error naming the specific limit and linking to the upgrade flow. An Organization exceeding its AI review quota on Growth or Enterprise continues to be served and is billed the metered rate; on Starter, claim submission is blocked until the next billing period or an upgrade.

**Payment failure:** on a failed subscription payment, the Organization enters a 14-day grace period with full access and escalating notifications. After 14 days it moves to `read_only`: it can log in, view, and export everything, but cannot create batches, log checkpoints, submit claims, or create listings. Existing credits are never seized, and existing listings remain honourable — non-payment of a subscription is a billing problem, not grounds for interfering with an Organization's assets or its counterparties' pending trades.

### 7.4 What Gets Displayed to the User (Plan Comparison Page)

The Plan comparison page should clearly display, side by side: monthly price, batch and checkpoint volume limits, included AI-reviewed claims per month, data submission methods available, rate limits, marketplace fee rate, review turnaround expectations, and support level. Every Organization should also be able to see, from within their own dashboard at any time, their **current usage against their Plan's limits** (e.g., "32 of 50 claims reviewed this month") so there are no surprise overage charges.

### 7.5 Plan Management (Admin Capability)

Plan details — pricing, quotas, rate limits, and feature availability — are **not hardcoded**. They are managed by the Admin/Ops team through the Admin Portal (Section 3.7), so pricing and limits can be adjusted over time without requiring a new deployment. Admins can also apply manual overrides to an individual Organization's plan (trial extensions, goodwill credit grants, or fully custom Enterprise terms), and every override records the Admin's identity, a justification, and an optional expiry date.

---

## 8. Explicit Out-of-Scope Items (This Version)

To keep this PRD honest about what's actually being built, the following are intentionally **not** included in this version of the product:

- **Live third-party business registry integration** — organization and facility verification runs against a seeded reference dataset (Section 3.1). The verification logic, gating, and ceiling discount factors are fully real; only the data source is a fixture.
- **Real fiat currency payouts** — the marketplace transacts in USDC only. Converting USDC to bank-account currency would require a licensed money transmission partner and is out of scope.
- **Real satellite/IoT sensor integrations** — sustainability evidence relies on uploaded documents and reference datasets rather than live third-party data feeds.
- **Cross-registry double-issuance detection** — the platform cannot detect that a reduction was also registered with Verra or Gold Standard. Mitigations and disclosure are covered in Section 1.3.
- **Automated self-healing infrastructure** — platform health issues are surfaced to Admin/Ops for manual action, not automatically remediated.
- **A general "cheaper, no-AI" plan** — every claim is AI-assisted by default; the only manual-review-only option is the Enterprise-tier compliance accommodation in Section 7.3.

---

## 9. Multi-Tenancy

CarbonCircuit serves many different Organizations on one shared platform, and each Organization's data must be logically isolated from every other Organization's — this is a foundational requirement once the platform involves paid plans and competing businesses using the same system.

### 9.1 What Isolation Means Here

- An Organization can only see its own batches, claims, credit balances, marketplace listings, evidence, and billing/usage information.
- An Organization can never see another Organization's internal data (e.g., a Manufacturer cannot see an Assembler's other suppliers, or another customer's claim evidence).
- Isolation is enforced at the data layer, not only in application code — a query that omits its tenant scope must fail rather than return another tenant's rows.

### 9.2 Batch History Visibility

**Decision:** Any Organization involved in a batch's chain of custody (originating Facility, any Logistics Partner or Assembler that has handled it) can view that batch's **full checkpoint history**, not just the checkpoints they personally logged. This keeps the traceability story consistent for every party in the chain — a Logistics Partner handling leg 3 of a shipment can still see legs 1 and 2, which is useful for dispute resolution and general supply chain visibility.

What remains restricted is unrelated internal data belonging to another Organization: their other facilities, other unrelated batches, evidence documents, financial and billing details. "Full batch history" applies specifically to batches an Organization has actually touched, never to every batch on the platform.

**Parent batch visibility is one level only.** An Assembler can see that their component batch descends from a specific upstream batch and which facility originated it, but cannot walk further up the chain into a supplier's own supplier relationships. Exposing the full upstream tree would let any downstream party map a competitor's entire sourcing network from a single purchase — which is commercially sensitive information no participant agreed to share by joining a traceability platform.

### 9.3 Public Data

Three things are public without authentication: the batch provenance page reached by its public batch reference (Section 3.2), the retirement log (Section 3.5), and marketplace listings. Everything else requires authentication and tenant-scoped authorization. The public batch page shows checkpoint types, locations, timestamps, the originating facility's name and country, and the Provenance Score — it does not show quantities, prices, counterparty organization names, or evidence.

### 9.4 Plans and Tenancy

Each Organization (tenant) is on exactly one Plan (Section 7) at a time, and all usage limits, quotas, rate limits, and feature availability are enforced per Organization, not per individual user — multiple users from the same company share their Organization's single Plan and quota.

---

## 10. Account Deletion & Data Export

Since CarbonCircuit stores real business data and some of it is anchored on an immutable blockchain by design, "delete my account" needs a precise definition rather than a blanket promise.

### 10.1 Data Export

Any Organization can request a full export of their own data — batches, checkpoints they logged, sustainability claims, uploaded evidence, marketplace activity, retirement records, and account/billing information — in a structured, portable format (JSON manifest plus original evidence files, delivered as a single archive via a signed, time-limited download link valid for 7 days). This export is genuinely complete: it does not silently omit categories of data the Organization is entitled to see.

Export requests are rate-limited to one per Organization per 24 hours and are processed asynchronously with an email notification on completion.

### 10.2 Where Evidence Lives, and Why That Matters for Deletion

Uploaded evidence documents are stored in **private, encrypted object storage** and are never published to a public content-addressed network. Only a **cryptographic hash** of each document is ever recorded publicly or on-chain — enough for anyone to later verify that a specific document was the one submitted, without the document itself being retrievable by anyone who lacks authorization.

This is a product requirement, not an infrastructure preference. Evidence documents are confidential business records: utility statements, audit reports, supplier contracts. Publishing them to a network from which content cannot be withdrawn would make the deletion commitment below impossible to honour, and would leak commercially sensitive data permanently on the strength of a leaked identifier.

### 10.3 Account Deletion

When an Organization requests account deletion, the following distinctions apply:

- **Off-chain personal and business data** — contact details, uploaded evidence documents, login credentials, account profile information — is deleted or fully anonymized within 30 days of the request. Because evidence lives in private storage (Section 10.2), this deletion is genuinely achievable.
- **On-chain records** — carbon credit issuances, trades, and retirements — **cannot be deleted**. They exist on an immutable ledger by design, since that immutability is what makes the carbon credit system trustworthy in the first place. These records remain, but are no longer linked to identifiable off-chain organization details; the organization's display name is replaced with a generic identifier such as "Deleted Organization #[ID]" in every future-facing view, including the public retirement log.
- **Checkpoints and batch data involving other Organizations** cannot simply be erased either, since doing so would break another Organization's provenance chain that depends on that checkpoint's existence. The deleted Organization's identifying details are anonymized going forward, while the checkpoint event itself — its existence, type, timestamp, and role in the chain — persists to preserve the integrity of other parties' data.
- **Credits still held** must be sold or retired before deletion completes. An Organization cannot delete its way out of holding assets, and the platform will not seize them; the deletion flow blocks with a clear explanation and links to the marketplace and retirement flows.

This nuance — that "delete" means different things for off-chain personal data, on-chain financial records, and shared multi-party supply chain data — is stated explicitly in the deletion request flow, with the specific consequences for that Organization's actual holdings shown before they confirm.

---

## 11. Open Questions / Things to Revisit

No open items at this time. This section is kept as a placeholder for questions arising as development progresses.

---

*This document describes product behavior and intent. Technical architecture, data models, and implementation details are covered separately.*
