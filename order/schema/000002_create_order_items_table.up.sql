CREATE TABLE order_items
(
    id         VARCHAR(255) PRIMARY KEY,
    order_id   VARCHAR(255) NOT NULL,
    product_id VARCHAR(255) NOT NULL,
    quantity   INTEGER      NOT NULL,
    created_at TIMESTAMP    NOT NULL,
    updated_at TIMESTAMP    NOT NULL
);
