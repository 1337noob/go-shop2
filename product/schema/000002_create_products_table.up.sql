CREATE TABLE products
(
    id          VARCHAR(255) PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    price       INTEGER      NOT NULL,
    category_id VARCHAR(255) NOT NULL
);