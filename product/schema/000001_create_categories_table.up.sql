CREATE TABLE categories
(
    id         VARCHAR(255) PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    created_at TIMESTAMP    NOT NULL
);

INSERT INTO categories (id, name, created_at)
VALUES ('cat-1', 'cat-1', NOW());
INSERT INTO categories (id, name, created_at)
VALUES ('cat-2', 'cat-2', NOW());
INSERT INTO categories (id, name, created_at)
VALUES ('cat-3', 'cat-3', NOW());
