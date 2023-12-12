package storage

import (
	"database/sql"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/collector/util"
)

type MySqlStorage struct{ MySql *sql.DB }

func (s *MySqlStorage) Get() *MySqlStorage {
	return NewMySqlStorage()
}

// constructor
func NewMySqlStorage() *MySqlStorage {

	logrus.Info("Retrieving database infomation")
	DBUsername, DBPassword, DBURL, DBPort := util.GetDatabaseEnvVars()

	DBConnStr := DBUsername + ":" + DBPassword + "@tcp(" + DBURL + ":" + DBPort + ")/"

	db, err := sql.Open("mysql", DBConnStr)
	if err != nil {
		return nil
	}
	logrus.Info("Checking connection to database")
	err = db.Ping()
	if err != nil {
		return nil
	}
	logrus.Info("Connection successful")
	return &MySqlStorage{MySql: db}
}
