package main

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/collector/api"
	"github.com/test-network-function/collector/storage"
	"github.com/test-network-function/collector/util"
)

func main() {
	readTimeOut, writeTimeOut, addr, envErr := util.GetServerEnvVars()
	if envErr != "" {
		logrus.Errorf(util.ServerEnvVarsError, envErr)
	}

	s3Store := storage.NewS3Storage()
	mysqlStore := storage.NewMySQLStorage()
	defer mysqlStore.MySQL.Close()

	server := api.NewServer(addr, mysqlStore, s3Store,
		time.Duration(readTimeOut)*time.Second, time.Duration(writeTimeOut)*time.Second)
	logrus.Fatal(server.Start())
}
