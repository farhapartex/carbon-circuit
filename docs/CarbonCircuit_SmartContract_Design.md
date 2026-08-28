# CarbonCircuit — Smart Contract Design

**Scope:** On-chain system design for batch provenance anchoring, carbon credit issuance, and the credit marketplace. Solidity, interface-first, modular contracts backed by shared libraries. This document defines interfaces, contract responsibilities, storage and upgrade strategy, access control, and security architecture, precisely enough to implement directly. Function signatures are specified with full types where they define a contract, but this remains system design — no implementation bodies.

**Target chain:** Base (Ethereum L2, OP Stack). Base Sepolia for testnet; a local Anvil node forked from Base for development. Settlement currency is Circle-issued native USDC on Base (6 decimals).

---

## 1. Design Principles

Every contract in this system is built against an interface first — the interface is the contract, the implementation is one (potentially swappable) realization of it. Utility logic that would otherwise be duplicated across contracts lives in shared libraries, never copy-pasted. Storage layout is verified by tooling after every change, never assumed from reading source top to bottom. Security controls (pause, role-gating, reentrancy protection, freeze) are applied uniformly across every contract that touches value or state that matters. Code is organized for readability first: one concern per contract, consistent internal ordering, and NatSpec documentation on every public surface.

Two principles specific to this system are worth stating explicitly, because they shape decisions throughout:

**Authorization to move value never originates from the address submitting the transaction.** The backend's transaction-signing component is a pure executor. Every mint carries an independent, off-chain-produced authorization that the contract verifies for itself. This is what makes on-chain enforcement a genuine second line of defense rather than a check controlled by the same key it is meant to constrain.

**On-chain cost must not scale with customer count.** Anchoring is one transaction per epoch platform-wide, not one per batch; a batch NFT is minted only when someone actually needs an on-chain batch identity, not automatically. Any design where adding a customer adds a recurring transaction is a design that stops working at the scale this platform targets.

---

## 2. Contract Inventory

Each upgradeable entry below is deployed as **two on-chain artifacts, never one**: an `ERC1967Proxy` (holds all storage, is the address every other contract and the frontend/backend ever refer to, never changes) and an implementation contract (holds all logic, holds no persistent data of its own, freely replaceable). This split is the entire point of Section 8 and is treated as a first-class part of the inventory.

| Logical Contract | Interface | Proxy Artifact | Implementation Artifact | Responsibility |
|---|---|---|---|---|
| `CarbonCreditToken` | `ICarbonCreditToken` | `CarbonCreditTokenProxy` (ERC1967Proxy) | `CarbonCreditTokenV1` | ERC-1155 multi-token carbon credit. One token ID per Credit Class. Verifier-authorized minting, per-claim consumption, rolling-window rate limits, retirement, compliance freezing |
| `ProvenanceAnchor` | `IProvenanceAnchor` | `ProvenanceAnchorProxy` (ERC1967Proxy) | `ProvenanceAnchorV1` | Global per-epoch Merkle roots covering all batch and checkpoint events platform-wide, with permanent inclusion verification against any historical epoch |
| `BatchRegistry` | `IBatchRegistry` | `BatchRegistryProxy` (ERC1967Proxy) | `BatchRegistryV1` | Optional ERC-721 on-chain identity for individual batches, minted on demand rather than automatically |
| `Marketplace` | `IMarketplace` | `MarketplaceProxy` (ERC1967Proxy) | `MarketplaceV1` | Listings with credit escrow, USDC settlement via pull payments, retirement initiation, platform fee collection |
| `CreditTokenFactory` | `ICreditTokenFactory` | — (no proxy) | `CreditTokenFactory` | Deterministic CREATE2 deployment of the canonical credit token proxy and implementation, with a one-shot safeguard |

| Library | Type | Responsibility |
|---|---|---|
| `CreditClassLib` | Internal, stateless | Deriving and decomposing the ERC-1155 token ID that encodes a Credit Class |
| `MerkleLib` | Internal, stateless | Domain-separated Merkle inclusion proof verification for anchored epoch roots |
| `ValidationLib` | Internal, stateless | Shared input-validation checks reused across every contract |

`CarbonCreditToken`, `ProvenanceAnchor`, `BatchRegistry`, and `Marketplace` are upgradeable (UUPS, Section 8) since they hold long-lived financial and provenance state that may need a bug fix or an additive feature without a migration and without losing existing data. `CreditTokenFactory` is deliberately **not** upgradeable and has no proxy — a factory's job is a one-time deterministic deployment, and giving it upgrade logic would add attack surface with no corresponding benefit.

---

## 3. The Credit Class Token Model

This section comes before the interfaces because every interface below depends on understanding it.

### 3.1 Why ERC-1155 and Not ERC-20

A carbon credit on CarbonCircuit is defined by three inseparable attributes — the **originating facility**, the **vintage year**, and the **activity type**. Together these form its Credit Class, and a credit that has lost them is not the product this platform sells: a buyer's entire reason for paying a premium is being able to name, in their own ESG report, the specific verified practice they funded.

A single fungible ERC-20 balance cannot carry those attributes. Two facilities' 2026 renewable-energy credits would become the same undifferentiated number the moment they landed in the same address, and no transfer could preserve what it never encoded. Keeping a parallel per-attribute mapping alongside an ERC-20 balance does not solve this either — a standard `transfer(to, amount)` carries no indication of which attributes are moving, so the two ledgers desynchronize on the first transfer anyone makes.

ERC-1155 encodes exactly this shape natively. Each Credit Class is a distinct token ID with its own balance and its own transfer semantics, so attribution is a property of the token rather than bookkeeping alongside it. Batch transfer and batch balance queries come for free, which matters for a portfolio view spanning many classes. There is no separate mapping to keep in step, because there is no separate mapping.

### 3.2 Token ID Derivation

```
tokenId = uint256(keccak256(abi.encode(CREDIT_CLASS_DOMAIN, facilityId, vintageYear, activityType)))
```

