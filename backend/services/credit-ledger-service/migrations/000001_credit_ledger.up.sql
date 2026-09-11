CREATE TYPE credit_ledger.activity_type AS ENUM (
    'renewable_energy', 'reduced_emission_logistics', 'responsible_sourcing'
);

CREATE TYPE credit_ledger.anchor_state AS ENUM ('unanchored', 'submitted', 'anchored');

CREATE TYPE credit_ledger.idempotency_state AS ENUM ('processing', 'completed', 'failed');

CREATE TABLE credit_ledger.credit_classes (
    id                 uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),
    deleted_at         timestamptz,
    version            integer NOT NULL DEFAULT 1,
    token_id           numeric(78,0) NOT NULL,
    facility_id        uuid NOT NULL,
    facility_name      text NOT NULL,
    facility_country   char(2) NOT NULL,
    vintage_year       integer NOT NULL,
    activity_type      credit_ledger.activity_type NOT NULL,
    CONSTRAINT credit_classes_vintage_fits_uint16 CHECK (vintage_year BETWEEN 0 AND 65535),
    CONSTRAINT credit_classes_token_positive CHECK (token_id > 0)
);

CREATE UNIQUE INDEX credit_classes_token ON credit_ledger.credit_classes (token_id);
CREATE UNIQUE INDEX credit_classes_attributes
    ON credit_ledger.credit_classes (facility_id, vintage_year, activity_type);

COMMENT ON TABLE credit_ledger.credit_classes IS
    'The permanent, inseparable combination of facility, vintage and activity that PRD 2.3 forbids pooling. Carries no organization_id: a class describes what a credit is, not who holds it, and the same class can be held by many organizations after a trade.';
COMMENT ON COLUMN credit_ledger.credit_classes.token_id IS
    'The ERC-1155 token id the contract will derive for this class, computed off-chain from the pinned convention in the smart contract design so the two agree before any mint.';

CREATE TABLE credit_ledger.credit_balances (
    id               uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz,
    version          integer NOT NULL DEFAULT 1,
    organization_id  uuid NOT NULL,
    credit_class_id  uuid NOT NULL REFERENCES credit_ledger.credit_classes (id),
    available        numeric(28,6) NOT NULL DEFAULT 0,
    escrowed         numeric(28,6) NOT NULL DEFAULT 0,
    retired          numeric(28,6) NOT NULL DEFAULT 0,
    CONSTRAINT credit_balances_never_negative CHECK (
        available >= 0 AND escrowed >= 0 AND retired >= 0
    )
);

CREATE UNIQUE INDEX credit_balances_holding
    ON credit_ledger.credit_balances (organization_id, credit_class_id);

CREATE INDEX credit_balances_organization
    ON credit_ledger.credit_balances (organization_id)
    WHERE deleted_at IS NULL;

COMMENT ON CONSTRAINT credit_balances_never_negative ON credit_ledger.credit_balances IS
    'The database refuses a negative holding outright. Overselling and over-retiring are then impossible rather than merely checked for, whatever a caller asks.';

CREATE TABLE credit_ledger.credit_issuances (
    id               uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz,
    version          integer NOT NULL DEFAULT 1,
    organization_id  uuid NOT NULL,
    credit_class_id  uuid NOT NULL REFERENCES credit_ledger.credit_classes (id),
    claim_id         uuid NOT NULL,
    amount           numeric(28,6) NOT NULL,
    treasury_address char(42) NOT NULL,
    anchor_state     credit_ledger.anchor_state NOT NULL DEFAULT 'unanchored',
    transaction_hash char(66),
    issued_at        timestamptz NOT NULL DEFAULT now(),
    anchored_at      timestamptz,
    CONSTRAINT credit_issuances_amount_positive CHECK (amount > 0),
    CONSTRAINT credit_issuances_hash_matches_state CHECK (
        (anchor_state = 'anchored') = (transaction_hash IS NOT NULL)
    )
);

CREATE UNIQUE INDEX credit_issuances_claim ON credit_ledger.credit_issuances (claim_id);

CREATE INDEX credit_issuances_organization
    ON credit_ledger.credit_issuances (organization_id, issued_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX credit_issuances_unanchored
    ON credit_ledger.credit_issuances (issued_at)
    WHERE anchor_state <> 'anchored' AND deleted_at IS NULL;

COMMENT ON INDEX credit_ledger.credit_issuances_claim IS
    'One issuance per claim, enforced by the database. A replayed decision event cannot mint a second time.';
COMMENT ON COLUMN credit_ledger.credit_issuances.anchor_state IS
    'Whether this issuance has reached the chain. Everything is unanchored while no chain writer exists, and that is shown rather than hidden, since an off-chain balance is only as good as the record that it has not yet been anchored.';
