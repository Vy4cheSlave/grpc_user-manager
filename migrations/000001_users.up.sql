CREATE TABLE users (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Уникальный идентификатор
   email VARCHAR(255) UNIQUE NOT NULL, -- Основной идентификатор
   username VARCHAR(50) UNIQUE NOT NULL, -- Публичное имя
   password_hash TEXT NOT NULL, -- Хэш пароля
   first_name VARCHAR(100), -- Имя
   last_name VARCHAR(100), -- Фамилия
   is_active BOOLEAN DEFAULT TRUE, -- Активен
   role VARCHAR(50) DEFAULT 'user', -- Роль
   last_login_at TIMESTAMP WITH TIME ZONE, -- Последний вход
   created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(), -- Дата создания
   updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() -- Дата обновления
);
