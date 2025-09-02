CREATE TABLE methods
(
    id           VARCHAR(255) PRIMARY KEY,
    user_id      VARCHAR(255) NOT NULL,
    gateway      VARCHAR(255) NOT NULL,
    payment_type VARCHAR(255) NOT NULL,
    token        VARCHAR(255) NOT NULL
);

INSERT INTO methods (id, user_id, gateway, payment_type, token) VALUES ('method-1', 'user-1', 'sber', 'card', 'token-1');
INSERT INTO methods (id, user_id, gateway, payment_type, token) VALUES ('method-2', 'user-1', 'tinkoff', 'card', 'token-2');
INSERT INTO methods (id, user_id, gateway, payment_type, token) VALUES ('method-fail', 'user-1', 'fail', 'card', 'token-fail');