| Component | Type | Notes |
|---|---|---|
| `CREDIT_CLASS_DOMAIN` | `bytes32` constant | Domain separator, so a token ID can never collide with a hash computed for another purpose in this system |
| `facilityId` | `bytes32` | The platform's facility identifier |
| `vintageYear` | `uint16` | Calendar year the reduction occurred |
| `activityType` | `uint8` | `1` renewable energy, `2` reduced-emission logistics, `3` responsible sourcing |

Because the ID is a hash, the contract cannot recover the components from it. It therefore stores the decomposition once, on first mint of a class, in `mapping(uint256 => CreditClass) creditClasses` — so any on-chain or off-chain reader can resolve a token ID back to its facility, vintage, and activity type without consulting off-chain data. `CreditClassLib` provides both directions, and every caller derives IDs through it so that no two call sites can construct the same class differently.

Credit amounts use **18 decimals**, where `1e18` represents one tCO2e. ERC-1155 has no decimals concept of its own, so this is a stated convention enforced consistently in the contracts, the backend, and the frontend rather than a property the standard provides.

---

## 4. Interfaces

### 4.1 `ICarbonCreditToken`

| Function | Visibility | Parameters | Returns | Purpose |
|---|---|---|---|---|
| `mint` | external | `MintAuthorization calldata auth, bytes[] calldata signatures` | `uint256 tokenId` | Mints credits against a single previously unused claim, authorized by verifier signatures the contract verifies itself |
| `retire` | external | `uint256 tokenId, uint256 amount, string calldata reason` | — | Burns the caller's own credits permanently, emitting a public retirement record |
| `retireFor` | external | `address account, uint256 tokenId, uint256 amount, string calldata reason` | — | Burns credits held by the Marketplace on behalf of `account`, attributing the retirement to `account`. Callable only by the registered Marketplace address |
| `getCreditClass` | external view | `uint256 tokenId` | `CreditClass memory` | Facility, vintage, and activity type behind a token ID |
| `isClaimConsumed` | external view | `bytes32 claimId` | `bool` | Whether a claim has already been used to mint |
| `mintedInWindow` | external view | `address account` | `uint256 facilityAmount, uint256 globalAmount` | Amounts minted in the current rolling window, for both the facility and the platform |
| `setVerifierAuthority` | external | `address authority, bool authorized` | — | Adds or removes an address whose signature authorizes a mint. `DEFAULT_ADMIN_ROLE`, timelocked |
| `setDualApprovalThreshold` | external | `uint256 amount` | — | Amount above which two distinct verifier signatures are required. `DEFAULT_ADMIN_ROLE`, timelocked |
| `setMintLimits` | external | `uint256 perTx, uint256 perFacilityWindow, uint256 perGlobalWindow` | — | Adjusts rate limits within the compiled-in absolute ceiling. `DEFAULT_ADMIN_ROLE`, timelocked |
| `setFrozen` | external | `address account, bool frozen` | — | Freezes or unfreezes an account's ability to transfer, receive, or retire. `COMPLIANCE_ROLE` |
| `pause` / `unpause` | external | — | — | Circuit breaker, role-gated asymmetrically (Section 7.1) |

**`MintAuthorization` struct** — the EIP-712 typed payload signed by verifier authorities:

| Field | Type | Purpose |
|---|---|---|
| `claimId` | `bytes32` | The approved claim; consumed exactly once |
| `to` | `address` | Recipient — the Organization's Treasury Address |
| `facilityId` | `bytes32` | Credit Class component |
| `vintageYear` | `uint16` | Credit Class component |
| `activityType` | `uint8` | Credit Class component |
| `amount` | `uint256` | Exact amount to mint, in 18-decimal tCO2e |
| `deadline` | `uint256` | Unix timestamp after which the authorization is void |
| `nonce` | `bytes32` | Prevents replay of an otherwise identical authorization |

### 4.2 `IProvenanceAnchor`

| Function | Visibility | Parameters | Returns | Purpose |
|---|---|---|---|---|
| `anchorEpoch` | external | `uint64 epoch, bytes32 root, uint64 leafCount` | — | Records the Merkle root covering every provenance event in that epoch, platform-wide. Reverts if the epoch is already anchored or is not the next expected epoch |
| `verifyInclusion` | external view | `uint64 epoch, bytes32 leaf, bytes32[] calldata proof` | `bool` | Verifies a leaf was included in that epoch's anchored root |
| `getEpochRoot` | external view | `uint64 epoch` | `bytes32 root, uint64 leafCount, uint64 anchoredAt` | The anchored root for an epoch, or zero if not yet anchored |
| `latestEpoch` | external view | — | `uint64` | Most recently anchored epoch |
| `pause` / `unpause` | external | — | — | Circuit breaker, role-gated |

### 4.3 `IBatchRegistry`

| Function | Visibility | Parameters | Returns | Purpose |
|---|---|---|---|---|
| `registerBatch` | external | `address owner, bytes32 batchRef, bytes32 productCategoryId, string calldata metadataURI` | `uint256 tokenId` | Mints an on-chain identity for one batch, on demand. Reverts if that `batchRef` is already registered |
| `setMetadataURI` | external | `uint256 tokenId, string calldata metadataURI` | — | Updates the metadata pointer while unfrozen |
| `freezeMetadata` | external | `uint256 tokenId` | — | Permanently locks the metadata URI, after which it can never change |
| `tokenIdForBatch` | external view | `bytes32 batchRef` | `uint256` | Resolves a platform batch reference to its token ID, or zero |
| `pause` / `unpause` | external | — | — | Circuit breaker, role-gated |

### 4.4 `IMarketplace`

