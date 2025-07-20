package request

type TaskCreateRequest struct {
	Title       string  `json:"title" validate:"required"`
	Description *string `json:"description,omitempty"`
}

type TaskUpdateRequest struct {
	Status string `json:"status" validate:"required,oneof=new in_progress done"`
}
