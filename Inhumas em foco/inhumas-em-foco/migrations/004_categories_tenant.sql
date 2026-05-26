ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE categories
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

ALTER TABLE categories
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE categories
    DROP CONSTRAINT IF EXISTS categories_slug_key;

ALTER TABLE categories
    ADD CONSTRAINT categories_tenant_slug_key UNIQUE (tenant_id, slug);

ALTER TABLE categories
    ADD CONSTRAINT categories_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_categories_tenant_active ON categories(tenant_id, active, sort_order);
