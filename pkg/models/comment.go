package models

import "time"

type Comment struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	ComponentID int       `json:"component_id"`
	ParentID    *int      `json:"parent_id"` // nullable — корневые комментарии
	Body        string    `json:"body"`
	Created     time.Time `json:"created"`

	// Поля из JOIN с users — не хранятся в таблице comments
	UserName string `json:"user_name"`
	Role     string `json:"role"`
}
