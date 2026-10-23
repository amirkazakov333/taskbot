CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,

    telegram_id BIGINT NOT NULL UNIQUE,
    username TEXT,

    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    preferred_name TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
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
        CHECK (
            category IN ('work', 'personal')
        ),

    CONSTRAINT tasks_status_check
        CHECK (
            status IN ('todo', 'done', 'cancelled')
        )
);


CREATE TABLE IF NOT EXISTS user_sessions (
    telegram_id BIGINT PRIMARY KEY,

    session_type VARCHAR(30) NOT NULL
        CHECK (
            session_type IN (
                'registration',
                'task_creation',
                'task_reschedule'
            )
        ),

    username TEXT,

    first_name TEXT NOT NULL DEFAULT '',
    last_name TEXT NOT NULL DEFAULT '',
    preferred_name TEXT NOT NULL DEFAULT '',

    category VARCHAR(20),
    title TEXT,
    task_date DATE,

    target_task_id BIGINT
        REFERENCES tasks(id)
        ON DELETE CASCADE,

    current_step TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT user_sessions_category_check
        CHECK (
            category IS NULL
            OR category IN ('work', 'personal')
        ),

    CONSTRAINT user_sessions_target_task_check
        CHECK (
            (
                session_type = 'task_reschedule'
                AND target_task_id IS NOT NULL
            )
            OR
            (
                session_type IN (
                    'registration',
                    'task_creation'
                )
                AND target_task_id IS NULL
            )
        )
);


CREATE INDEX IF NOT EXISTS idx_tasks_user_category_status
    ON tasks (user_id, category, status);


CREATE INDEX IF NOT EXISTS idx_tasks_due_todo
    ON tasks (due_at)
    WHERE status = 'todo';