| Function | Visibility | Parameters | Returns | Purpose |
|---|---|---|---|---|
| `createListing` | external | `uint256 tokenId, uint256 amount, uint256 pricePerUnit, uint256 minPurchaseQuantity, uint64 expiresAt` | `uint256 listingId` | Escrows the seller's credits and opens a listing |
| `purchase` | external | `uint256 listingId, uint256 quantity, uint256 maxPricePerUnit, uint64 deadline` | — | Buyer pays USDC, receives credits, seller proceeds credited for withdrawal, platform fee taken |
| `purchaseAndRetire` | external | `uint256 listingId, uint256 quantity, uint256 maxPricePerUnit, uint64 deadline, string calldata reason` | — | Purchase and immediately retire in one transaction, attributed to the buyer |
| `cancelListing` | external | `uint256 listingId` | — | Seller reclaims unsold escrowed credits |
| `sweepExpiredListing` | external | `uint256 listingId` | — | Returns escrowed credits from an expired listing to its seller. Callable by anyone, so expiry cannot strand credits if a seller goes inactive |
| `withdrawProceeds` | external | — | `uint256 amount` | Seller withdraws accumulated USDC proceeds |
| `withdrawFees` | external | `address to` | `uint256 amount` | Treasury withdraws accumulated platform fees. `DEFAULT_ADMIN_ROLE` |
| `setSellerFee` | external | `address seller, uint16 basisPoints` | — | Sets a seller-specific fee reflecting their plan tier. `FEE_MANAGER_ROLE` |
| `clearSellerFee` | external | `address seller` | — | Removes an override, returning the seller to the default rate. `FEE_MANAGER_ROLE` |
| `setDefaultFee` | external | `uint16 basisPoints` | — | Sets the platform default fee. `DEFAULT_ADMIN_ROLE`, timelocked |
| `quote` | external view | `uint256 listingId, uint256 quantity` | `uint256 cost, uint256 fee, uint256 sellerProceeds` | Exact settlement figures for a prospective purchase |
| `pause` / `unpause` | external | — | — | Circuit breaker, role-gated |

### 4.5 `ICreditTokenFactory`

| Function | Visibility | Parameters | Returns | Purpose |
|---|---|---|---|---|
| `deployCreditToken` | external | `bytes32 salt, address admin, address minter, address timelock` | `address proxyAddress` | Deterministically deploys the implementation and its proxy via CREATE2, initializing the proxy in the same transaction. Reverts if this factory has already deployed |
| `getDeployedToken` | external view | — | `address proxy, address implementation` | The canonical deployed addresses, or zero addresses if none yet |
| `computeProxyAddress` | external view | `bytes32 salt` | `address` | Predicts the proxy address before deployment, for off-chain pre-configuration and post-hoc verification |

---

## 5. Libraries

### 5.1 `CreditClassLib`

| Function | Purpose |
|---|---|
| `toTokenId(bytes32 facilityId, uint16 vintageYear, uint8 activityType)` | Derives the domain-separated ERC-1155 token ID for a Credit Class. Every call site that needs an ID uses this, so two sites can never derive the same class differently |
| `validateActivityType(uint8 activityType)` | Reverts on an unrecognised activity type, so an invalid class can never be created by a typo in a parameter |

### 5.2 `MerkleLib`

| Function | Purpose |
|---|---|
| `hashLeaf(uint64 epoch, bytes32 batchRef, bytes32 eventDataHash)` | Canonical leaf hashing, prefixed with a leaf domain byte. This consistency is what makes off-chain-computed proofs verifiable on-chain |
| `hashNode(bytes32 left, bytes32 right)` | Internal node hashing, prefixed with a distinct node domain byte |
| `verify(bytes32 root, bytes32 leaf, bytes32[] calldata proof)` | Standard inclusion proof verification using the domain-separated hashing above |

The leaf and node domain prefixes are not a stylistic choice. Without them, an internal node of the tree is a valid 32-byte value that can be presented as if it were a leaf, letting an attacker prove inclusion of data that was never anchored — the classic second-preimage attack on Merkle trees. Distinct prefixes make the two hash spaces disjoint and the attack impossible.

### 5.3 `ValidationLib`

| Function | Purpose |
|---|---|
| `requireNonZeroAddress(address value)` | Reverts with a shared custom error if the address is zero |
| `requireNonZeroAmount(uint256 value)` | Reverts if an amount parameter is zero |
| `requireInRange(uint256 value, uint256 min, uint256 max)` | Bounds-checks a value, used for fee basis points, prices, and rate limits |
| `requireNotExpired(uint64 deadline)` | Reverts if the current block timestamp is past a caller-supplied deadline |

Every contract imports these via `using X for Y` rather than reimplementing the same checks — a validation rule defined once here is the only place it is ever defined.

---

## 6. Access Control & Roles

Every upgradeable contract inherits OpenZeppelin's `AccessControlUpgradeable` directly — a bespoke role registry would add audit surface without improving on a well-audited standard. The same role identifiers are used consistently across every contract that needs them.

| Role | Held by | Capability |
|---|---|---|
| `DEFAULT_ADMIN_ROLE` | 3-of-5 multisig, routed through `TimelockController` (48h delay) | Grants and revokes all other roles, manages verifier authorities, sets rate limits and default fee |
| `UPGRADER_ROLE` | Same multisig, same timelock | Authorizes a UUPS upgrade |
| `MINTER_ROLE` | Chain Writer Service `minter` key | Submits `mint` transactions. Cannot authorize one — see below |
| `ANCHOR_ROLE` | Chain Writer Service `anchor` key | Calls `anchorEpoch` on `ProvenanceAnchor` |
| `REGISTRAR_ROLE` | Chain Writer Service `ops` key | Calls `registerBatch` and metadata functions on `BatchRegistry` |
| `PAUSER_ROLE` | Multisig, plus a separate fast-response guardian key | Can pause any contract; cannot unpause |
| `COMPLIANCE_ROLE` | Chain Writer Service `ops` key | Freezes and unfreezes accounts following an Admin escalation |
| `FEE_MANAGER_ROLE` | Chain Writer Service `ops` key | Sets and clears per-seller fee overrides on behalf of Billing Service |
| *Verifier authorities* | Verifier signing keys, registered individually | Not an `AccessControl` role. An allowlisted set of addresses whose EIP-712 signatures authorize a mint |

Two aspects of this matrix carry most of its security value.

