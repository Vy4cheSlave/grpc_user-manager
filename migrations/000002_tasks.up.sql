CREATE TABLE tasks (
   id SERIAL PRIMARY KEY,             -- Уникальный идентификатор задачи
   user_id UUID REFERENCES users(id) ON DELETE CASCADE,
   title TEXT NOT NULL,               -- Заголовок задачи
   description TEXT,                  -- Описание задачи (необязательное поле)
   status TEXT CHECK (status IN ('new', 'in_progress', 'done')) DEFAULT 'new', -- Статус задачи
   created_at TIMESTAMP DEFAULT now(), -- Время создания задачи
   updated_at TIMESTAMP DEFAULT now()  -- Время последнего обновления задачи
);

CREATE INDEX idx_tasks_user_id ON tasks(user_id);