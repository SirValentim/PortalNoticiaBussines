ALTER TABLE media_assets
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE media_assets
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

ALTER TABLE media_assets
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE media_assets
    DROP CONSTRAINT IF EXISTS media_assets_key_key;

ALTER TABLE media_assets
    ADD CONSTRAINT media_assets_tenant_key_key UNIQUE (tenant_id, key);

ALTER TABLE media_assets
    ADD CONSTRAINT media_assets_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_media_assets_tenant_created ON media_assets(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_media_assets_tenant_search ON media_assets(tenant_id, original_name, title, alt_text);
