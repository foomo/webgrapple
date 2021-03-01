package server

import "net/http"

type server struct {
}

func NewServer(port string) (http.Handler, error) {
	return &server{}, nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
