package models

import "time"

type Comment struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	ComponentID int       `json:"component_id"`
	ParentID    int       `json:"parent_id"`
	Body        string    `json:"body"`
	Created     time.Time `json:"created"`
}
