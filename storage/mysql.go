package storage

import (
	"database/sql"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/collector/util"
)

type MySQLStorage struct{ MySql *sql.DB }

func (s *MySQLStorage) Get() *MySQLStorage {
	return NewMySQLStorage()
}

// constructor
func NewMySQLStorage() *MySQLStorage {

	logrus.Info("Retrieving database information")
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
	return &MySQLStorage{MySql: db}
}
