package api

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/collector/storage"
	"github.com/test-network-function/collector/util"
)

type Server struct {
	database    *storage.MySQLStorage
	objectStore *storage.S3Storage
	server      *http.Server
}

func NewServer(listenAddr string, db *storage.MySQLStorage, objectStore *storage.S3Storage, rTimeout, wTimeout time.Duration) *Server {
	return &Server{
		database:    db,
		objectStore: objectStore,
		server: &http.Server{
			Addr:         listenAddr,
			ReadTimeout:  rTimeout,
			WriteTimeout: wTimeout,
		},
	}
}

func (s *Server) Start() error {
	logrus.Info("Starting server")
	http.HandleFunc("/", s.handler)
	return s.server.ListenAndServe()
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	printServerUpMessage(w)

	switch r.Method {
	case http.MethodGet:
		ResultsHandler(w, r, s.database)
	case http.MethodPost:
		ParserHandler(w, r, s.database)
	default:
		util.WriteError(w, util.InvalidRequestErr)
		logrus.Errorf(util.InvalidRequestErr)
	}
}

func printServerUpMessage(w http.ResponseWriter) {
	logrus.Info(util.ServerIsUpMsg)
	_, writeErr := w.Write([]byte(util.ServerIsUpMsg + "\n"))
	if writeErr != nil {
		logrus.Errorf(util.WritingResponseErr, writeErr)
	}
}
