CREATE TABLE inbox
(
--     id           VARCHAR(255) PRIMARY KEY,
    message_id   VARCHAR(255) PRIMARY KEY,
    message_type VARCHAR(255) NOT NULL,
    topic        VARCHAR(255) NOT NULL,
    key          VARCHAR(255) NOT NULL,
    payload      BYTEA        NOT NULL,
    status       VARCHAR(255) NOT NULL,
    created_at   TIMESTAMP    NOT NULL
);

CREATE INDEX inbox_status_index ON inbox (status);
