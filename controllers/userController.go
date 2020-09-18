package controllers

import (
	"errors"

	"github.com/jeevanantham123/golang-tmdb-api/model"
	"github.com/jeevanantham123/golang-tmdb-api/repo"
	"github.com/jinzhu/gorm"
)

//Signup func
func Signup(db *gorm.DB, user *model.User) (string, error) {

	if len(user.Username) > 0 && len(user.Password) > 0 {
		success, err := repo.Signup(db, user)
		return success, err
	}
	return "", errors.New("null value")
}

//Login func
func Login(db *gorm.DB, user *model.User) (string, error) {
	if len(user.Username) > 0 && len(user.Password) > 0 {
		success, err := repo.Login(db, user)
		return success, err
	}
	return "", errors.New("null value")
}
