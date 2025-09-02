CREATE TABLE methods
(
    id           VARCHAR(255) PRIMARY KEY,
    user_id      VARCHAR(255) NOT NULL,
    gateway      VARCHAR(255) NOT NULL,
    payment_type VARCHAR(255) NOT NULL,
    token        VARCHAR(255) NOT NULL
);
