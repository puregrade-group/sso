create table if not exists profiles (
    id bigint primary key,
    first_name varchar(50) not null,
    last_name varchar(50),
    date_of_birth date,
    created_at timestamp,
    updated_at timestamp
);

create table if not exists credentials (
    id bigint primary key,
    email varchar(256) not null unique,
    pass_hash text not null, -- password (salted)
    foreign key (id) references profiles (id) on delete cascade
);

create index if not exists idx_email on credentials (email);

create table if not exists refresh_tokens (
    value varchar(48) primary key,
    user_id bigint not null unique,
    expires_in timestamp not null,
    foreign key (user_id) references profiles (id) on delete cascade
);

-- CREATE TABLE IF NOT EXISTS apps (
--     id INTEGER PRIMARY KEY AUTOINCREMENT,
--     name TEXT NOT NULL UNIQUE,
--     secret TEXT NOT NULL UNIQUE,
--     redirect_uri text not null,
--     user_id BIGINT NOT NULL,
--     foreign key (user_id) references users (id) ON DELETE CASCADE
-- );
--
-- CREATE TABLE IF NOT EXISTS permissions (
--     id INTEGER PRIMARY KEY AUTOINCREMENT,
--     resource TEXT NOT NULL,
--     action TEXT NOT NULL,
--     description TEXT NOT NULL,
--     CONSTRAINT resource_action_uq UNIQUE (resource, action)
-- );
--
-- -- The "app_permissions" table links applications and permissions:
-- -- each row represents a specific permission for a specific application.
-- CREATE TABLE IF NOT EXISTS app_permissions (
--     app_id int references apps (id) in delete cascade,
--     permission_id int references permissions (id) in delete cascade,
--     CONSTRAINT apps_permissions_pk primary key (app_id, permission_id),
-- );
--
-- INSERT INTO permissions (resource, action, description) VALUES
--     -- Permissions to act on permissions
--     ('permission', 'create', 'Permission to create new permissions'),
--     ('permission', 'delete', 'Permission to delete permissions'),
--     -- Permissions to act on app permissions
--     ('permission', 'grant', 'Permission to grant permissions to app'),
--     ('permission', 'revoke', 'Permission to revoke permissions from app'),
--     -- Permissions to act on users
--     ('user', 'create', 'Permission to create new users'),
--     ('user', 'read', 'Permission to read user info'),
--     ('user', 'update', 'Permission to update users'),
--     ('user', 'delete', 'Permission to delete users'),
--     -- Permissions to act on user permissions
--     ('user.permission', 'grant', 'Permission to grant permissions to user')
--     ('user.permission', 'revoke', 'Permission to revoke permissions from user')
-- ON CONFLICT DO NOTHING;