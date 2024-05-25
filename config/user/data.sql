CREATE TABLE IF NOT EXISTS users {
    id uuid PRIMARY KEY,
    username varchar(20) NOT NULL UNIQUE,
    password varchar(100) NOT NULL,
    first_name varchar(20),
    last_name varchar(20),
    email varchar(100) NOT NULL UNIQUE,
    phone_number varchar(20),
    created_at timestamp DEFAULT current_timestamp,
    last_updated timestamp DEFAULT current_timestamp,
    avatar varchar(200)
};
CREATE UNIQUE INDEX username_idx ON users (username);
CREATE UNIQUE INDEX email_idx ON users (email);