**Every on-chain role that submits a transaction is held by a Chain Writer Service key, and by nothing else.** `FEE_MANAGER_ROLE` and `COMPLIANCE_ROLE` are exercised on behalf of Billing Service and Admin Backend respectively, but those services never hold a key — they call Chain Writer Service, which is the single component in the entire architecture with signing capability. Distributing key material to a second service to save one internal call would trade the system's clearest security boundary for a marginal convenience.

**`MINTER_ROLE` grants the ability to submit a mint, never the ability to authorize one.** Authorization comes from verifier-authority signatures over the exact mint parameters, verified by the contract itself. An attacker holding the minter key can submit only mints they already possess a valid, unexpired, unconsumed authorization for — which means the practical value of compromising it is close to zero. This separation is the reason the per-claim cap is meaningful: a design where the same key both registers the cap and mints against it provides no protection at all against that key's compromise, because the attacker simply registers whatever cap they want first.

---

## 7. Contract Design Detail

### 7.1 `CarbonCreditToken`

**State:** ERC-1155 balances (standard), `mapping(uint256 => CreditClass) creditClasses`, `mapping(bytes32 => bool) claimConsumed`, `mapping(bytes32 => bool) authorizationNonceUsed`, `mapping(address => bool) verifierAuthorities`, `mapping(address => bool) frozen`, `mapping(uint256 => uint256) totalRetiredByClass`, a `totalRetired` counter, rolling-window mint accounting, and the configurable limits.

**Mint flow.** A single atomic call performs, in order:

1. Verify the contract is not paused and neither `auth.to` nor the caller is frozen.
2. Verify `auth.deadline` has not passed and `auth.nonce` is unused.
3. Recover signers from `signatures` over the EIP-712 hash of `auth`. Require every recovered signer to be a registered verifier authority, require them to be distinct, and require **two** of them when `auth.amount` exceeds the dual-approval threshold.
4. Verify `claimConsumed[auth.claimId]` is false.
5. Verify `auth.amount` does not exceed the per-transaction limit, the facility's rolling-window remaining allowance, or the platform's rolling-window remaining allowance.
6. Mark the claim consumed and the nonce used; advance both rolling-window counters.
7. Derive the token ID via `CreditClassLib`, storing the class decomposition if this is its first mint.
8. Mint to `auth.to` and emit `CreditMinted`.

Every check precedes every effect, and there are no external calls in the flow at all, so reentrancy is not merely guarded against but structurally absent. Because claim consumption and the balance increase happen in one function call, atomicity is native rather than something requiring coordination with the off-chain distributed lock — that lock protects the off-chain mirror, not this contract.

**Rolling-window rate limiting.** A true continuously-rolling window is unimplementable on-chain at acceptable cost, so the limit uses 24 hourly buckets per subject: each mint credits the current hour's bucket, and the window total is the sum of the 24 most recent buckets, with buckets older than 24 hours treated as zero rather than being explicitly cleared. This gives rolling-window behaviour with bounded, predictable gas.

Limits are configurable within a **compiled-in absolute ceiling** that no administrative action can raise. The compiled ceiling exists because a rate limit whose maximum a compromised admin can lift is not a rate limit; the configurable range beneath it exists because the correct operating value will change as the platform grows.

**Retire flow.** `retire` burns from the caller's own balance. `retireFor` exists solely so that a combined purchase-and-retire can attribute the retirement to the buyer rather than to the Marketplace contract that momentarily custodied the credits — without it, the public retirement log would record the Marketplace as having retired everything, which would make the transparency feature worthless. `retireFor` is callable only by the registered Marketplace address, only for credits that address currently holds, and emits `CreditRetired` naming the buyer as the retiring party.

`CreditRetired` carries the retiring account, the token ID, the amount, the reason, and the decomposed Credit Class. This event is the entire basis for the public retirement log, read directly off-chain rather than duplicated into any other on-chain record.

**Freezing.** A frozen account cannot send, receive, or retire credits. The check sits in the ERC-1155 transfer hook, so it applies uniformly to every transfer path including batch transfers, with no possibility of a path that forgets it. Freezing is a targeted alternative to the blunt global pause: a single Organization under fraud escalation is contained without halting the marketplace for everyone else.

### 7.2 `ProvenanceAnchor`

**State:** `mapping(uint64 => EpochAnchor) epochs` where `EpochAnchor` holds the root, leaf count, and anchoring timestamp; and `latestEpoch`.

**Anchoring flow.** Chain Observer Service accumulates every provenance event across the entire platform during a 10-minute epoch, builds one Merkle tree over them using `MerkleLib.hashLeaf`, and submits a single `anchorEpoch` transaction. Epochs must be anchored strictly in sequence, so a gap is a detectable, alertable condition rather than a silent hole in the record.

Two properties follow from this design and both matter.

**Cost is independent of platform size.** One transaction per epoch covers every batch and every checkpoint from every customer. A design anchoring per batch would tie recurring on-chain cost directly to customer count, which is exactly the scaling characteristic an L2 batching design exists to avoid.

**Every anchor is permanently verifiable.** Each epoch's root is retained in storage indefinitely, so a checkpoint anchored in epoch 400 remains provable years later. Retaining only the most recent root would make historical proofs impossible the moment the next epoch was anchored — the inclusion proof would have nothing left to verify against — which would quietly destroy the property the anchoring exists to provide. Storage cost is one slot per epoch, roughly 52,000 slots per year, which is a negligible price for permanent verifiability.

### 7.3 `BatchRegistry`

**State:** standard ERC-721 storage, `mapping(bytes32 => uint256) batchRefToTokenId`, `mapping(uint256 => string) metadataURI`, `mapping(uint256 => bool) metadataFrozen`.

Batch NFTs are **minted on demand, never automatically.** The platform's provenance guarantee comes from epoch anchoring, which already covers every batch at no per-batch cost. An NFT adds something anchoring does not — a transferable, individually addressable on-chain identity — and that is genuinely useful for a finished-good batch a brand wants to reference publicly or attach to a product passport. It is not useful for the vast majority of intermediate component batches, and minting one for each would put a transaction on every batch creation for no benefit to anyone.

