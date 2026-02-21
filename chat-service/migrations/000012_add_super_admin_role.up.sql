-- manage_admins is intentionally excluded from the admin role.
-- Only super_admin can promote or demote admins.
INSERT INTO permissions (name, description) VALUES
    ('manage_admins', 'Promote or demote admin accounts');

INSERT INTO roles (name, description) VALUES
    ('super_admin', 'Root administrator; the only role that can manage admin accounts');

-- super_admin receives every permission, including manage_admins.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM   roles r, permissions p
WHERE  r.name = 'super_admin';
