package dao

import (
	"errors"
	"github.com/sirupsen/logrus"
	"gochat_my/db"
	"time"
)

var dbInts = db.GetDB("gochat")

type User struct {
	Id         int `gorm:"primary_key"`
	UserName   string
	Password   string
	CreateTime time.Time
	db.DbGoChat
}

func (*User) TableName() string {
	return "user"
}
func (u *User) Add() (userId int, err error) {
	if u.UserName == "" || u.Password == "" {
		logrus.Infof("name and pwd is %s,%s", u.UserName, u.Password)
		return 0, errors.New("username or password empty !")
	}
	oUser := u.CheckHaveUserName(u.UserName)
	if oUser.Id > 0 {
		return oUser.Id, nil
	}
	u.CreateTime = time.Now()
	if err = dbInts.Table(u.TableName()).Create(&u).Error; err != nil {
		return 0, err
	}
	return u.Id, nil
}
func (u *User) CheckHaveUserName(userName string) (data User) {
	dbInts.Table(u.TableName()).Where("user_name=?", userName).Take(&data)
	return
}
func (u *User) GetUserNameByUserId(userId int) (userName string) {
	var data User
	dbInts.Table(u.TableName()).Where("id=?", userId).Take(&data)
	return data.UserName
}
