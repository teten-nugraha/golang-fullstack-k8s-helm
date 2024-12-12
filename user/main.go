package main

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"sync"
)

// User represents a user object
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// InMemoryStore holds user data in memory
type InMemoryStore struct {
	data map[string]User
	mu   sync.RWMutex
}

// NewInMemoryStore initializes the in-memory store
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{data: make(map[string]User)}
}

// AddUser adds a user to the store
func (store *InMemoryStore) AddUser(user User) error {
	store.mu.RLock()
	defer store.mu.RUnlock()

	if _, exists := store.data[user.Email]; exists {
		return errors.New("user already exists")
	}

	store.data[user.Email] = user
	return nil
}

// GetUser retrieves a user by email
func (store *InMemoryStore) GetUser(email string) (User, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	user, exists := store.data[email]
	if !exists {
		return User{}, errors.New("user not found")
	}

	return user, nil
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Initialize in-memory store
	store := NewInMemoryStore()

	r.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if user.Name == "" || user.Email == "" {
			http.Error(w, "name and email are required", http.StatusBadRequest)
			return
		}

		if err := store.AddUser(user); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
	r.Get("/users/{email}", func(w http.ResponseWriter, r *http.Request) {
		email := chi.URLParam(r, "email")
		user, err := store.GetUser(email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	})

	http.ListenAndServe(":8081", r)
}