Metadata points at the platform's public batch page rather than at customer evidence. Evidence documents are confidential and are never published; only their hashes appear in anchored leaves. `freezeMetadata` gives a batch owner the option of making their metadata permanently immutable once the batch's journey is complete, which is what a downstream party auditing a product passport actually needs.

### 7.4 `Marketplace`

**State:** a `Listing` struct per `listingId` (seller, tokenId, remaining amount, price per unit, minimum purchase quantity, expiry, status), `mapping(address => FeeOverride) sellerFees`, a default fee, `mapping(address => uint256) pendingProceeds`, `accumulatedFees`, and references to `CarbonCreditToken` and USDC.

**Listing.** `createListing` transfers the seller's credits into the Marketplace's own custody. The seller does not retain a spendable balance of listed credits anywhere, which makes overselling structurally impossible rather than merely checked for. Bounds enforced at creation: amount at least 1 tCO2e, minimum purchase quantity at least 0.1 tCO2e and no greater than the listed amount, price between 0.50 and 5,000 USDC per tCO2e, expiry between 1 and 90 days out.

**Settlement arithmetic**, stated exactly because rounding in a financial contract is a correctness property:

```
cost            = ceil(quantity × pricePerUnit / 1e18)      [USDC base units, 6 decimals]
fee             = floor(cost × feeBasisPoints / 10_000)
sellerProceeds  = cost − fee
```

Cost rounds **up** so a dust quantity can never settle for zero, and the result must additionally be at least `MIN_NOTIONAL` (1.00 USDC) — together these close the free-credits path that truncation alone would leave open. Fee rounds **down**, so the rounding remainder always favours the seller rather than the platform. `quote` returns all three figures so the frontend displays exactly what will settle, with no client-side reimplementation of this arithmetic to drift out of step.

**Purchase.** `purchase` validates the deadline, validates `pricePerUnit` against the caller's `maxPricePerUnit`, validates quantity against the listing's minimum, enforces the remainder rule, updates listing state, then pulls USDC from the buyer and transfers credits from escrow.

The `maxPricePerUnit` and `deadline` parameters are not ceremony. A transaction sitting in the mempool can be observed, and without a price bound and an expiry the buyer is signing an open-ended commitment against a listing whose state may change before inclusion. Both parameters convert that into a bounded one, which is standard practice for any on-chain purchase and is the kind of thing that is trivial to add up front and impossible to add later without a new interface version.

The **remainder rule** requires a partial purchase to either leave at least `minPurchaseQuantity` remaining or consume the listing entirely. Without it, a 100-unit listing with a 10-unit minimum, after a 95-unit purchase, holds 5 units that no purchase is permitted to buy — dead stock until expiry.

**Pull payments.** Seller proceeds accrue to `pendingProceeds` and are withdrawn separately rather than pushed during settlement. USDC maintains a blacklist, and a push payment to a blacklisted seller would revert the entire purchase — meaning one seller's compliance status could block unrelated buyers' transactions. Accrual-and-withdraw isolates that failure to the party it concerns and removes the external-call surface from the settlement path.

**Reentrancy.** Every value-moving function is guarded by `ReentrancyGuardUpgradeable` and follows checks-effects-interactions explicitly: listing state is updated before any token transfer, never after. Every ERC-20 interaction uses `SafeERC20`.

### 7.5 Fee Override Semantics

`FeeOverride` is a struct carrying both a `uint16 basisPoints` and a `bool isSet`, rather than a bare `uint16`.

This is deliberate and load-bearing. A bare mapping defaults to zero, which makes "this seller has no override, apply the default" indistinguishable from "this seller has an explicitly negotiated 0% fee" — so either every seller who has never been touched by the fee manager pays nothing, or a legitimate zero-fee arrangement is silently ignored. Neither is acceptable, and the ambiguity is invisible in testing until the first Enterprise contract negotiates a zero rate.

Fees are bounded at 0–1,000 basis points (0–10%) by `ValidationLib`, so a compromised or mistaken fee manager cannot set a confiscatory rate.

### 7.6 `CreditTokenFactory`

**State:** `address deployedProxy`, `address deployedImplementation`, both set once.

`deployCreditToken` reverts if a token has already been deployed by this factory — a one-shot safeguard, since accidentally deploying a second canonical credit token would fragment balances across two contracts with no way to merge them. It deploys the implementation, then deploys the proxy via CREATE2 with the `initialize` call encoded into the proxy's own construction, so the contract is never left uninitialized even momentarily.

CREATE2 is used so the address is computable in advance: Chain Writer Service and Credit Ledger Service can be configured with the correct token address before deployment completes, and the deployment can be independently verified afterwards by recomputing the expected address from the same salt and comparing.

---

## 8. Security Architecture

### 8.1 Circuit Breakers

Every contract inherits `PausableUpgradeable`, and every state-changing external function checks `whenNotPaused`.

Pausing is deliberately easier to trigger than unpausing: `PAUSER_ROLE` is granted to the multisig and separately to a fast-response guardian key, while unpausing requires `DEFAULT_ADMIN_ROLE`. The asymmetry is intentional — a false-positive pause costs availability, a false-negative failure to pause during a live incident costs money.

That asymmetry creates its own risk, and the design accounts for it. A compromised guardian key could otherwise halt the marketplace indefinitely, since recovery requires convening multisig signers. **A guardian-initiated pause therefore expires automatically after 72 hours** unless the multisig ratifies it, converting an indefinite denial-of-service into a bounded one that self-heals. A multisig-initiated pause has no expiry.

Contracts pause independently, so credit issuance can be halted without stopping marketplace trading or provenance anchoring. Freezing individual accounts (Section 7.1) is the preferred response to a single-Organization problem; the global pause is reserved for a genuine systemic incident.

### 8.2 Access Control Discipline

Roles are granted with least privilege and kept separate — no single role bundles minting, pausing, and upgrading, so a single compromised key has bounded blast radius. `MINTER_ROLE` and `ANCHOR_ROLE` are granted only to Chain Writer Service's dedicated addresses, never to the admin multisig: the multisig can pause the system or replace which address holds a role, but cannot mint by virtue of holding `DEFAULT_ADMIN_ROLE`.

