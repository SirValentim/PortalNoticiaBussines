CREATE TABLE IF NOT EXISTS tenant_users (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(40) NOT NULL DEFAULT 'editor',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, user_id)
);

INSERT INTO tenant_users (tenant_id, user_id, role, active, created_at, updated_at)
SELECT tenant_id, id, role, active, NOW(), NOW()
FROM users
ON CONFLICT (tenant_id, user_id) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_tenant_users_user ON tenant_users(user_id, active);
CREATE INDEX IF NOT EXISTS idx_tenant_users_tenant_role ON tenant_users(tenant_id, role, active);
