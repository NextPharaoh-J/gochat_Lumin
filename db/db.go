package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
	"gochat_my/config"
	"path/filepath"
	"sync"
	"time"
)

var dbMap = map[string]*gorm.DB{}
var syncLock sync.Mutex

func init() {
	initDB("gochat")
}
func initDB(dbName string) {
	var err error
	realPath, _ := filepath.Abs("./")
	configFilePath := realPath + "/db/gochat.sqlite3"
	syncLock.Lock()
	dbMap[dbName], err = gorm.Open("sqlite3", configFilePath)
	dbMap[dbName].DB().SetMaxIdleConns(5)
	dbMap[dbName].DB().SetMaxOpenConns(20)
	dbMap[dbName].DB().SetConnMaxLifetime(8 * time.Second)
	if config.GetGinRunMode() == "dev" {
		dbMap[dbName].LogMode(true)
	}
	syncLock.Unlock()
	if err != nil {
		logrus.Errorf("connect db fail :%s", err.Error())
	}
}
func GetDB(dbName string) *gorm.DB {
	if db, ok := dbMap[dbName]; ok {
		return db
	} else {
		return nil
	}
}

type DbGoChat struct{}

func (*DbGoChat) GetDbName() string {
	return "gochat"
}