Verifier authorities are managed as a separate allowlist from the role system, because they represent a different kind of authority — the right to *authorize* value creation rather than the right to *execute* an operation. Adding or removing one goes through the timelock.

### 8.3 Multisig and Timelock

`DEFAULT_ADMIN_ROLE` and `UPGRADER_ROLE` are held by a 3-of-5 multisig, and every action they take is routed through OpenZeppelin's `TimelockController`:

| Action class | Delay |
|---|---|
| Contract upgrade | 48 hours |
| Role grant or revocation | 48 hours |
| Verifier authority change | 48 hours |
| Rate limit adjustment | 24 hours |
| Default fee change | 24 hours |
| Unpause | 6 hours |

A compromised single signer cannot unilaterally change roles or push an upgrade — both require multiple signers, and even after agreement there is a publicly visible window on-chain during which the pending action can be observed and reacted to. Unpause carries the shortest delay because it is the action most likely to be needed urgently and least likely to be dangerous.

### 8.4 Reentrancy and Transfer Safety

Every function moving ERC-20 value uses `SafeERC20` rather than raw `transfer`/`transferFrom`, and is guarded by `ReentrancyGuardUpgradeable`. State is always updated before an external call. The pull-payment pattern (Section 7.4) further reduces the external-call surface in the settlement path to a single USDC pull from the buyer.

`CarbonCreditToken.mint` makes no external calls whatsoever, so the highest-value operation in the system has no reentrancy surface at all.

### 8.5 Custom Errors

Every `require` is replaced with a custom error type (`CarbonCreditToken__ClaimAlreadyConsumed(bytes32 claimId)`), prefixed with the contract name to avoid selector collisions. This is a gas optimization and a readability improvement, and it lets the backend's error-mapping layer branch precisely on error type rather than parse strings.

### 8.6 Fixed Compiler Version and Static Analysis

Every contract pins an exact Solidity compiler version, with no floating pragma, so the bytecode produced is reproducible across every environment and audit. Slither and Mythril run as a mandatory CI gate — a finding above a defined severity threshold blocks merge rather than producing a post-hoc report.

### 8.7 Invariant and Fuzz Testing

Beyond conventional unit tests, the system defines and continuously fuzz-tests the properties that actually define correctness, independent of any individual function:

- The sum of all balances plus all retirements for a token ID equals total ever minted for that ID
- A consumed claim flag, once true, is never false again
- No authorization nonce is ever accepted twice
- Total minted within any 24-hour window never exceeds the configured global limit
- The sum of all active listings' escrowed amounts, per token ID, never exceeds the Marketplace's actual balance of that ID
- `totalRetired` is monotonically non-decreasing
- The sum of `pendingProceeds` plus `accumulatedFees` never exceeds the Marketplace's USDC balance
- A frozen account's balance never changes
- Every anchored epoch is exactly one greater than the previous, with no gaps

The last invariant on solvency — proceeds plus fees never exceeding the actual USDC balance — is the one that catches the entire class of rounding and accounting bugs that unit tests structurally cannot, because it only fails across a sequence of individually-valid operations.

---

## 9. Proxy Architecture, Upgradeability, and Storage Layout Verification

### 9.1 Proxy and Logic Are Two Separate Contracts, Always

Every upgradeable contract is physically split into two contracts that are never merged, never conflated, and never referred to interchangeably:

- **The Proxy** (`ERC1967Proxy`) — deployed exactly once per logical contract, holds **all persistent storage** (every balance, every listing, every epoch root, every claim flag), and is the **only address** ever referenced by any other contract, the backend, or the frontend. This address never changes for the lifetime of the system.
- **The Implementation** — holds **all logic** and no persistent data of its own. The proxy `delegatecall`s into whichever implementation it currently points at, executing that logic in the proxy's own storage context; the implementation's own storage, were it ever called directly, is irrelevant and unused.

This split is what delivers "fix a bug without losing data": fixing a bug means deploying a new implementation and pointing the existing, unchanged proxy at it. The proxy's address and everything stored there — every balance, every anchored root, every open listing — is untouched. Nothing is migrated, because the address callers have always used still resolves to the storage it always has; only the logic behind it changed.

### 9.2 Why UUPS

UUPS is used rather than the Transparent Proxy pattern: the upgrade function lives in the implementation rather than in a separate, permanently-deployed proxy-admin contract. This keeps the proxy minimal — cheaper to deploy, smaller attack surface at the one address everything points to — and puts upgrade authorization through the same `AccessControlUpgradeable` plus timelock machinery used for every other privileged action, rather than a separate admin-contract pattern with its own rules.

### 9.3 Initialization Replaces Constructors

Because the proxy holds state and a constructor runs only in the context of the contract being deployed, **no upgradeable contract uses a constructor for any state-setting logic.** Every implementation exposes an `initialize` function protected by the `initializer` modifier, called in the proxy's own deployment transaction. The implementation's constructor is used only to call `_disableInitializers()`, preventing the implementation from ever being initialized and used directly, bypassing its proxy — a mandatory safeguard against a known class of proxy-implementation confusion attack.

### 9.4 Storage Slot Verification Is Mandatory

**No storage layout is ever reasoned about by reading source code top to bottom.** After every change to any implementation — including one that looks storage-neutral — the layout is generated by tooling (`forge inspect <Contract> storage-layout --pretty`) and compared against the previously committed layout. That generated layout is committed to the repository as a versioned artifact alongside the contract, so every change to it appears in code review as an explicit diff rather than something a reviewer must infer.

Before any upgrade is proposed through the timelock, the layout diff between the deployed implementation and the proposed one is generated and reviewed. Any slot reordering, type-width change, or unexpected insertion is a blocking finding, not a judgment call. OpenZeppelin's upgrade-safety validator runs in the same CI gate as the static analysis tools — a proposed upgrade failing it cannot reach the timelock queue at all.

### 9.5 Storage Gaps

