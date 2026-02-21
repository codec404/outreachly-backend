CREATE TABLE role_permissions (
    role_id       UUID        NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID        NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role_id       ON role_permissions (role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions (permission_id);

-- Seed: admin receives every permission.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM   roles r, permissions p
WHERE  r.name = 'admin';

-- Seed: user receives standard self-service permissions.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM   roles r, permissions p
WHERE  r.name = 'user'
  AND  p.name IN (
      'create_campaign',
      'view_campaign',
      'view_templates',
      'create_template',
      'manage_recruiters',
      'view_analytics'
  );
