CREATE TABLE payments
(
    id          VARCHAR(255) PRIMARY KEY,
    order_id    VARCHAR(255) NOT NULL,
    user_id     VARCHAR(255) NOT NULL,
    amount      INTEGER      NOT NULL,
    external_id VARCHAR(255),
    status      VARCHAR(255) NOT NULL,
    method_id   VARCHAR(255) NOT NULL,
    created_at  TIMESTAMP    NOT NULL,
    updated_at  TIMESTAMP    NOT NULL
);
