from xml.sax.saxutils import escape

BASE = [
    ("id", "uuid", "PK"),
    ("created_at", "timestamptz", ""),
    ("updated_at", "timestamptz", ""),
    ("deleted_at", "timestamptz NULL", ""),
    ("version", "integer", ""),
]

IDENTITY = [
    ("organizations", "Tenant root. One row per customer company.", [
        ("name", "text", ""),
        ("type", "organization_type", "ENUM"),
        ("country_of_incorporation", "char(2)", "UQ1"),
        ("business_registration_number", "text", "UQ1"),
        ("verification_status", "verification_status", "ENUM"),
        ("state", "organization_state", "ENUM"),
        ("product_categories", "product_category[]", ""),
        ("registry_record_id", "uuid NULL", "FK business_registry_records"),
        ("name_similarity", "numeric(4,3) NULL", ""),
        ("rejection_reason", "registry_rejection NULL", "ENUM"),
        ("verified_at", "timestamptz NULL", ""),
        ("verification_source", "verification_source", "ENUM"),
        ("overridden_by_user_id", "uuid NULL", ""),
        ("override_justification", "text NULL", ""),
    ]),
    ("users", "A person. Auth0 owns authentication; this owns identity.", [
        ("auth0_subject", "text NULL", "UQ"),
        ("email", "citext", "UQ"),
        ("email_verified", "boolean", ""),
        ("name", "text", ""),
        ("platform_role", "platform_role NULL", "ENUM"),
        ("mfa_enrolled_at", "timestamptz NULL", ""),
        ("last_active_at", "timestamptz NULL", ""),
    ]),
    ("organization_memberships", "Join table. A user may belong to several orgs.", [
        ("organization_id", "uuid", "FK organizations / UQ2"),
        ("user_id", "uuid", "FK users / UQ2"),
        ("role", "organization_role", "ENUM"),
        ("state", "membership_state", "ENUM"),
        ("invited_by_user_id", "uuid NULL", "FK users"),
        ("joined_at", "timestamptz NULL", ""),
        ("revoked_at", "timestamptz NULL", ""),
    ]),
    ("invitations", "Pending membership. Binds to a subject on acceptance.", [
        ("organization_id", "uuid", "FK organizations"),
        ("email", "citext", "UQ3 partial"),
        ("role", "organization_role", "ENUM"),
        ("token_hash", "bytea", "UQ"),
        ("state", "invitation_state", "ENUM"),
        ("invited_by_user_id", "uuid", "FK users"),
        ("expires_at", "timestamptz", ""),
        ("accepted_at", "timestamptz NULL", ""),
        ("accepted_by_user_id", "uuid NULL", "FK users"),
    ]),
    ("facilities", "A physical site. Its registry match sets the ceiling discount.", [
        ("organization_id", "uuid", "FK organizations / RLS"),
        ("name", "text", ""),
        ("address", "text", ""),
        ("country_code", "char(2)", ""),
        ("grid_region", "grid_region", "ENUM"),
        ("type", "facility_type", "ENUM"),
        ("facility_reference", "text NULL", ""),
        ("verification_status", "facility_verification", "ENUM"),
        ("ceiling_discount_factor", "numeric(3,2)", ""),
        ("trust_tier", "trust_tier", "ENUM"),
        ("declared_capacity", "numeric(20,6)", ""),
        ("declared_energy_kwh", "numeric(20,6)", ""),
        ("attested_capacity", "numeric(20,6) NULL", ""),
        ("attested_energy_kwh", "numeric(20,6) NULL", ""),
    ]),
    ("treasury_addresses", "Where an org's credits live. One active per org.", [
        ("organization_id", "uuid", "FK organizations / UQ4 partial"),
        ("address", "char(42)", ""),
        ("state", "treasury_state", "ENUM"),
        ("proof_signature", "bytea", ""),
        ("nonce_id", "uuid", "FK siwe_nonces"),
        ("designated_by_user_id", "uuid", "FK users"),
        ("designated_at", "timestamptz", ""),
        ("superseded_at", "timestamptz NULL", ""),
    ]),
    ("treasury_address_changes", "The 72 hour delayed change. Cancellable by any Owner.", [
        ("organization_id", "uuid", "FK organizations"),
        ("current_address", "char(42) NULL", ""),
        ("requested_address", "char(42)", ""),
        ("state", "treasury_change_state", "ENUM"),
        ("proof_signature", "bytea", ""),
        ("mfa_verified_at", "timestamptz", ""),
        ("initiated_by_user_id", "uuid", "FK users"),
        ("initiated_at", "timestamptz", ""),
        ("effective_at", "timestamptz", ""),
        ("resolved_at", "timestamptz NULL", ""),
        ("resolved_by_user_id", "uuid NULL", "FK users"),
    ]),
    ("siwe_nonces", "Single use, domain bound, short lived.", [
        ("nonce", "text", "UQ"),
        ("domain", "text", ""),
        ("user_id", "uuid NULL", "FK users"),
        ("issued_at", "timestamptz", ""),
        ("expires_at", "timestamptz", ""),
        ("consumed_at", "timestamptz NULL", ""),
    ]),
    ("api_keys", "Only prefix and HMAC stored. Secret is never recoverable.", [
        ("organization_id", "uuid", "FK organizations / RLS"),
        ("name", "text", ""),
        ("prefix", "char(8)", "UQ"),
        ("secret_hmac", "bytea", ""),
        ("created_by_user_id", "uuid", "FK users"),
        ("last_used_at", "timestamptz NULL", ""),
        ("revoked_at", "timestamptz NULL", ""),
        ("revoked_by_user_id", "uuid NULL", "FK users"),
    ]),
    ("sessions", "Surfaced on the profile page so a user can revoke a device.", [
        ("user_id", "uuid", "FK users"),
        ("user_agent", "text", ""),
        ("ip_address", "inet", ""),
        ("started_at", "timestamptz", ""),
        ("last_seen_at", "timestamptz", ""),
        ("revoked_at", "timestamptz NULL", ""),
    ]),
    ("token_revocations", "Durable source for the Redis denylist, so it can be rebuilt.", [
        ("subject", "text", "IDX"),
        ("reason", "revocation_reason", "ENUM"),
        ("revoked_at", "timestamptz", ""),
        ("expires_at", "timestamptz", "IDX"),
    ]),
    ("business_registry_records", "Seeded fixture. Swappable for a live registry API.", [
        ("country_code", "char(2)", "UQ5"),
        ("registration_number", "text", "UQ5"),
        ("legal_name", "text", ""),
        ("registered_address", "text", ""),
        ("incorporation_date", "date", ""),
        ("entity_status", "registry_entity_status", "ENUM"),
        ("industry_codes", "text[]", ""),
        ("sanctioned", "boolean", ""),
    ]),
    ("facility_registry_records", "Independently attested capacity. Feeds the discount factor.", [
        ("organization_registration_number", "text", "UQ6"),
        ("facility_reference", "text", "UQ6"),
        ("attested_capacity", "numeric(20,6)", ""),
        ("attested_energy_kwh", "numeric(20,6)", ""),
    ]),
    ("data_export_requests", "One per organization per 24 hours.", [
        ("organization_id", "uuid", "FK organizations / RLS"),
        ("requested_by_user_id", "uuid", "FK users"),
        ("state", "export_state", "ENUM"),
        ("requested_at", "timestamptz", ""),
        ("completed_at", "timestamptz NULL", ""),
        ("download_token_hash", "bytea NULL", ""),
        ("expires_at", "timestamptz NULL", ""),
    ]),
    ("deletion_requests", "Blocks while credits are held. Off chain purge only.", [
        ("organization_id", "uuid", "FK organizations / RLS"),
        ("requested_by_user_id", "uuid", "FK users"),
        ("state", "deletion_state", "ENUM"),
        ("blocked_reason", "text NULL", ""),
        ("requested_at", "timestamptz", ""),
        ("purge_after", "timestamptz", ""),
        ("completed_at", "timestamptz NULL", ""),
    ]),
]

