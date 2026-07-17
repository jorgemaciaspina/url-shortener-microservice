package main

import "net/http"
import "encoding/json"
import "math/rand"
import "time"
import "os"
import "strings"

// REDIS
import "context"
import "github.com/redis/go-redis/v9"

// LOGGING
import logger "github.com/sirupsen/logrus"

// MONITORING
import "github.com/prometheus/client_golang/prometheus"
import "github.com/prometheus/client_golang/prometheus/promhttp"

var prometheus_requests = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "http_requests_total",
	Help: "Total number of HTTP requests",
})
func init_prometheus() {
	prometheus.MustRegister(prometheus_requests)
}

type URL struct {
	Original string `json:"original"`
	ShortCode string `json:"short_code"`
}

// In-memory store (Replace with Redis for production)
// var urlStore = make(map[string]string)
// Redis client
var ctx = context.Background()
var redis_db *redis.Client
func init() {
	redis_url := os.Getenv("URL_SHORTENER_REDIS")
	if len(redis_url) == 0 {
		redis_url = "localhost"
	}
	redis_port := os.Getenv("URL_SHORTENER_REDIS_PORT")
	if len(redis_port) == 0 {
		redis_port = "6379"
	}
	
	var string_builder strings.Builder
	string_builder.WriteString(redis_url)
	string_builder.WriteString(":")
	string_builder.WriteString(redis_port)
	redis_db = redis.NewClient(&redis.Options{
		Addr: string_builder.String(),
	})

	init_prometheus()
}

// Generate random short code
func generateCode() string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	code := make([]rune, 6)
	for i:= range code {
		code[i] = letters[rand.Intn(len(letters))]
	}
	return string(code)
}

// Post: shortener
func shortenHandler(response_writer http.ResponseWriter, request *http.Request) {
	prometheus_requests.Inc()
	if request.Method != http.MethodPost {
		http.Error(response_writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input URL
	err := json.NewDecoder(request.Body).Decode(&input)
	if err != nil || input.Original == "" {
		http.Error(response_writer, "Invalid input", http.StatusBadRequest)
		return
	}

	code := generateCode()
	// Replaced by redis
	// urlStore[code] = input.Original
	err = redis_db.Set(ctx, code, input.Original, 0).Err()
	if err != nil {
		http.Error(response_writer, "Database error", http.StatusInternalServerError)
		return
	}

	response := URL{Original: input.Original, ShortCode: code}
	response_writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response_writer).Encode(response)
}

// GET: /
func redirectHandler(response_writer http.ResponseWriter, request *http.Request) {
	prometheus_requests.Inc()
	code := request.URL.Path[1:] // Remove leading /
	// Replaced by redis
	// original, exists := urlStore[code]
	original, err := redis_db.Get(ctx, code).Result()
	// fmt.Println("original", original, "exists", exists)
	// if !exists {
	// 	http.Error(response_writer, "Not found", http.StatusNotFound)
	// 	return
	// }
	if err == redis.Nil {
		http.Error(response_writer, "Not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(response_writer, "Database error", http.StatusInternalServerError)
		return
	}
	http.Redirect(response_writer, request, original, http.StatusFound)
}

func main() {
	logger.SetFormatter(&logger.JSONFormatter{})
	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/", redirectHandler)
	http.Handle("/metrics", promhttp.Handler())
	logger.Info("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}