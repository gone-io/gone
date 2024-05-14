package domain

type User struct {
	Id    int64  `json:"id,omitempty"`
	Name  string `json:"name,omitempty" binding:"required"`
	Email string `json:"email,omitempty"`
}