BILLING = [
    ("plans", "Data, not constants. Admins change pricing without a redeploy.", [
        ("tier", "plan_tier", "ENUM / UQ7"),
        ("name", "text", ""),
        ("audience", "text", ""),
        ("monthly_price_usd", "numeric(10,2)", ""),
        ("price_note", "text NULL", ""),
        ("allowed_organization_types", "organization_type[]", ""),
        ("evidence_storage_gb", "integer NULL", ""),
        ("portal_rate_per_minute", "integer", ""),
        ("api_rate_per_minute", "integer NULL", ""),
        ("api_key_limit", "integer NULL", ""),
        ("marketplace_fee_bps", "smallint NULL", ""),
        ("review_turnaround", "text", ""),
        ("support_level", "text", ""),
        ("effective_from", "timestamptz", "UQ7"),
        ("effective_to", "timestamptz NULL", ""),
    ]),
    ("plan_limits", "Normalised so a quota check is a query, not JSON parsing.", [
        ("plan_id", "uuid", "FK plans / UQ8"),
        ("dimension", "usage_dimension", "ENUM / UQ8"),
        ("included", "bigint NULL", ""),
        ("fair_use_ceiling", "bigint NULL", ""),
        ("overage_rate_usd", "numeric(10,2) NULL", ""),
        ("blocks_on_exhaustion", "boolean", ""),
    ]),
    ("subscriptions", "Exactly one plan per organization at a time.", [
        ("organization_id", "uuid", "UQ / RLS"),
        ("plan_id", "uuid", "FK plans"),
        ("state", "subscription_state", "ENUM"),
        ("stripe_customer_id", "text NULL", "UQ"),
        ("stripe_subscription_id", "text NULL", "UQ"),
        ("current_period_start", "timestamptz", ""),
        ("current_period_end", "timestamptz", ""),
        ("grace_period_ends_at", "timestamptz NULL", ""),
        ("cancel_at", "timestamptz NULL", ""),
        ("cancelled_at", "timestamptz NULL", ""),
    ]),
    ("plan_overrides", "Trial extensions and negotiated Enterprise terms.", [
        ("organization_id", "uuid", "RLS / UQ9"),
        ("dimension", "usage_dimension", "ENUM / UQ9"),
        ("included", "bigint NULL", ""),
        ("overage_rate_usd", "numeric(10,2) NULL", ""),
        ("justification", "text", ""),
        ("created_by_admin_id", "uuid", ""),
        ("expires_at", "timestamptz NULL", ""),
    ]),
    ("usage_counters", "Incremented under a lock, mirrored into Redis for local reads.", [
        ("organization_id", "uuid", "RLS / UQ10"),
        ("dimension", "usage_dimension", "ENUM / UQ10"),
        ("period_start", "timestamptz", "UQ10"),
        ("period_end", "timestamptz", ""),
        ("used", "bigint", ""),
    ]),
    ("invoices", "", [
        ("organization_id", "uuid", "RLS"),
        ("stripe_invoice_id", "text", "UQ"),
        ("number", "text", ""),
        ("amount_usd", "numeric(10,2)", ""),
        ("status", "invoice_status", "ENUM"),
        ("issued_at", "timestamptz", ""),
        ("paid_at", "timestamptz NULL", ""),
    ]),
    ("payment_methods", "", [
        ("organization_id", "uuid", "RLS"),
        ("stripe_payment_method_id", "text", "UQ"),
        ("brand", "text", ""),
        ("last4", "char(4)", ""),
        ("expiry_month", "smallint", ""),
        ("expiry_year", "smallint", ""),
        ("is_default", "boolean", ""),
    ]),
    ("webhook_events", "Stripe retries and reorders. The unique key makes both safe.", [
        ("stripe_event_id", "text", "UQ"),
        ("type", "text", ""),
        ("payload", "jsonb", ""),
        ("state", "webhook_state", "ENUM"),
        ("received_at", "timestamptz", ""),
        ("processed_at", "timestamptz NULL", ""),
    ]),
]

