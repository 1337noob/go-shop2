CREATE TABLE order_history
(
    id                  VARCHAR(255) PRIMARY KEY,
    user_id             VARCHAR(255) NOT NULL,
    order_items         JSONB        NOT NULL,
    payment_id          VARCHAR(255),
    payment_method_id   VARCHAR(255) NOT NULL,
    payment_type        VARCHAR(255),
    payment_gateway     VARCHAR(255),
    payment_sum         INTEGER,
    payment_external_id VARCHAR(255),
    payment_status      VARCHAR(255),
    status              VARCHAR(255) NOT NULL,
    created_at          TIMESTAMP    NOT NULL,
    updated_at          TIMESTAMP    NOT NULL
);
