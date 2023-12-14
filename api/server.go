package api

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/collector/storage"
	"github.com/test-network-function/collector/util"
)

type Server struct {
	listenAddr   string
	database     *storage.MySqlStorage
	objectStore  *storage.S3Storage
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func NewServer(listenAddr string, db *storage.MySqlStorage, objectStore *storage.S3Storage, rTimeout, wTimeout time.Duration) *Server {
	return &Server{
		listenAddr:   listenAddr,
		database:     db,
		objectStore:  objectStore,
		readTimeout:  rTimeout,
		writeTimeout: wTimeout,
	}
}

func (s *Server) Start() error {
	logrus.Info("Starting server")
	http.HandleFunc("/", s.handler)
	return http.ListenAndServe(s.listenAddr, nil)
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	printServerUpMessage(w)

	switch r.Method {
	case http.MethodGet:
		ResultsHandler(w, r, s.database)
	case http.MethodPost:
		ParserHandler(w, r, s.database)
	default:
		_, writeErr := w.Write([]byte(util.InvalidRequestErr + "\n"))
		if writeErr != nil {
			logrus.Errorf(util.WritingResponseErr, writeErr)
		}
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
