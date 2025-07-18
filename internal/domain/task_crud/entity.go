package task_crud

import "time"

type Task struct {
	Id          int       `json:"id" mapstructure:"id,omitempty"`
	UserId      string    `json:"user_id" mapstructure:"user_id,omitempty"`
	Title       string    `json:"title" mapstructure:"title,omitempty"`
	Description string    `json:"description,omitempty" mapstructure:"description,omitempty"`
	Status      string    `json:"status" mapstructure:"status,omitempty"`
	CreatedAt   time.Time `json:"created_at" mapstructure:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at" mapstructure:"updated_at,omitempty"`
}
