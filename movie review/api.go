package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Movie struct {
	ID       int      `json:"id"`
	Title    string   `json:"title"`
	Director string   `json:"director"`
	Reviews  []Review `json:"reviews"`
}

type Review struct {
	Reviewer string `json:"reviewer"`
	Rating   int    `json:"rating"` 
	Comment  string `json:"comment"`
}

var (
	movies   = make(map[int]Movie)
	nextID   = 1
	mu       sync.Mutex
)

func main() {
	http.HandleFunc("/movies", handleMovies)      
	http.HandleFunc("/movies/review", addReview) 

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}


func handleMovies(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getMovies(w)
	case http.MethodPost:
		createMovie(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}


func getMovies(w http.ResponseWriter) {
	mu.Lock()
	defer mu.Unlock()

	movieList := make([]Movie, 0, len(movies))
	for _, movie := range movies {
		movieList = append(movieList, movie)
	}
	writeJSON(w, http.StatusOK, movieList)
}


func createMovie(w http.ResponseWriter, r *http.Request) {
	var movie Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	movie.ID = nextID
	nextID++
	movies[movie.ID] = movie
	writeJSON(w, http.StatusCreated, movie)
}


func addReview(w http.ResponseWriter, r *http.Request) {
	var input struct {
		MovieID int    `json:"movie_id"`
		Review  Review `json:"review"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	movie, exists := movies[input.MovieID]
	if !exists {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	movie.Reviews = append(movie.Reviews, input.Review)
	movies[input.MovieID] = movie
	writeJSON(w, http.StatusOK, movie)
}

	
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
