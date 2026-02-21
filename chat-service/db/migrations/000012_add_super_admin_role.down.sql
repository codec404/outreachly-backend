DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE name = 'super_admin');

DELETE FROM roles WHERE name = 'super_admin';

DELETE FROM permissions WHERE name = 'manage_admins';
