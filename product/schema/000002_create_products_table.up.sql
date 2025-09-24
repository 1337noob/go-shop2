CREATE TABLE products
(
    id          VARCHAR(255) PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    price       INTEGER      NOT NULL,
    category_id VARCHAR(255) NOT NULL,
    created_at  TIMESTAMP    NOT NULL
);

INSERT INTO products (id, name, price, category_id, created_at)
VALUES ('product-1', 'product-1', 10000, 'cat-1', NOW());
INSERT INTO products (id, name, price, category_id, created_at)
VALUES ('product-11', 'product-11', 10000, 'cat-1', NOW());
INSERT INTO products (id, name, price, category_id, created_at)
VALUES ('product-12', 'product-11', 10000, 'cat-1', NOW());
INSERT INTO products (id, name, price, category_id, created_at)
VALUES ('product-13', 'product-13', 10000, 'cat-1', NOW());
INSERT INTO products (id, name, price, category_id, created_at)
VALUES ('product-2', 'product-2', 20000, 'cat-2', NOW());
INSERT INTO products (id, name, price, category_id, created_at)
VALUES ('product-3', 'product-3', 30000, 'cat-3', NOW());
