CREATE TABLE inventory
(
    product_id VARCHAR(255) PRIMARY KEY,
    quantity   INTEGER   NOT NULL CHECK ( quantity >= 0 ),
--     reserved_quantity INTEGER   NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

INSERT INTO inventory (product_id, quantity, created_at, updated_at)
VALUES ('product-1', 10000, NOW(), NOW());

INSERT INTO inventory (product_id, quantity, created_at, updated_at)
VALUES ('product-2', 20000, NOW(), NOW());

INSERT INTO inventory (product_id, quantity, created_at, updated_at)
VALUES ('product-3', 30000, NOW(), NOW());
