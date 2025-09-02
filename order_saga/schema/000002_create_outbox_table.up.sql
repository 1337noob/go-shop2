CREATE TABLE outbox
(
    id         VARCHAR(255) PRIMARY KEY,
--     event_id   VARCHAR(255) NOT NULL,
    topic      VARCHAR(255) NOT NULL,
    key        VARCHAR(255) NOT NULL,
    payload    BYTEA        NOT NULL,
    status     VARCHAR(255) NOT NULL,
    created_at TIMESTAMP    NOT NULL
);

CREATE INDEX outbox_status_index ON outbox (status);