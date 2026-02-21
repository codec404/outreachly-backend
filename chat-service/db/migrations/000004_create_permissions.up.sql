CREATE TABLE permissions (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Seed all system permissions.
INSERT INTO permissions (name, description) VALUES
    ('create_campaign',   'Create new outreach campaigns'),
    ('view_campaign',     'View own campaigns and their stats'),
    ('delete_campaign',   'Delete campaigns'),
    ('view_templates',    'View email templates'),
    ('create_template',   'Create and edit email templates'),
    ('delete_template',   'Delete email templates'),
    ('manage_recruiters', 'Add, edit, and remove recruiter contacts'),
    ('view_analytics',    'View campaign performance analytics'),
    ('manage_users',      'Manage user accounts (admin only)'),
    ('manage_roles',      'Manage roles and permission assignments (admin only)');
