CREATE TYPE billing.plan_tier AS ENUM ('buyer', 'starter', 'growth', 'enterprise');
CREATE TYPE billing.organization_type AS ENUM ('manufacturer', 'assembler', 'logistics', 'credit_buyer');
CREATE TYPE billing.subscription_state AS ENUM ('active', 'grace_period', 'read_only', 'cancelled');
CREATE TYPE billing.usage_dimension AS ENUM ('batches', 'checkpoints', 'facilities', 'users', 'ai_reviews', 'storage_gb');
CREATE TYPE billing.invoice_status AS ENUM ('paid', 'open', 'failed', 'void');
CREATE TYPE billing.webhook_state AS ENUM ('received', 'processed', 'failed');
CREATE TYPE billing.idempotency_state AS ENUM ('processing', 'completed', 'failed');
