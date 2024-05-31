CREATE TABLE IF NOT EXISTS groups (
    id varchar(255) PRIMARY KEY,
    group_name varchar(200) NOT NULL UNIQUE,
    conv_id varchar(255) REFERENCES conversations (id),
    created_at timestamp DEFAULT current_timestamp,
    last_updated timestamp DEFAULT current_timestamp,
    admins uuid[] NOT NULL check (array_length(admins, 1) > 0),
    deleted bool DEFAULT false
);

CREATE UNIQUE INDEX group_name_idx ON groups(group_name);

CREATE TABLE IF NOT EXISTS conversations (
    id varchar(255) PRIMARY KEY,
);

CREATE TABLE IF NOT EXISTS conv_map_user (
    id varchar(255) PRIMARY KEY,
    conv_id varchar(255) REFERENCES conversations(id),
    user_id varchar(255) REFERENCES users(id)
);
