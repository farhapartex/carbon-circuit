CREATE TYPE sustainability.factor_kind AS ENUM ('grid_emission', 'logistics_baseline', 'material_avoided');

CREATE TABLE sustainability.reference_factors (
    id             uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    deleted_at     timestamptz,
    version        integer NOT NULL DEFAULT 1,
    kind           sustainability.factor_kind NOT NULL,
    lookup_key     text NOT NULL,
    unit           text NOT NULL,
    factor         numeric(20,6) NOT NULL,
    effective_from date NOT NULL,
    effective_to   date,
    CONSTRAINT reference_factors_range CHECK (effective_to IS NULL OR effective_to > effective_from),
    CONSTRAINT reference_factors_positive CHECK (factor > 0)
);

CREATE UNIQUE INDEX reference_factors_effective
    ON sustainability.reference_factors (kind, lookup_key, effective_from)
    WHERE deleted_at IS NULL;

CREATE INDEX reference_factors_lookup
    ON sustainability.reference_factors (kind, lookup_key, effective_from DESC)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE sustainability.reference_factors IS
    'Seeded reference values from PRD 2.6. Rows are versioned by effective date range and never updated in place, because a claim pins the row it used and a recomputation years later must produce the same answer.';
