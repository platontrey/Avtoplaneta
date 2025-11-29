package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var store *sessions.CookieStore

func initDB() {
	var err error

	// Строка подключения к PostgreSQL
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Локальное подключение к PostgreSQL по умолчанию
		dsn = "host=localhost user=postgres password=qewret123 dbname=autoplanet port=5432 sslmode=disable"
	}

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}

	// Автоматическая миграция схем заказов и запчастей
	if err := db.AutoMigrate(&Order{}, &OrderItem{}, &Part{}); err != nil {
		log.Fatal("Не удалось выполнить миграцию базы данных:", err)
	}
}

func initAuth() {
	// Инициализация хранилища сессий с безопасным случайным ключом
	sessionKey := os.Getenv("SESSION_SECRET")
	if sessionKey == "" {
		// Генерация безопасного случайного ключа, если не предоставлен (минимум 32 байта)
		log.Println("ПРЕДУПРЕЖДЕНИЕ: SESSION_SECRET не установлен, используем ключ по умолчанию. УСТАНОВИТЕ ЭТО В ПРОДАКШЕНЕ!")
		sessionKey = "CHANGE_THIS_IN_PRODUCTION_TO_A_SECURE_RANDOM_KEY_32_CHARS_MIN"
	}

	// Проверка длины ключа сессии
	if len(sessionKey) < 32 {
		log.Fatal("БЕЗОПАСНОСТЬ: SESSION_SECRET должен быть не менее 32 символов")
	}

	store = sessions.NewCookieStore([]byte(sessionKey))

	// Параметры безопасного использования cookies
	isProduction := os.Getenv("NODE_ENV") == "production"
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,            // 7 дней
		HttpOnly: true,                 // Предотвращает XSS атаки
		Secure:   isProduction,         // HTTPS только в продакшене
		SameSite: http.SameSiteLaxMode, // Защита от CSRF
	}
}

func main() {
	initDB()
	initAuth()

	r := mux.NewRouter()

	// CORS middleware для кросс-доменных запросов
	r.Use(func(next http.Handler) http.Handler {
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
	})

	// Применение middleware аутентификации ко всем маршрутам
	r.Use(authMiddleware)

	// Маршруты только для администраторов
	admin := r.PathPrefix("/admin").Subrouter()
	admin.Use(adminMiddleware)
	admin.HandleFunc("/orders", getOrdersHandler).Methods("GET", "OPTIONS")
	admin.HandleFunc("/orders/{id:[0-9]+}/status", updateOrderStatusHandler).Methods("PUT", "OPTIONS")
	admin.HandleFunc("/orders/{id:[0-9]+}/complete", completeOrderHandler).Methods("PUT", "OPTIONS")
	admin.HandleFunc("/orders/{id:[0-9]+}", deleteOrderHandler).Methods("DELETE", "OPTIONS")

	// Обычные пользовательские маршруты (для создания заказов)
	r.HandleFunc("/orders", createOrderHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/orders", getOrdersHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/orders/{id:[0-9]+}/items", addOrderItemHandler).Methods("POST", "OPTIONS")

	log.Println("Сервис заказов запущен на порту 8082")
	log.Fatal(http.ListenAndServe(":8082", r))
}
