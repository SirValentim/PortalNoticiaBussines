ALTER TABLE banners
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE banners
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

ALTER TABLE banners
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE banners
    ADD CONSTRAINT banners_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_banners_tenant_position ON banners(tenant_id, position, active, start_date, end_date);