Every upgradeable contract reserves `uint256[50] private __gap` at the end of its storage layout, so a future version adding state variables does not shift the slots of anything inheriting from it. This is verified as part of the same layout diff rather than assumed sufficient by convention.

---

## 10. Code Style and Modularity Conventions

- **One concern per contract.** Token accounting, provenance anchoring, batch identity, and marketplace trading are four separate contracts — no contract accumulates responsibilities across these domains, mirroring the single-responsibility discipline applied to the backend services.
- **Consistent internal ordering**, per the Solidity style guide: state variables, events, custom errors, modifiers, constructor/initializer, external, public, internal, private — the same order in every contract, so any contributor can navigate an unfamiliar file by structure alone.
- **NatSpec on every public and external function** (`@notice`, `@param`, `@return`), treated as a completeness requirement rather than optional documentation.
- **Functions kept short and single-purpose.** Core logic exceeding roughly 20–30 lines is a signal to extract an internal helper — a readability rule as much as a testability one, since a shorter function is easier to reason about and easier to fuzz in isolation.
- **Explicit visibility everywhere**, with `external` preferred over `public` for anything never called internally — minimizing the public interface surface is a security property, not just style.
- **No `tx.origin` for authorization**, anywhere, under any circumstance — only `msg.sender`.
- **Naming:** `PascalCase` for contracts and structs, `camelCase` for functions and variables, `UPPER_SNAKE_CASE` for constants and role identifiers, an `I` prefix for interfaces, a `Lib` suffix for libraries — applied uniformly so the type of any identifier is inferable from its name.

---

## 11. Cross-Contract and Cross-System Interaction Flows

### 11.1 Mint Flow

A Verifier's approval in Sustainability Service produces an EIP-712 `MintAuthorization` signed by that Verifier's authority key — two signatures where the amount exceeds the dual-approval threshold. Sustainability Service publishes `claim.decision.recorded` carrying the authorization; Credit Ledger Service consumes it and calls `ChainWriterService.SubmitMintTransaction`. Chain Writer Service, holding `MINTER_ROLE`, submits `mint` with the authorization and its signatures attached.

The contract verifies the signatures itself. Chain Writer Service's role is to pay gas and sequence nonces; it contributes no authority to the operation. The resulting `CreditMinted` event is observed by Chain Observer Service at confirmation depth, converted into `chain.event.observed`, and consumed by Credit Ledger Service to reconcile its mirror against the now-authoritative on-chain state.

If the mint reverts on a rate limit, Chain Writer Service treats it as a **deferral**: the transaction re-queues with backoff, the claim stays approved and pending issuance, and the Organization is notified that issuance is delayed. A safety limit must never become a data-loss path.

### 11.2 Anchoring Flow

Chain Observer Service accumulates every batch and checkpoint event across the platform during a 10-minute epoch, builds one Merkle tree over the epoch's leaves, and calls `ChainWriterService.SubmitAnchorTransaction`, which — holding `ANCHOR_ROLE` — calls `anchorEpoch`. One transaction covers the entire platform for that epoch.

The public batch page never calls this contract directly; it reads from Provenance Read Service's projection, which is kept current by consuming the resulting `EpochAnchored` event. A consumer wanting cryptographic proof rather than the platform's word for it can request the inclusion proof for any specific checkpoint and verify it against `verifyInclusion` themselves — which is the entire point of anchoring, and only works because historical roots are retained permanently.

### 11.3 Marketplace Flow

A purchase, listing, or retirement initiated from the frontend is signed directly by the user's own wallet against `Marketplace` or `CarbonCreditToken` — the one on-chain interaction path in the system that does **not** route through Chain Writer Service, since it is the user's own transaction rather than one the backend signs on anyone's behalf.

The signing wallet must be the Organization's Treasury Address, or an address the Treasury Address has approved as an ERC-1155 operator. Credits are Organization assets, and an individual employee's personal wallet is never their custodian.

Chain Observer Service picks up the resulting events and feeds them into the same reconciliation and notification pipelines as every other on-chain action, so a purchase whose confirmation the user's browser never saw still resolves correctly on the platform side.

### 11.4 Compliance Freeze Flow

An Admin escalating a fraud flag causes Admin Backend to call `ChainWriterService.SubmitFreezeTransaction`, which — holding `COMPLIANCE_ROLE` — calls `setFrozen` on the token for that Organization's Treasury Address. The freeze is enforced in the ERC-1155 transfer hook and therefore applies to every transfer path uniformly. Marketplace Service concurrently cancels the Organization's active listings, returning escrowed credits to the now-frozen address, where they remain until the freeze lifts.

---

## 12. Local Development, Testing, and Deployment

The contract suite is built and verified entirely on **Anvil** before any testnet deployment is considered — no contract is deployed to a public network as part of this phase.

### 12.1 Toolchain

Foundry (`forge`, `cast`, `anvil`) for the entire suite — compilation, testing, scripted deployment, and the storage-layout inspection in Section 9.4 all come from the same toolchain, avoiding the inconsistencies of mixing Hardhat and Foundry workflows.

### 12.2 Deployment Scripts

Deployment is scripted (`forge script`), never performed by manually pasted commands against a live node — a deployment script is a versioned, reviewable artifact exactly like the contracts it deploys. The script executes, in order:

1. Deploy `TimelockController` with the configured multisig as proposer and executor.
2. Deploy each implementation contract.
3. Deploy each proxy, with the `initialize` call encoded into the proxy's own deployment transaction so no contract is ever momentarily uninitialized.
4. Deploy `CreditTokenFactory` and use it to deploy the canonical `CarbonCreditToken` proxy and implementation.
5. Grant every role from the Section 6 matrix to its correct holder.
6. Register the initial verifier authorities and the Marketplace address on the token.
7. Configure rate limits, default fee, and epoch length.
8. Renounce every deployer-held role, so the deploying key retains no authority once deployment completes.
9. Run storage-layout inspection against every deployed implementation as a final automated check that deployed bytecode matches what was reviewed.

Step 8 is the one most often skipped and most consequential: a deployment script that leaves the deployer holding `DEFAULT_ADMIN_ROLE` has quietly created a single-key backdoor around the entire multisig-and-timelock design.

