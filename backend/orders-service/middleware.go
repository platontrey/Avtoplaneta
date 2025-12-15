package main

import (
	"log"
	"net/http"
)

// CORSMiddleware добавляет CORS заголовки для кросс-доменных запросов
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("CORS: Получен %s запрос к %s от %s", req.Method, req.URL.Path, req.RemoteAddr)
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if req.Method == "OPTIONS" {
			log.Printf("CORS: Обработка предварительного OPTIONS запроса к %s", req.URL.Path)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, req)
	})
}