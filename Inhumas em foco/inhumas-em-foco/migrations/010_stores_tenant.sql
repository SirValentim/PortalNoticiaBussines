ALTER TABLE stores
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE stores
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

ALTER TABLE stores
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE stores
    DROP CONSTRAINT IF EXISTS stores_slug_key;

ALTER TABLE stores
    ADD CONSTRAINT stores_tenant_slug_key UNIQUE (tenant_id, slug);

ALTER TABLE stores
    ADD CONSTRAINT stores_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_stores_tenant_active ON stores(tenant_id, active, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_stores_tenant_featured ON stores(tenant_id, active, is_featured, created_at DESC);
