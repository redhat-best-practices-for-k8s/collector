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
	http.HandleFunc("/healthz", s.healthHandler)
	http.HandleFunc("/", s.handler)
	return s.server.ListenAndServe()
}

func (s *Server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	if err := s.database.MySQL.Ping(); err != nil {
		util.WriteMsg(w, http.StatusServiceUnavailable, "database unreachable")
		return
	}
	util.WriteMsg(w, http.StatusOK, "ok")
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	logrus.Info(util.ServerIsUpMsg)

	switch r.Method {
	case http.MethodGet:
		ResultsHandler(w, r, s.database)
	case http.MethodPost:
		ParserHandler(w, r, s.database)
	default:
		util.WriteMsg(w, http.StatusMethodNotAllowed, util.InvalidRequestErr)
		logrus.Errorf(util.InvalidRequestErr)
	}
}
