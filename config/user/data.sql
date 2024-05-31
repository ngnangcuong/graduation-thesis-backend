CREATE TABLE IF NOT EXISTS users {
    id varchar(255) PRIMARY KEY,
    username varchar(255) NOT NULL UNIQUE,
    password varchar(255) NOT NULL,
    first_name varchar(255),
    last_name varchar(255),
    email varchar(255) NOT NULL UNIQUE,
    phone_number varchar(255),
    created_at timestamp DEFAULT current_timestamp,
    last_updated timestamp DEFAULT current_timestamp,
    avatar varchar(255)
};
CREATE UNIQUE INDEX username_idx ON users (username);
CREATE UNIQUE INDEX email_idx ON users (email);