SHARED = [
    ("idempotency_records", "Reserved by atomic insert. Completed in the business txn.", [
        ("organization_id", "uuid", "UQ11"),
        ("endpoint", "text", "UQ11"),
        ("idempotency_key", "text", "UQ11"),
        ("request_hash", "bytea", ""),
        ("state", "idempotency_state", "ENUM"),
        ("response_status", "integer NULL", ""),
        ("response_body", "jsonb NULL", ""),
        ("resource_id", "uuid NULL", ""),
        ("completed_at", "timestamptz NULL", ""),
    ]),
    ("inbox_events", "Inserted in the same txn as the effect. Makes replay a no-op.", [
        ("event_id", "uuid", "UQ"),
        ("topic", "text", ""),
        ("consumer_group", "text", ""),
        ("processed_at", "timestamptz", ""),
    ]),
    ("outbox_events", "Written with the business row. A relay publishes from here.", [
        ("aggregate_type", "text", ""),
        ("aggregate_id", "uuid", "IDX"),
        ("event_type", "text", ""),
        ("payload", "jsonb", ""),
        ("headers", "jsonb", ""),
        ("published_at", "timestamptz NULL", "IDX partial"),
    ]),
]

FOREIGN_KEYS = [
    ("organization_memberships", "organizations"),
    ("organization_memberships", "users"),
    ("invitations", "organizations"),
    ("invitations", "users"),
    ("facilities", "organizations"),
    ("treasury_addresses", "organizations"),
    ("treasury_addresses", "siwe_nonces"),
    ("treasury_address_changes", "organizations"),
    ("api_keys", "organizations"),
    ("api_keys", "users"),
    ("sessions", "users"),
    ("organizations", "business_registry_records"),
    ("data_export_requests", "organizations"),
    ("deletion_requests", "organizations"),
    ("plan_limits", "plans"),
    ("subscriptions", "plans"),
]

