ALTER TABLE jobs
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE jobs
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

ALTER TABLE jobs
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE jobs
    ADD CONSTRAINT jobs_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_jobs_tenant_pending ON jobs(tenant_id, run_at)
WHERE status = 'pending';

ALTER TABLE dead_jobs
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE dead_jobs
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

ALTER TABLE dead_jobs
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE dead_jobs
    ADD CONSTRAINT dead_jobs_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_dead_jobs_tenant_created ON dead_jobs(tenant_id, created_at DESC);
