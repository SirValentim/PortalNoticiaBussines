ALTER TABLE influencers
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE influencers
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

ALTER TABLE influencers
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE influencers
    DROP CONSTRAINT IF EXISTS influencers_slug_key;

ALTER TABLE influencers
    ADD CONSTRAINT influencers_tenant_slug_key UNIQUE (tenant_id, slug);

ALTER TABLE influencers
    ADD CONSTRAINT influencers_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_influencers_tenant_active ON influencers(tenant_id, active, is_featured, created_at DESC);
