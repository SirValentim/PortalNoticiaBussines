ALTER TABLE classifieds
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE classifieds
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

ALTER TABLE classifieds
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE classifieds
    DROP CONSTRAINT IF EXISTS classifieds_slug_key;

ALTER TABLE classifieds
    ADD CONSTRAINT classifieds_tenant_slug_key UNIQUE (tenant_id, slug);

ALTER TABLE classifieds
    ADD CONSTRAINT classifieds_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_classifieds_tenant_status_category ON classifieds(tenant_id, status, category, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_classifieds_tenant_featured ON classifieds(tenant_id, is_featured, status, created_at DESC);
