package models

import "gorm.io/gorm"

type GORMUser struct {
	gorm.Model
	Username string `gorm:"unique"`
	Email    string `gorm:"unique"`
}
