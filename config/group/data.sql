CREATE TABLE IF NOT EXISTS groups (
    id uuid PRIMARY KEY,
    group_name varchar(200) NOT NULL UNIQUE,
    conv_id uuid REFERENCES conversations (id),
    created_at timestamp DEFAULT current_timestamp,
    last_updated timestamp DEFAULT current_timestamp,
    admins uuid[] NOT NULL check (array_length(admins, 1) > 0),
    deleted bool DEFAULT false
);

CREATE UNIQUE INDEX group_name_idx ON groups(group_name);

CREATE TABLE IF NOT EXISTS conversations (
    id uuid PRIMARY KEY,
);

CREATE TABLE IF NOT EXISTS conv_map_user (
    id uuid PRIMARY KEY,
    conv_id uuid REFERENCES conversations(id),
    user_id uuid REFERENCES users(id)
);
