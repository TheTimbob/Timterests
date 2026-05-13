package server

import "net/http"

func (s *Server) MaxBytesMiddleware(next http.Handler) http.Handler {
	return s.maxBytesMiddleware(next)
}
