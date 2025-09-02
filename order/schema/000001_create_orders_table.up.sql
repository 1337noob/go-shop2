CREATE TABLE orders
(
    id                VARCHAR(255) PRIMARY KEY,
    user_id           VARCHAR(255) NOT NULL,
    payment_method_id VARCHAR(255) NOT NULL,
    phone             VARCHAR(255) NOT NULL,
    email             VARCHAR(255) NOT NULL,
    status            VARCHAR(255) NOT NULL,
    created_at        TIMESTAMP    NOT NULL,
    updated_at        TIMESTAMP    NOT NULL
);
