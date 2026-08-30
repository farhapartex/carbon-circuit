CREATE TYPE identity.organization_type AS ENUM ('manufacturer', 'assembler', 'logistics', 'credit_buyer');
CREATE TYPE identity.organization_state AS ENUM ('active', 'restricted', 'read_only', 'suspended');
CREATE TYPE identity.verification_status AS ENUM ('verified', 'unverified', 'rejected');
CREATE TYPE identity.verification_source AS ENUM ('registry', 'manual_override');
CREATE TYPE identity.registry_rejection AS ENUM ('entity_dissolved', 'sanctions_flag', 'name_mismatch');
CREATE TYPE identity.registry_entity_status AS ENUM ('active', 'dissolved');
CREATE TYPE identity.organization_role AS ENUM ('owner', 'admin', 'member');
CREATE TYPE identity.platform_role AS ENUM ('verifier', 'platform_admin');
CREATE TYPE identity.membership_state AS ENUM ('active', 'revoked');
CREATE TYPE identity.invitation_state AS ENUM ('pending', 'accepted', 'revoked', 'expired');
CREATE TYPE identity.product_category AS ENUM ('electronics', 'agriculture', 'pharma', 'textiles');
CREATE TYPE identity.facility_type AS ENUM ('raw_material_processing', 'component_fabrication', 'assembly', 'distribution');
CREATE TYPE identity.facility_verification AS ENUM ('facility_matched', 'organization_matched', 'self_declared');
CREATE TYPE identity.trust_tier AS ENUM ('new', 'verified', 'trusted');
CREATE TYPE identity.grid_region AS ENUM (
    'US-CAISO', 'US-ERCOT', 'US-PJM', 'US-MISO',
    'EU-DE', 'EU-FR', 'EU-PL', 'UK',
    'CN-East', 'CN-South', 'IN-North', 'JP', 'KR', 'TW', 'VN', 'MY', 'SG', 'TH'
);
CREATE TYPE identity.treasury_state AS ENUM ('active', 'superseded');
CREATE TYPE identity.treasury_change_state AS ENUM ('pending', 'completed', 'cancelled');
CREATE TYPE identity.revocation_reason AS ENUM ('membership_revoked', 'role_changed', 'session_revoked', 'admin_action');
CREATE TYPE identity.export_state AS ENUM ('processing', 'ready', 'expired', 'failed');
CREATE TYPE identity.deletion_state AS ENUM ('requested', 'blocked', 'purging', 'completed');
CREATE TYPE identity.idempotency_state AS ENUM ('processing', 'completed', 'failed');
