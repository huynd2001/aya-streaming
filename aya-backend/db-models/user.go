package models

import "gorm.io/gorm"

type GORMUser struct {
	gorm.Model
	ID       uint   `json:"id"`
	Username string `gorm:"unique" json:"userName"`
	Email    string `json:"email"`
}