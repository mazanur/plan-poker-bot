CREATE TABLE button
(
    id     UUID PRIMARY KEY,
    action VARCHAR NOT NULL,
    data   JSONB   NOT NULL
);

CREATE TABLE chat_info
(
    chat_id           BIGINT PRIMARY KEY,
    active_chain      VARCHAR NOT NULL,
    active_chain_step VARCHAR NOT NULL,
    chain_data        JSONB   NOT NULL
);

CREATE TABLE profile
(
    user_id      BIGINT PRIMARY KEY,
    user_name    VARCHAR NOT NULL,
    display_name VARCHAR NOT NULL
);

