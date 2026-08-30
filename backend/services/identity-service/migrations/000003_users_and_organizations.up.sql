CREATE TABLE identity.users (
    id               uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    deleted_at       timestamptz,
    version          integer NOT NULL DEFAULT 1,
    auth0_subject    text,
    email            citext NOT NULL,
    email_verified   boolean NOT NULL DEFAULT false,
    name             text NOT NULL,
    platform_role    identity.platform_role,
    mfa_enrolled_at  timestamptz,
    last_active_at   timestamptz
);

CREATE UNIQUE INDEX users_auth0_subject ON identity.users (auth0_subject)
    WHERE auth0_subject IS NOT NULL AND deleted_at IS NULL;
CREATE UNIQUE INDEX users_email ON identity.users (email) WHERE deleted_at IS NULL;

CREATE TABLE identity.organizations (
    id                            uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at                    timestamptz NOT NULL DEFAULT now(),
    updated_at                    timestamptz NOT NULL DEFAULT now(),
    deleted_at                    timestamptz,
    version                       integer NOT NULL DEFAULT 1,
    name                          text NOT NULL,
    type                          identity.organization_type NOT NULL,
    country_of_incorporation      char(2) NOT NULL,
    business_registration_number  text NOT NULL,
    verification_status           identity.verification_status NOT NULL,
    state                         identity.organization_state NOT NULL DEFAULT 'active',
    product_categories            identity.product_category[] NOT NULL DEFAULT '{}',
    registry_record_id            uuid REFERENCES identity.business_registry_records (id),
    name_similarity               numeric(4,3),
    rejection_reason              identity.registry_rejection,
    verified_at                   timestamptz,
    verification_source           identity.verification_source NOT NULL DEFAULT 'registry',
    overridden_by_user_id         uuid REFERENCES identity.users (id),
    override_justification        text,
    CONSTRAINT organizations_rejection_requires_reason
        CHECK (verification_status <> 'rejected' OR rejection_reason IS NOT NULL),
    CONSTRAINT organizations_override_requires_justification
        CHECK (verification_source <> 'manual_override' OR override_justification IS NOT NULL)
);

CREATE UNIQUE INDEX organizations_registration
    ON identity.organizations (country_of_incorporation, business_registration_number)
    WHERE deleted_at IS NULL;
CREATE INDEX organizations_registry_record_id ON identity.organizations (registry_record_id);
CREATE INDEX organizations_overridden_by_user_id ON identity.organizations (overridden_by_user_id);

CREATE TABLE identity.organization_memberships (
    id                  uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    deleted_at          timestamptz,
    version             integer NOT NULL DEFAULT 1,
    organization_id     uuid NOT NULL REFERENCES identity.organizations (id),
    user_id             uuid NOT NULL REFERENCES identity.users (id),
    role                identity.organization_role NOT NULL,
    state               identity.membership_state NOT NULL DEFAULT 'active',
    invited_by_user_id  uuid REFERENCES identity.users (id),
    joined_at           timestamptz,
    revoked_at          timestamptz
);

CREATE UNIQUE INDEX organization_memberships_unique
    ON identity.organization_memberships (organization_id, user_id)
    WHERE deleted_at IS NULL;
CREATE INDEX organization_memberships_organization_id ON identity.organization_memberships (organization_id);
CREATE INDEX organization_memberships_user_id ON identity.organization_memberships (user_id);
CREATE INDEX organization_memberships_invited_by_user_id ON identity.organization_memberships (invited_by_user_id);
CREATE INDEX organization_memberships_active_owners
    ON identity.organization_memberships (organization_id)
    WHERE role = 'owner' AND state = 'active' AND deleted_at IS NULL;

CREATE TABLE identity.invitations (
    id                   uuid PRIMARY KEY DEFAULT uuidv7(),
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz,
    version              integer NOT NULL DEFAULT 1,
    organization_id      uuid NOT NULL REFERENCES identity.organizations (id),
    email                citext NOT NULL,
    role                 identity.organization_role NOT NULL,
    token_hash           bytea NOT NULL,
    state                identity.invitation_state NOT NULL DEFAULT 'pending',
    invited_by_user_id   uuid NOT NULL REFERENCES identity.users (id),
    expires_at           timestamptz NOT NULL,
    accepted_at          timestamptz,
    accepted_by_user_id  uuid REFERENCES identity.users (id)
);

CREATE UNIQUE INDEX invitations_token_hash ON identity.invitations (token_hash);
CREATE UNIQUE INDEX invitations_one_pending_per_email
    ON identity.invitations (organization_id, email)
    WHERE state = 'pending' AND deleted_at IS NULL;
CREATE INDEX invitations_organization_id ON identity.invitations (organization_id);
CREATE INDEX invitations_invited_by_user_id ON identity.invitations (invited_by_user_id);
CREATE INDEX invitations_accepted_by_user_id ON identity.invitations (accepted_by_user_id);
