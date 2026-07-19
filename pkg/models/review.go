package models

import "time"

type Review struct {
	ID          int       `json:"id"`
	UserId      int       `json:"user_id"`
	ComponentId int       `json:"component_id"`
	Rating      int       `json:"rating"`
	Body        string    `json:"body"`
	Likes       int       `json:"likes"`
	Comments    string    `json:"comments"`
	Created     time.Time `json:"created"`
}
