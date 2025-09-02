CREATE TABLE sagas
(
    id           VARCHAR(255) NOT NULL PRIMARY KEY,
    current_step INTEGER      NOT NULL,
    status       VARCHAR(50)  NOT NULL,
    payload      JSONB        NOT NULL,
    steps        JSONB        NOT NULL,
    compensating BOOLEAN      NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL,
    updated_at   TIMESTAMPTZ  NOT NULL
);