Running this same script against Anvil and later against a testnet, with only the RPC endpoint and signer changing, is what makes the local environment a faithful rehearsal rather than a divergent setup.

### 12.3 Test Suite Structure

- **Unit tests**, one file per contract, covering every function's success path, every documented revert condition mapped to its custom error, and every access-control boundary — a call from an address lacking the required role must revert, tested explicitly for every privileged function.
- **Signature tests** specifically targeting the mint authorization path: expired deadline, reused nonce, unregistered signer, duplicate signer presented twice for a dual-approval mint, signature over mutated parameters, and signature replay against a different chain ID.
- **Fuzz tests** targeting the Section 8.7 invariants with randomized call sequences and inputs — the goal is finding a sequence of individually-valid operations that breaks an invariant, which example-based tests are structurally unable to discover.
- **Integration tests** exercising the Section 11 flows end to end in a single test (authorize, mint, list, purchase, retire; anchor an epoch and verify inclusion of a specific leaf) to catch issues appearing only across function boundaries.
- **Fork tests** against a Base fork, exercising settlement against the real USDC contract rather than a mock — including a blacklisted-address scenario, which is precisely the case a mock would not reproduce and the pull-payment design exists to handle.
- **Invariant coverage is tracked as its own metric**, separate from line and branch coverage. A contract can reach 100% line coverage from unit tests while an invariant has never been fuzzed, and that gap is an incomplete test suite regardless of the coverage number.

### 12.4 Local Environment Parity

Anvil is configured to mirror Base's 2-second block time and gas behaviour, so gas assumptions and time-window logic — the rolling mint rate limit, listing expiry, the guardian pause auto-expiry — behave locally as they will in a real deployment. A local environment that diverges materially from production chain behaviour lets exactly this class of bug pass locally and surface only later.

---

## 13. Event Design and Log Management

Every contract emits a complete, structured event for every state-changing action. This is not incidental logging — it is the **primary data source** for Chain Observer Service, which converts each event into both a structured log line and a Prometheus metric. A contract action that doesn't emit a sufficiently detailed event is an incomplete implementation, not a minor omission, since the off-chain system has no other way to learn what happened.

### 13.1 Event Inventory

| Event | Contract | Indexed | Purpose |
|---|---|---|---|
| `CreditMinted` | `CarbonCreditToken` | `claimId`, `to`, `tokenId` | Drives Credit Ledger reconciliation and the credits-issued notification. Carries the decomposed Credit Class so consumers need no additional lookup |
| `CreditRetired` | `CarbonCreditToken` | `account`, `tokenId` | The sole basis for the public retirement log — carries the retiring account, amount, reason, and Credit Class |
| `MintRateLimitExceeded` | `CarbonCreditToken` | `facilityId` | High-priority Fraud Detection signal, and the trigger for Chain Writer's deferral path |
| `AccountFrozen` / `AccountUnfrozen` | `CarbonCreditToken` | `account` | Compliance action audit trail |
| `VerifierAuthorityChanged` | `CarbonCreditToken` | `authority` | Security-critical — changes who can authorize value creation |
| `MintLimitsChanged` | `CarbonCreditToken` | — | Security-critical parameter change |
| `EpochAnchored` | `ProvenanceAnchor` | `epoch` | Drives the Provenance Read projection and is the permanent record of every anchor |
| `BatchRegistered` | `BatchRegistry` | `batchRef`, `tokenId` | Confirms on-chain batch identity creation |
| `MetadataFrozen` | `BatchRegistry` | `tokenId` | Records that a batch's metadata is now permanently immutable |
| `ListingCreated` | `Marketplace` | `listingId`, `seller`, `tokenId` | Marketplace projection update |
| `Purchased` | `Marketplace` | `listingId`, `buyer` | Trade settlement confirmation, drives Notification Service, carries cost, fee, and proceeds |
| `ListingCancelled` / `ListingExpired` | `Marketplace` | `listingId` | Projection update and escrow return confirmation |
| `ProceedsWithdrawn` | `Marketplace` | `seller` | Settlement completion for the seller |
| `SellerFeeChanged` | `Marketplace` | `seller` | Ties an on-chain fee to a plan change |
| `Paused` / `Unpaused` | Every pausable contract | — | Highest-priority alert to both Notification Service and the observability alerting pipeline, bypassing normal-priority routing |
| `RoleGranted` / `RoleRevoked` | Every contract (inherited) | `role`, `account` | Security-relevant by definition — every role change is logged and alerted, never silent |
| `Upgraded` | Every UUPS contract (inherited) | `implementation` | Every upgrade execution logged with the new implementation address, cross-referenced against that upgrade's storage-layout diff artifact |

### 13.2 Correlating On-Chain Events with Off-Chain Logs

The blockchain has no native correlation ID, so business identifiers already present in event parameters — `claimId`, `batchRef`, `listingId`, `epoch` — serve as the correlation key between an on-chain event and the off-chain request that triggered it. When Chain Writer Service submits a transaction it logs the submission with both its own request-scoped correlation ID and the business identifier; Chain Observer Service's log line for the resulting event carries the same business identifier. Searching for a given `claimId` therefore surfaces the entire lifecycle — the Verifier's decision, the authorization, the transaction submission, and the on-chain confirmation — as one continuous trail, even though the on-chain event itself carries no tracing metadata beyond its transaction hash and block number.

### 13.3 What Chain Observer Extracts Per Event

For every event: transaction hash, block number, block timestamp, confirmation depth at observation, emitting contract address, every indexed parameter, and a fully decoded payload. Pushed as a structured log line and, for events with numeric significance (mint amounts, retirement amounts, settlement costs, fee amounts), also as a Prometheus metric with the indexed parameters available as labels.

An event is emitted onward as `chain.event.observed` **only once it reaches 30 confirmations.** Events seen at shallower depth update provisional UI state only and are explicitly rolled back if a reorg invalidates them before reaching confirmation depth — the reconciliation job is the backstop for any rollback the observer misses.
