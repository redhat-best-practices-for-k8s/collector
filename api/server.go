package api

import (
	"net/http"
	"time"

	"github.com/redhat-best-practices-for-k8s/collector/storage"
	"github.com/redhat-best-practices-for-k8s/collector/util"
	"github.com/sirupsen/logrus"
)

type Server struct {
	database *storage.MySQLStorage
	server   *http.Server
}

func NewServer(listenAddr string, db *storage.MySQLStorage, rTimeout, wTimeout time.Duration) *Server {
	return &Server{
		database: db,
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
	logrus.Info(util.ServerIsUpMsg)

	switch r.Method {
	case http.MethodGet:
		ResultsHandler(w, r, s.database)
	case http.MethodPost:
		ParserHandler(w, r, s.database)
	default:
		util.WriteMsg(w, util.InvalidRequestErr)
		logrus.Errorf(util.InvalidRequestErr)
	}
}
