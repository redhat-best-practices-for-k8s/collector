package storage

import (
	"database/sql"

	"github.com/redhat-best-practices-for-k8s/collector/util"
	"github.com/sirupsen/logrus"
)

type MySQLStorage struct{ MySQL *sql.DB }

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
	return &MySQLStorage{MySQL: db}
}
