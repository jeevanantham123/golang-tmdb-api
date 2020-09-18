package repo

import (
	"fmt"

	"github.com/jeevanantham123/golang-tmdb-api/model"
	"github.com/jinzhu/gorm"
)

//Signup func
func Signup(db *gorm.DB, user *model.User) (string, error) {
	err := db.Create(&user)
	if err.Error != nil {
		return "", err.Error
	}
	return "success", nil
}

//Login func
func Login(db *gorm.DB, user *model.User) (string, error) {
	fmt.Println(user.Username)
	var users model.User
	err := db.Where("username = ? AND password = ?", user.Username, user.Password).First(&users)
	if err.Error != nil {
		return "", err.Error
	}
	return "success", nil
}
