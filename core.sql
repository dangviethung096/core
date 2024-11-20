DROP TABLE IF EXISTS scheduler_tasks;

CREATE TABLE scheduler_tasks(
    id serial PRIMARY KEY,
    task_name text,
    queue_name text,
    data bytea,
    done boolean,
    loop_index bigint,
    loop_count bigint,
    next BIGINT,
    next_time text,
    interval bigint,
    start_time text,
    source text
);

DROP TABLE IF EXISTS scheduler_todo;

CREATE TABLE scheduler_todo(
    id bigserial PRIMARY KEY,
    task_id int,
    bucket bigint,
    next_time text,
    source text
);

DROP TABLE IF EXISTS scheduler_done;

CREATE TABLE scheduler_done(
    id bigserial PRIMARY KEY,
    bucket bigint,
    task_id bigint,
    operation_time text,
    status text
);

