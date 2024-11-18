package main

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redhat-best-practices-for-k8s/collector/api"
	"github.com/redhat-best-practices-for-k8s/collector/storage"
	"github.com/redhat-best-practices-for-k8s/collector/util"
	"github.com/sirupsen/logrus"
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
