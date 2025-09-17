package server

import (
	"errors"
	"net/http"
	"time"
	"todo-api/internal/server/handlers"

	"github.com/gorilla/mux"
)

type Server struct {
	httpHandlers *handlers.Handlers
}

func New(h *handlers.Handlers) *Server {
	return &Server{
		httpHandlers: h,
	}
}

func (s *Server) Start(address string, cfgIdleTimeout time.Duration, cfgTimeout time.Duration) error {
	router := mux.NewRouter()

	router.Path("/tasks").Methods("POST").HandlerFunc(s.httpHandlers.AddNewTask)
	router.Path("/tasks/{id}").Methods("GET").HandlerFunc(s.httpHandlers.GetTaskByID)
	router.Path("/tasks").Methods("PATCH").HandlerFunc(s.httpHandlers.CompleteTask)
	router.Path("/tasks").Methods("GET").HandlerFunc(s.httpHandlers.GetAllTasks)
	router.Path("/tasks/{id}").Methods("DELETE").HandlerFunc(s.httpHandlers.DeleteTask)

	srv := &http.Server{
		Addr:         address,
		Handler:      router,
		IdleTimeout:  cfgIdleTimeout, 
		ReadTimeout:  cfgTimeout,     
		WriteTimeout: cfgTimeout,     
	}

	if err := srv.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
	}

	return nil
}
