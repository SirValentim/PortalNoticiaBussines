ALTER TABLE tags
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE tags
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

ALTER TABLE tags
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE tags
    DROP CONSTRAINT IF EXISTS tags_slug_key;

ALTER TABLE tags
    ADD CONSTRAINT tags_tenant_slug_key UNIQUE (tenant_id, slug);

ALTER TABLE tags
    ADD CONSTRAINT tags_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_tags_tenant_active ON tags(tenant_id, active, name);
