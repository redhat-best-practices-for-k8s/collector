package storage

import (
	"database/sql"
	"time"

	"github.com/redhat-best-practices-for-k8s/collector/util"
	"github.com/sirupsen/logrus"
)

const (
	dbMaxOpenConns    = 25
	dbMaxIdleConns    = 5
	dbConnMaxLifetime = 5 * time.Minute
)

type MySQLStorage struct{ MySQL *sql.DB }

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
	db.SetMaxOpenConns(dbMaxOpenConns)
	db.SetMaxIdleConns(dbMaxIdleConns)
	db.SetConnMaxLifetime(dbConnMaxLifetime)

	logrus.Info("Connection successful")
	return &MySQLStorage{MySQL: db}
}