ENUMS = [
    ("organization_type", "manufacturer | assembler | logistics | credit_buyer"),
    ("organization_state", "active | restricted | read_only | suspended"),
    ("verification_status", "verified | unverified | rejected"),
    ("verification_source", "registry | manual_override"),
    ("registry_rejection", "entity_dissolved | sanctions_flag | name_mismatch"),
    ("registry_entity_status", "active | dissolved"),
    ("organization_role", "owner | admin | member"),
    ("platform_role", "verifier | platform_admin"),
    ("membership_state", "active | revoked"),
    ("invitation_state", "pending | accepted | revoked | expired"),
    ("product_category", "electronics | agriculture | pharma | textiles"),
    ("facility_type", "raw_material_processing | component_fabrication | assembly | distribution"),
    ("facility_verification", "facility_matched | organization_matched | self_declared"),
    ("trust_tier", "new | verified | trusted"),
    ("grid_region", "US-CAISO | US-ERCOT | US-PJM | US-MISO | EU-DE | EU-FR | EU-PL | UK | CN-East"),
    ("", "CN-South | IN-North | JP | KR | TW | VN | MY | SG | TH   (18 total, PRD 2.6.1)"),
    ("treasury_state", "active | superseded"),
    ("treasury_change_state", "pending | completed | cancelled"),
    ("revocation_reason", "membership_revoked | role_changed | session_revoked | admin_action"),
    ("export_state", "processing | ready | expired | failed"),
    ("deletion_state", "requested | blocked | purging | completed"),
    ("plan_tier", "buyer | starter | growth | enterprise"),
    ("subscription_state", "active | grace_period | read_only | cancelled"),
    ("usage_dimension", "batches | checkpoints | facilities | users | ai_reviews | storage_gb"),
    ("invoice_status", "paid | open | failed | void"),
    ("webhook_state", "received | processed | failed"),
    ("idempotency_state", "processing | completed | failed"),
]

ROW_H = 20
HEADER_H = 30
NOTE_H = 26
WIDTH = 330


