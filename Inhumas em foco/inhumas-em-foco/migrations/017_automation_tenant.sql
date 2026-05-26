ALTER TABLE automation_sources
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE automation_sources
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

ALTER TABLE automation_sources
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE automation_sources
    ADD CONSTRAINT automation_sources_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

DROP INDEX IF EXISTS idx_automation_sources_active;
CREATE INDEX IF NOT EXISTS idx_automation_sources_tenant_active ON automation_sources(tenant_id, active, source_type);

ALTER TABLE automation_runs
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE automation_runs
SET tenant_id = COALESCE(
    (SELECT tenant_id FROM automation_sources WHERE automation_sources.id = automation_runs.source_id),
    (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
)
WHERE tenant_id IS NULL;

ALTER TABLE automation_runs
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE automation_runs
    ADD CONSTRAINT automation_runs_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

DROP INDEX IF EXISTS idx_automation_runs_started;
CREATE INDEX IF NOT EXISTS idx_automation_runs_tenant_started ON automation_runs(tenant_id, started_at DESC);
