ALTER TABLE posts
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE posts
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

ALTER TABLE posts
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE posts
    DROP CONSTRAINT IF EXISTS posts_slug_key;

ALTER TABLE posts
    ADD CONSTRAINT posts_tenant_slug_key UNIQUE (tenant_id, slug);

ALTER TABLE posts
    ADD CONSTRAINT posts_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_posts_tenant_status ON posts(tenant_id, status, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_tenant_category ON posts(tenant_id, category_id, status);
CREATE INDEX IF NOT EXISTS idx_posts_tenant_featured ON posts(tenant_id, is_featured, status, updated_at DESC);