def build():
    cells = []
    positions = {}
    edges = []

    def table(name, note, columns, x, y, accent):
        rows = BASE + columns
        height = HEADER_H + (NOTE_H if note else 0) + len(rows) * ROW_H
        tid = "t_" + name
        positions[name] = (x, y, height)

        cells.append(
            f'<mxCell id="{tid}" value="{escape(name)}" style="shape=table;startSize={HEADER_H};'
            f'container=1;collapsible=0;childLayout=tableLayout;fixedRows=1;rowLines=0;fontStyle=1;'
            f'align=center;resizeLast=1;html=1;fillColor={accent};strokeColor=#57534E;fontColor=#1C1917;'
            f'fontSize=13;" vertex="1" parent="1">'
            f'<mxGeometry x="{x}" y="{y}" width="{WIDTH}" height="{height}" as="geometry"/></mxCell>'
        )

        offset = HEADER_H
        if note:
            rid = f"{tid}_note"
            cells.append(
                f'<mxCell id="{rid}" value="" style="shape=tableRow;horizontal=0;startSize=0;'
                f'swimlaneHead=0;swimlaneBody=0;fillColor=none;collapsible=0;dropTarget=0;'
                f'points=[[0,0.5],[1,0.5]];portConstraint=eastwest;top=0;left=0;right=0;bottom=0;" '
                f'vertex="1" parent="{tid}">'
                f'<mxGeometry y="{offset}" width="{WIDTH}" height="{NOTE_H}" as="geometry"/></mxCell>'
            )
            cells.append(
                f'<mxCell id="{rid}_c" value="{escape(note)}" style="shape=partialRectangle;'
                f'connectable=0;fillColor=#FAFAF9;align=left;verticalAlign=middle;spacingLeft=8;'
                f'overflow=hidden;html=1;fontSize=10;fontColor=#57534E;fontStyle=2;strokeColor=none;" '
                f'vertex="1" parent="{rid}">'
                f'<mxGeometry width="{WIDTH}" height="{NOTE_H}" as="geometry"/></mxCell>'
            )
            offset += NOTE_H

        for index, (col, coltype, marker) in enumerate(rows):
            rid = f"{tid}_r{index}"
            label = f"{col}   {coltype}"
            if marker:
                label += f"   [{marker}]"
            weight = "fontStyle=1;" if marker.startswith("PK") else ""
            colour = "#F5F5F4" if index < len(BASE) else "none"
            cells.append(
                f'<mxCell id="{rid}" value="" style="shape=tableRow;horizontal=0;startSize=0;'
                f'swimlaneHead=0;swimlaneBody=0;fillColor=none;collapsible=0;dropTarget=0;'
                f'points=[[0,0.5],[1,0.5]];portConstraint=eastwest;top=0;left=0;right=0;bottom=0;" '
                f'vertex="1" parent="{tid}">'
                f'<mxGeometry y="{offset}" width="{WIDTH}" height="{ROW_H}" as="geometry"/></mxCell>'
            )
            cells.append(
                f'<mxCell id="{rid}_c" value="{escape(label)}" style="shape=partialRectangle;'
                f'connectable=0;fillColor={colour};align=left;verticalAlign=middle;spacingLeft=8;'
                f'overflow=hidden;html=1;fontSize=10;{weight}strokeColor=none;" '
                f'vertex="1" parent="{rid}">'
                f'<mxGeometry width="{WIDTH}" height="{ROW_H}" as="geometry"/></mxCell>'
            )
            offset += ROW_H

    def group(label, x, y, w, h, colour):
        cells.append(
            f'<mxCell id="g_{label.replace(" ", "_")}" value="{escape(label)}" '
            f'style="rounded=0;whiteSpace=wrap;html=1;fillColor=none;strokeColor={colour};'
            f'dashed=1;verticalAlign=top;fontSize=16;fontStyle=1;fontColor={colour};align=left;'
            f'spacingLeft=12;spacingTop=6;" vertex="1" parent="1">'
            f'<mxGeometry x="{x}" y="{y}" width="{w}" height="{h}" as="geometry"/></mxCell>'
        )

    columns_x = [60, 440, 820, 1200]
    y_cursor = [120, 120, 120, 120]

    for index, (name, note, cols) in enumerate(IDENTITY):
        slot = index % 4
        table(name, note, cols, columns_x[slot], y_cursor[slot], "#ECFDF5")
        y_cursor[slot] += HEADER_H + (NOTE_H if note else 0) + (len(BASE) + len(cols)) * ROW_H + 40

    identity_bottom = max(y_cursor)
    group("identity schema", 20, 60, 1560, identity_bottom - 60, "#047857")

    billing_top = identity_bottom + 60
    y_cursor = [billing_top + 60] * 4
    for index, (name, note, cols) in enumerate(BILLING):
        slot = index % 4
        table(name, note, cols, columns_x[slot], y_cursor[slot], "#EFF6FF")
        y_cursor[slot] += HEADER_H + (NOTE_H if note else 0) + (len(BASE) + len(cols)) * ROW_H + 40

    billing_bottom = max(y_cursor)
    group("billing schema", 20, billing_top, 1560, billing_bottom - billing_top, "#1D4ED8")

    shared_top = billing_bottom + 60
    y_cursor = [shared_top + 60] * 4
    for index, (name, note, cols) in enumerate(SHARED):
        slot = index % 4
        table(name, note, cols, columns_x[slot], y_cursor[slot], "#FFFBEB")
        y_cursor[slot] += HEADER_H + (NOTE_H if note else 0) + (len(BASE) + len(cols)) * ROW_H + 40

    shared_bottom = max(y_cursor)
    group(
        "shared per-service tables  (one copy inside EVERY service schema)",
        20, shared_top, 1560, shared_bottom - shared_top, "#B45309",
    )

    enum_top = shared_bottom + 60
    enum_height = HEADER_H + len(ENUMS) * ROW_H + 20
    cells.append(
        f'<mxCell id="enum_box" value="postgres ENUM types" style="shape=table;startSize={HEADER_H};'
        f'container=1;collapsible=0;childLayout=tableLayout;fixedRows=1;rowLines=0;fontStyle=1;'
        f'align=center;resizeLast=1;html=1;fillColor=#F5F5F4;strokeColor=#57534E;fontSize=14;" '
        f'vertex="1" parent="1">'
        f'<mxGeometry x="60" y="{enum_top + 40}" width="1000" height="{enum_height}" as="geometry"/></mxCell>'
    )
    offset = HEADER_H
    for index, (enum_name, values) in enumerate(ENUMS):
        rid = f"enum_r{index}"
        label = f"{enum_name}" + (" " * max(1, 30 - len(enum_name))) + values if enum_name else "      " + values
        cells.append(
            f'<mxCell id="{rid}" value="" style="shape=tableRow;horizontal=0;startSize=0;'
            f'swimlaneHead=0;swimlaneBody=0;fillColor=none;collapsible=0;dropTarget=0;'
            f'points=[[0,0.5],[1,0.5]];portConstraint=eastwest;top=0;left=0;right=0;bottom=0;" '
            f'vertex="1" parent="enum_box">'
            f'<mxGeometry y="{offset}" width="1000" height="{ROW_H}" as="geometry"/></mxCell>'
        )
        cells.append(
            f'<mxCell id="{rid}_c" value="{escape(label)}" style="shape=partialRectangle;'
            f'connectable=0;fillColor=none;align=left;verticalAlign=middle;spacingLeft=8;'
            f'overflow=hidden;html=1;fontSize=10;fontFamily=Courier New;strokeColor=none;" '
            f'vertex="1" parent="{rid}">'
            f'<mxGeometry width="1000" height="{ROW_H}" as="geometry"/></mxCell>'
        )
        offset += ROW_H

    group("enum reference", 20, enum_top, 1560, enum_height + 80, "#57534E")

    for index, (source, target) in enumerate(FOREIGN_KEYS):
        if source in positions and target in positions:
            edges.append(
                f'<mxCell id="e{index}" style="edgeStyle=orthogonalEdgeStyle;rounded=1;html=1;'
                f'strokeColor=#57534E;endArrow=ERone;startArrow=ERmany;endFill=0;startFill=0;'
                f'exitX=0;exitY=0.5;entryX=1;entryY=0.5;" edge="1" parent="1" '
                f'source="t_{source}" target="t_{target}"><mxGeometry relative="1" as="geometry"/></mxCell>'
            )

    body = "\n".join(cells + edges)
    return (
        '<mxfile host="app.diagrams.net" agent="carboncircuit" version="24.0.0">\n'
        '  <diagram id="carboncircuit-db" name="identity + billing">\n'
        '    <mxGraphModel dx="1400" dy="900" grid="1" gridSize="10" guides="1" tooltips="1" '
        'connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="1600" pageHeight="4000" '
        'math="0" shadow="0">\n'
        "      <root>\n"
        '        <mxCell id="0"/>\n'
        '        <mxCell id="1" parent="0"/>\n'
        f"{body}\n"
        "      </root>\n"
        "    </mxGraphModel>\n"
        "  </diagram>\n"
        "</mxfile>\n"
    )


if __name__ == "__main__":
    import pathlib
    import sys

    out = pathlib.Path(sys.argv[1])
    out.write_text(build())
    total = len(IDENTITY) + len(BILLING) + len(SHARED)
    print(f"tables: identity={len(IDENTITY)} billing={len(BILLING)} shared={len(SHARED)} total={total}")
    print(f"foreign keys drawn: {len(FOREIGN_KEYS)}")
    print(f"written: {out}")
