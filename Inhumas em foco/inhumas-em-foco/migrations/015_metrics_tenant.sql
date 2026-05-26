ALTER TABLE metrics
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE metrics
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

ALTER TABLE metrics
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE metrics
    ADD CONSTRAINT metrics_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_metrics_tenant_entity ON metrics(tenant_id, entity_type, entity_id, metric_type);
CREATE INDEX IF NOT EXISTS idx_metrics_tenant_date ON metrics(tenant_id, created_at);
