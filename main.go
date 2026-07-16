package main

import "fmt"
import "net/http"
import "encoding/json"
import "math/rand"
import "time"

type URL struct {
	Original string `json:"original"`
	ShortCode string `json:"short_code"`
}

// In-memory store
var urlStore = make(map[string]string)

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
	urlStore[code] = input.Original

	response := URL{Original: input.Original, ShortCode: code}
	response_writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response_writer).Encode(response)
}

// GET: /
func redirectHandler(response_writer http.ResponseWriter, request *http.Request) {
	code := request.URL.Path[1:] // Remove leading /
	original, exists := urlStore[code]
	fmt.Println("original", original, "exists", exists)
	if !exists {
		http.Error(response_writer, "Not found", http.StatusNotFound)
		return
	}
	http.Redirect(response_writer, request, original, http.StatusFound)
}

func main() {
	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/", redirectHandler)
	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}