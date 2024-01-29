DROP TABLE IF EXISTS scheduler_tasks;

CREATE TABLE scheduler_tasks (
    id SERIAL PRIMARY KEY,
    queue_name TEXT,
    data BYTEA,
    done BOOLEAN,
    loop_index BIGINT,
    loop_count BIGINT,
    next BIGINT,
    next_time TEXT,
    interval BIGINT,
    start_time TEXT,
    source TEXT
);

DROP TABLE IF EXISTS scheduler_todo;

CREATE TABLE scheduler_todo (
    id BIGSERIAL PRIMARY KEY,
    task_id INT,
    bucket BIGINT,
    next_time TEXT,
    source TEXT
);

DROP TABLE IF EXISTS scheduler_done;

CREATE TABLE scheduler_done (
    id BIGSERIAL PRIMARY KEY,
    bucket BIGINT,
    task_id BIGINT,
    operation_time TEXT,
    status TEXT
);