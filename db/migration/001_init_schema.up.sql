CREATE TABLE tasks (
  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  task_code text,
  task_notes text,
  start_date DATETIME DEFAULT CURRENT_TIMESTAMP,
  completed_date DATETIME DEFAULT NULL
);

