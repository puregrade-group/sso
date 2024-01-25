CREATE TABLE IF NOT EXISTS users (
    id BLOB PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    pass_hash BLOB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_email ON users (email);

CREATE TABLE IF NOT EXISTS apps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    secret TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    description TEXT NOT NULL,
    CONSTRAINT resource_action_uq UNIQUE (resource, action)
);

CREATE TABLE IF NOT EXISTS roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL
);

-- -- The "app_permissions" table links applications and permissions:
-- -- each row represents a specific permission for a specific application.
-- CREATE TABLE IF NOT EXISTS app_permissions (
--     id INTEGER PRIMARY KEY,
--     app_id INTEGER,
--     permission_id INTEGER,
--     CONSTRAINT app_permissions_uq UNIQUE (app_id, permission_id),
--     FOREIGN KEY (app_id) REFERENCES apps (id) ON DELETE CASCADE,
--     FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE
-- );

-- Each row in the "user_permissions" table indicates what permission a specific user has for a specific service.
CREATE TABLE IF NOT EXISTS roles_permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    role_id INTEGER,
    permission_id INTEGER,
    CONSTRAINT role_permission_uq UNIQUE (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS users_roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    role_id INTEGER,
    CONSTRAINT user_role_uq UNIQUE (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE
);

INSERT INTO permissions (resource, action, description) VALUES
    -- Permissions to act on permissions
    ('permission', 'create', 'Permission to create new permissions'),
    ('permission', 'delete', 'Permission to delete permissions'),
    ('permission', 'grant', 'Permission to grant permissions to roles/users'),
    ('permission', 'revoke', 'Permission to revoke permissions from roles/users'),
    -- Permissions to act on roles
    ('roles', 'create', 'Permission to create new roles'),
    ('roles', 'delete', 'Permission to delete roles'),
    ('roles', 'grant', 'Permission to grant roles to users'),
    ('roles', 'revoke', 'Permission to revoke roles from users'),
    -- Permissions to act on users
    ('user', 'create', 'Permission to create new users'),
    ('user', 'read', 'Permission to read user info'),
    ('user', 'update', 'Permission to update users'),
    ('user', 'delete', 'Permission to delete users')
ON CONFLICT DO NOTHING;

-- Creating the "ADMIN" role
INSERT INTO roles (name, description) VALUES ('ADMIN', 'desc');

-- Adding all existing permissions to the "ADMIN" role
INSERT INTO roles_permissions (role_id, permission_id)
SELECT 1, id
FROM permissions;