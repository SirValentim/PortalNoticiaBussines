ALTER TABLE users
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE users
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

ALTER TABLE users
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_email_key;

ALTER TABLE users
    ADD CONSTRAINT users_tenant_email_key UNIQUE (tenant_id, email);

ALTER TABLE users
    ADD CONSTRAINT users_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_users_tenant_role ON users(tenant_id, role, active);
