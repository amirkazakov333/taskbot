CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,

    telegram_id BIGINT NOT NULL UNIQUE,
    username TEXT,

    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    preferred_name TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS registration_sessions (
    telegram_id BIGINT PRIMARY KEY,

    username TEXT,

    first_name TEXT NOT NULL DEFAULT '',
    last_name TEXT NOT NULL DEFAULT '',
    preferred_name TEXT NOT NULL DEFAULT '',

    current_step TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS tasks (
    id BIGSERIAL PRIMARY KEY,

    user_id BIGINT NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    category VARCHAR(20) NOT NULL,
    title TEXT NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'todo',

    due_at TIMESTAMPTZ NOT NULL,

    pre_reminder_sent BOOLEAN NOT NULL DEFAULT FALSE,
    due_reminder_sent BOOLEAN NOT NULL DEFAULT FALSE,

    last_overdue_reminder_at TIMESTAMPTZ,

    completed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT tasks_category_check
        CHECK (category IN ('work', 'personal')),

    CONSTRAINT tasks_status_check
        CHECK (status IN ('todo', 'done'))
);


CREATE TABLE IF NOT EXISTS task_creation_sessions (
    telegram_id BIGINT PRIMARY KEY,
    category VARCHAR(20),
    title TEXT,
    task_date DATE,
    current_step TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT task_creation_category_check
        CHECK (
            category IS NULL
            OR category IN ('work', 'personal')
        )
);


CREATE INDEX IF NOT EXISTS idx_tasks_user_category_status
    ON tasks (user_id, category, status);


CREATE INDEX IF NOT EXISTS idx_tasks_due_todo
    ON tasks (due_at)
    WHERE status = 'todo';