CREATE TABLE products
(
    id          VARCHAR(255) PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    price       INTEGER      NOT NULL,
    category_id VARCHAR(255) NOT NULL
);

INSERT INTO products (id, name, price, category_id) VALUES ('product-1', 'product-1', 10000, 'cat-1');
INSERT INTO products (id, name, price, category_id) VALUES ('product-2', 'product-2', 20000, 'cat-2');
INSERT INTO products (id, name, price, category_id) VALUES ('product-3', 'product-3', 30000, 'cat-3');
