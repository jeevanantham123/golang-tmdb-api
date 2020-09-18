package model

import "github.com/jinzhu/gorm"

//User structure
type User struct {
	gorm.Model
	Username string `json:"username" query:"username" gorm:"unique" gorm:"NOT NULL"`
	Password string `json:"password" query:"password" gorm:"NOT NULL"`
}
