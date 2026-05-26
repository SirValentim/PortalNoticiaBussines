ALTER TABLE portal_settings
    ADD COLUMN IF NOT EXISTS tenant_id BIGINT;

UPDATE portal_settings
SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default' LIMIT 1)
WHERE tenant_id IS NULL;

ALTER TABLE portal_settings
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE portal_settings
    DROP CONSTRAINT IF EXISTS portal_settings_id_check;

ALTER TABLE portal_settings
    DROP CONSTRAINT IF EXISTS portal_settings_pkey;

CREATE SEQUENCE IF NOT EXISTS portal_settings_id_seq;

SELECT setval(
    'portal_settings_id_seq',
    GREATEST(COALESCE((SELECT MAX(id) FROM portal_settings), 0), 1),
    true
);

ALTER TABLE portal_settings
    ALTER COLUMN id SET DEFAULT nextval('portal_settings_id_seq');

ALTER SEQUENCE portal_settings_id_seq
    OWNED BY portal_settings.id;

ALTER TABLE portal_settings
    ADD CONSTRAINT portal_settings_pkey PRIMARY KEY (id);

ALTER TABLE portal_settings
    ADD CONSTRAINT portal_settings_tenant_id_key UNIQUE (tenant_id);

ALTER TABLE portal_settings
    ADD CONSTRAINT portal_settings_tenant_id_fkey
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
