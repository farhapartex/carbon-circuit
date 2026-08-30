CREATE TABLE billing.invoices (
    id                 uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),
    deleted_at         timestamptz,
    version            integer NOT NULL DEFAULT 1,
    organization_id    uuid NOT NULL,
    stripe_invoice_id  text NOT NULL,
    number             text NOT NULL,
    amount_usd         numeric(10,2) NOT NULL,
    status             billing.invoice_status NOT NULL,
    issued_at          timestamptz NOT NULL,
    paid_at            timestamptz,
    CONSTRAINT invoices_paid_has_timestamp CHECK (status <> 'paid' OR paid_at IS NOT NULL)
);

CREATE UNIQUE INDEX invoices_stripe_invoice_id ON billing.invoices (stripe_invoice_id);
CREATE INDEX invoices_organization_issued ON billing.invoices (organization_id, issued_at DESC);

CREATE TABLE billing.payment_methods (
    id                        uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at                timestamptz NOT NULL DEFAULT now(),
    updated_at                timestamptz NOT NULL DEFAULT now(),
    deleted_at                timestamptz,
    version                   integer NOT NULL DEFAULT 1,
    organization_id           uuid NOT NULL,
    stripe_payment_method_id  text NOT NULL,
    brand                     text NOT NULL,
    last4                     char(4) NOT NULL,
    expiry_month              smallint NOT NULL,
    expiry_year               smallint NOT NULL,
    is_default                boolean NOT NULL DEFAULT false,
    CONSTRAINT payment_methods_month_valid CHECK (expiry_month BETWEEN 1 AND 12)
);

CREATE UNIQUE INDEX payment_methods_stripe_id ON billing.payment_methods (stripe_payment_method_id);
CREATE UNIQUE INDEX payment_methods_one_default
    ON billing.payment_methods (organization_id) WHERE is_default AND deleted_at IS NULL;
CREATE INDEX payment_methods_organization_id ON billing.payment_methods (organization_id);

CREATE TABLE billing.webhook_events (
    id               uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz,
    version          integer NOT NULL DEFAULT 1,
    stripe_event_id  text NOT NULL,
    type             text NOT NULL,
    payload          jsonb NOT NULL,
    state            billing.webhook_state NOT NULL DEFAULT 'received',
    received_at      timestamptz NOT NULL DEFAULT now(),
    processed_at     timestamptz
);

CREATE UNIQUE INDEX webhook_events_stripe_event_id ON billing.webhook_events (stripe_event_id);
CREATE INDEX webhook_events_unprocessed ON billing.webhook_events (received_at) WHERE processed_at IS NULL;
