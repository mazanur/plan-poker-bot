CREATE TABLE room
(
    id           UUID PRIMARY KEY,
    name         VARCHAR,
    user_id      BIGINT,
    chat_id      BIGINT,
    status       VARCHAR,
    created_date TIMESTAMP NOT NULL
);

CREATE TABLE room_member
(
    user_id BIGINT,
    room_id UUID,
    PRIMARY KEY (user_id, room_id)
);

CREATE TABLE task
(
    id           UUID PRIMARY KEY,
    name         VARCHAR   NOT NULL,
    room_id      UUID,
    created_date TIMESTAMP NOT NULL,
    finished     BOOLEAN DEFAULT FALSE,
    grade        INT     DEFAULT 0,
    FOREIGN KEY (room_id) REFERENCES room (id)
);

CREATE TABLE rate
(
    id           UUID PRIMARY KEY,
    task_id      UUID,
    user_id      BIGINT,
    sum          INT       NOT NULL,
    created_date TIMESTAMP NOT NULL,
    FOREIGN KEY (task_id) REFERENCES task (id),
    FOREIGN KEY (user_id) REFERENCES profile (user_id)
);



