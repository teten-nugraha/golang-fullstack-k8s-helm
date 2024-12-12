package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"sync"
)

func init() {
	// Get the environment (dev, staging, prod) from the environment variable
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev" // Default to dev if no ENV is set
	}

	// Set the config file path based on the environment
	configFile := fmt.Sprintf("./config/.env.%s", env)

	// Set the configuration type to 'env' for .env file
	viper.SetConfigFile(configFile)
	viper.SetConfigType("env")

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
}

type Product struct {
	ID       int    `json:"id"`
	Quantity int    `json:"quantity"`
	Name     string `json:"name"`
}

type Booking struct {
	ProductId int    `json:"product_id"`
	Email     string `json:"email"`
}

type InMemoryStore struct {
	products map[int]Product
	bookings map[int]string
	mu       sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		products: make(map[int]Product),
		bookings: make(map[int]string),
	}
}

func (store *InMemoryStore) AddProduct(product Product) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.products[product.ID]; exists {
		return errors.New("product already exists")
	}
	store.products[product.ID] = product
	return nil
}

func (store *InMemoryStore) BookProduct(productId int, email string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	product, exists := store.products[productId]
	if !exists {
		return errors.New("product not found")
	}
	if product.Quantity <= 0 {
		return errors.New("product out of stock")
	}
	if _, booked := store.bookings[productId]; booked {
		return errors.New("product already booked")
	}

	// reduce quantity
	product.Quantity--
	store.products[productId] = product
	store.bookings[productId] = email
	return nil
}

func (store *InMemoryStore) GetBookings(email string) ([]map[string]interface{}, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	bookings := []map[string]interface{}{}
	for productId, bookedEmail := range store.bookings {
		if bookedEmail == email {
			product, exists := store.products[productId]
			if !exists {
				continue
			}
			booking := map[string]interface{}{
				"productId": productId,
				"nama":      product.Name,
			}
			bookings = append(bookings, booking)
		}
	}
	return bookings, nil
}

// UserResponse represents the structure for user response from User Service
type UserResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func GetUserData(email string) (UserResponse, error) {
	// Get URL from environment variables
	url := viper.GetString("USER_SERVICE_URL")
	if url == "" {
		return UserResponse{}, fmt.Errorf("USER_SERVICE_URL is not set in the environment")
	}

	// Construct the full URL
	fullURL := url + email

	// Call the User Service API to get user details
	resp, err := http.Get(fullURL)
	if err != nil {
		return UserResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UserResponse{}, fmt.Errorf("failed to get user data: %s", resp.Status)
	}

	var user UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return UserResponse{}, err
	}
	return user, nil
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Initialize in-memory store
	store := NewInMemoryStore()

	// Routes
	r.Post("/products", func(w http.ResponseWriter, r *http.Request) {
		var product Product
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if product.Name == "" || product.ID <= 0 || product.Quantity <= 0 {
			http.Error(w, "invalid product data", http.StatusBadRequest)
			return
		}

		if err := store.AddProduct(product); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})

	r.Post("/products/book", func(w http.ResponseWriter, r *http.Request) {

		var booking Booking
		if err := json.NewDecoder(r.Body).Decode(&booking); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if booking.Email == "" {
			http.Error(w, "email is required", http.StatusBadRequest)
			return
		}

		if err := store.BookProduct(booking.ProductId, booking.Email); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	r.Get("/users/{email}/bookings", func(w http.ResponseWriter, r *http.Request) {
		email := chi.URLParam(r, "email")

		// Get user data from User Service
		user, err := GetUserData(email)
		if err != nil {
			http.Error(w, "failed to get user data", http.StatusInternalServerError)
			return
		}

		// Get user bookings
		bookings, err := store.GetBookings(email)
		if err != nil {
			http.Error(w, "failed to get bookings", http.StatusInternalServerError)
			return
		}

		// Combine user data and bookings
		response := map[string]interface{}{
			"name":     user.Name,
			"email":    user.Email,
			"age":      user.Age,
			"bookings": bookings,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Start the server
	http.ListenAndServe(":8082", r)
}
