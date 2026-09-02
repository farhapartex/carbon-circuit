# Architecture

[← Back to the README](../README.md)

## System design — request and event flow

![CarbonCircuit system design: request and event flow](../System-design.png)

Fifteen services across eight domains. Requests enter through a single API Gateway that resolves caller context **once** and stamps it into an internal service token, so no write path needs more than two synchronous hops. Everything that isn't a question-and-answer runs through Kafka with a transactional outbox on produce and a transactional inbox on consume. The public QR-scan page is served from a pre-computed read model behind a CDN and never touches the write side or the chain.

One boundary is drawn harder than any other: **Chain Writer Service is the only component in the system with access to signing keys**, and even it cannot authorize anything — a mint carries EIP-712 verifier signatures that the contract verifies for itself.

## Smart contract design

![CarbonCircuit smart contract design on Base L2](../Smart-Contract-Design.png)

Four upgradeable contracts on Base, each deployed as a proxy holding all storage plus a freely replaceable implementation holding all logic. Credits are **ERC-1155**, one token ID per Credit Class, so attribution is a property of the token rather than bookkeeping alongside it. Provenance is anchored as **one Merkle root per 10-minute epoch platform-wide**, so on-chain cost doesn't scale with customer count and every historical root stays permanently verifiable.

Governance sits behind a 3-of-5 multisig routed through a timelock, with a separate hot guardian key that can pause but never unpause — and whose pause auto-expires after 72 hours so a compromised guardian causes a bounded outage rather than an indefinite one.
