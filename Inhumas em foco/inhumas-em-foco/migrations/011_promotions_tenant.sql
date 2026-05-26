ALTER TABLE promotions
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE promotions
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

ALTER TABLE promotions
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE promotions
    DROP CONSTRAINT IF EXISTS promotions_slug_key;

ALTER TABLE promotions
    ADD CONSTRAINT promotions_tenant_slug_key UNIQUE (tenant_id, slug);

ALTER TABLE promotions
    ADD CONSTRAINT promotions_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_promotions_tenant_active ON promotions(tenant_id, status, end_date);
