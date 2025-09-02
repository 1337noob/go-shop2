CREATE TABLE inventory
(
    product_id VARCHAR(255) PRIMARY KEY,
    quantity   INTEGER    NOT NULL,
--     reserved_quantity INTEGER   NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
