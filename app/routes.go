package app

import (
	"context"
	"log"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/showbaba/movies-api/data"
)

var (
	ctx    = context.Background()
	models *data.Models
	redisClient  *redis.Client
)

func (a *App) Initialize(dbModels *data.Models, redisCLient *redis.Client) {
	a.Router = mux.NewRouter()
	a.setRouters()
	models = dbModels
	redisClient = redisCLient
}

func (a *App) setRouters() {
	a.Get("/ping", Ping)
	a.Post("/movies/{movie_id}/comment", AddComment)
	a.Get("/movies", FetchMovies)
	a.Get("/movies/{movie_id}", FetchMovie)
	a.Get("/movies/{movie_id}/characters", FetchMovieCharacters)
}

// handler method
func (a *App) Post(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("Post")
}

func (a *App) Get(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("Get")
}

func (a *App) Put(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("Put")
}

func (a *App) Delete(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("Delete")
}

// run
func (a *App) Run(host string) {
	// CORS
	log.Fatal(
		http.ListenAndServe(
			host,
			handlers.CORS(
				handlers.AllowCredentials(),
				handlers.AllowedMethods([]string{"POST", "GET", "PUT", "OPTIONS"}),
				handlers.AllowedHeaders([]string{"Authorization", "Content-Type"}),
				handlers.MaxAge(3600),
			)(a.Router),
		),
	)
}
