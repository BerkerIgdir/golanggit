-- name: CreateTask :one
INSERT INTO tasks (
    task_code
) VALUES (
    ?
) RETURNING *;

-- name: UpdateTaskNote :one
UPDATE tasks SET task_notes=? WHERE task_code=? RETURNING *;

-- name: FinishTask :one
UPDATE tasks SET completed_date=? WHERE task_code=? RETURNING *;
