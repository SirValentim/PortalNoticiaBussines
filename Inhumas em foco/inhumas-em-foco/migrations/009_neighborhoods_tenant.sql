ALTER TABLE neighborhoods
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE neighborhoods
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

ALTER TABLE neighborhoods
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE neighborhoods
    DROP CONSTRAINT IF EXISTS neighborhoods_slug_key;

ALTER TABLE neighborhoods
    ADD CONSTRAINT neighborhoods_tenant_slug_key UNIQUE (tenant_id, slug);

ALTER TABLE neighborhoods
    ADD CONSTRAINT neighborhoods_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_neighborhoods_tenant_name ON neighborhoods(tenant_id, name);
