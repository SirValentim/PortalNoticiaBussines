ALTER TABLE events
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE events
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

ALTER TABLE events
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE events
    DROP CONSTRAINT IF EXISTS events_slug_key;

ALTER TABLE events
    ADD CONSTRAINT events_tenant_slug_key UNIQUE (tenant_id, slug);

ALTER TABLE events
    ADD CONSTRAINT events_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_events_tenant_status_start ON events(tenant_id, status, start_at);
CREATE INDEX IF NOT EXISTS idx_events_tenant_featured ON events(tenant_id, is_featured, status, start_at);
