package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/showbaba/movies-api/data"
	"github.com/showbaba/movies-api/utils"
)

func Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	response := utils.APIResponse{
		Status:  http.StatusOK,
		Message: `yes Lord, I'm alive and listening`,
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		utils.Dispatch500Error(w, err)
		return
	}
	w.Write(responseJSON)
}

func AddComment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	movieID := vars["movie_id"]

	var input CreateCommentPayload
	if body, err := io.ReadAll(r.Body); err != nil {
		utils.Dispatch400Error(w, "invalid request payload", err)
		return
	} else if err := json.Unmarshal(body, &input); err != nil {
		utils.Dispatch400Error(w, "invalid request payload", nil)
		return
	}
	validate := validator.New()
	err := validate.Struct(input)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		utils.Dispatch400Error(w, "validation error", validationErrors)
		return
	}

	// check if movie with id exist in the redis db
	var movie *Movie
	movie, err = getMovieByIDFromRedis(redisClient, movieID)
	if err != nil {
		utils.Dispatch500Error(w, err)
		return
	}
	if movie == nil {
		// check the movies api
		movie, err = getMovieByIDFromAPI(movieID)
		if err != nil {
			utils.Dispatch500Error(w, err)
			return
		}
		if movie == nil {
			utils.Dispatch404Error(w, "movie with id %s not found", err)
			return
		} else {
			// cache movie
			if err := cacheMovie(movieID, movie, redisClient); err != nil {
				utils.Dispatch500Error(w, err)
				return
			}
		}
	}

	// create comment with movie id
	comment := data.Comment{
		MovieID:      movieID,
		Body:         input.Body,
		UserPublicIP: input.UserPublicIP,
	}
	id, err := comment.Insert()
	if err != nil {
		utils.Dispatch500Error(w, err)
		return
	}
	response := utils.APIResponse{
		Status:  http.StatusOK,
		Message: "comment added successfully",
		Data:    map[string]string{"id": fmt.Sprint(id)},
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		utils.Dispatch500Error(w, err)
		return
	}
	w.Write(responseJSON)
}

func FetchMovies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	resp, err := http.Get("https://swapi.dev/api/films/")
	if err != nil {
		utils.Dispatch500Error(w, err)
	}
	defer resp.Body.Close()

	var movieData struct {
		Results []Movie `json:"results"`
	}
	err = json.NewDecoder(resp.Body).Decode(&movieData)
	if err != nil {
		utils.Dispatch500Error(w, err)
		return
	}

	movies := movieData.Results

	// Sort movies by release date
	sort.Slice(movies, func(i, j int) bool {
		dateI, _ := time.Parse("2006-01-02", movies[i].ReleaseDate)
		dateJ, _ := time.Parse("2006-01-02", movies[j].ReleaseDate)
		return dateI.Before(dateJ)
	})

	var cachedMovies []Movie

	for _, movie := range movies {
		key := "movie_title:" + movie.Title
		exists, err := redisClient.Exists(key).Result()
		if err != nil {
			utils.Dispatch500Error(w, err)
			return
		}
		if exists == 1 {
			movieID, err := redisClient.Get(key).Result()
			if err != nil {
				if err == redis.Nil {
					utils.Dispatch404Error(w, "movie with title %s not found", strings.Trim(key, "movie_title:"))
					return
				}
				utils.Dispatch500Error(w, err)
				return
			}
			movie, err := getMovieByIDFromRedis(redisClient, movieID)
			if err != nil {
				utils.Dispatch500Error(w, err)
				return
			}
			// Fetch comments for the movie from PostgreSQL
			comments, err := models.Comment.Fetch(movieID)
			if err != nil {
				utils.Dispatch500Error(w, err)
				return
			}
			movie.Comments = comments
			movie.CommentCount = len(comments)
			cachedMovies = append(cachedMovies, *movie)
			continue
		}

		id, err := redisClient.Incr("movie_id_counter").Result()
		if err != nil {
			utils.Dispatch500Error(w, err)
			return
		}
		movie.ID = id

		jsonData, err := json.Marshal(movie)
		if err != nil {
			utils.Dispatch500Error(w, err)
			return
		}

		err = redisClient.Set(strconv.Itoa(int(id)), jsonData, 0).Err()
		if err != nil {
			utils.Dispatch500Error(w, err)
			return
		}

		// store the movie ID under the key "movie_title:{title}", this helps for faster lookup by title
		err = redisClient.Set(key, id, 0).Err()
		if err != nil {
			utils.Dispatch500Error(w, err)
			return
		}

		// Fetch comments for the movie from PostgreSQL
		comments, err := models.Comment.Fetch(strconv.FormatInt(id, 10))
		if err != nil {
			utils.Dispatch500Error(w, err)
			return
		}

		// Join comments to the movie object in Redis
		movie.Comments = comments
		movie.CommentCount = len(comments)
		cachedMovies = append(cachedMovies, movie)
	}

	response := utils.APIResponse{
		Status:  http.StatusOK,
		Message: "fetch movies successfully",
		Data:    cachedMovies,
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		utils.Dispatch500Error(w, err)
		return
	}
	w.Write(responseJSON)
}

func FetchMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	movieID := vars["movie_id"]

	var err error

	var movie *Movie
	movie, err = getMovieByIDFromRedis(redisClient, movieID)
	if err != nil {
		utils.Dispatch500Error(w, err)
		return
	}
	if movie == nil {
		// check the movies api
		movie, err = getMovieByIDFromAPI(movieID)
		if err != nil {
			utils.Dispatch500Error(w, err)
			return
		}
		if movie == nil {
			utils.Dispatch404Error(w, "movie with id %s not found", err)
			return
		} else {
			// cache movie
			if err := cacheMovie(movieID, movie, redisClient); err != nil {
				utils.Dispatch500Error(w, err)
				return
			}
		}
	}
	// Fetch comments for the movie from PostgreSQL
	comments, err := models.Comment.Fetch(movieID)
	if err != nil {
		utils.Dispatch500Error(w, err)
		return
	}
	movie.Comments = comments
	movie.CommentCount = len(comments)
	response := utils.APIResponse{
		Status:  http.StatusOK,
		Message: "fetch movie successfully",
		Data:    movie,
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		utils.Dispatch500Error(w, err)
		return
	}
	w.Write(responseJSON)
}

func FetchMovieCharacters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	movieID := vars["movie_id"]
	queryParams := r.URL.Query()
	sortBy := queryParams.Get("sort_by")
	sortOrder := queryParams.Get("sort_order")

	var movie *Movie
	var err error
	movie, err = getMovieByIDFromRedis(redisClient, movieID)
	if err != nil {
		utils.Dispatch500Error(w, err)
		return
	}
	if movie == nil {
		movie, err = getMovieByIDFromAPI(movieID)
		if err != nil {
			utils.Dispatch500Error(w, err)
			return
		}
		if movie == nil {
			utils.Dispatch404Error(w, "movie with id %s not found", err)
			return
		} else {
			if err := cacheMovie(movieID, movie, redisClient); err != nil {
				utils.Dispatch500Error(w, err)
				return
			}
		}
	}

	characters := make([]Character, 0)
	for _, characterURL := range movie.Characters {
		var character *Character
		// fetch from redis first
		key := fmt.Sprintf("movie_character:%s:%s", movieID, characterURL)
		val, err := redisClient.Get(key).Result()
		if err != nil {
			if err == redis.Nil {
				resp, err := http.Get(characterURL)
				if err != nil {
					utils.Dispatch500Error(w, err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					err := fmt.Errorf("received non-OK status code %d from API", resp.StatusCode)
					utils.Dispatch500Error(w, err)
					return
				}

				err = json.NewDecoder(resp.Body).Decode(&character)
				if err != nil {
					utils.Dispatch500Error(w, err)
					return
				}

				if character == nil {
					utils.Dispatch404Error(w, "character with id %s not found", err)
					return
				}

				// Cache character data in Redis
				characterJSON, err := json.Marshal(character)
				if err != nil {
					utils.Dispatch500Error(w, err)
					return
				}
				if err := redisClient.Set(key, string(characterJSON), 0).Err(); err != nil {
					utils.Dispatch500Error(w, err)
					return
				}

			} else {
				utils.Dispatch500Error(w, err)
				return
			}
		} else {
			if err := json.Unmarshal([]byte(val), &character); err != nil {
				utils.Dispatch500Error(w, err)
				return
			}
		}

		// Sort characters based on the specified field and order
		sort.Slice(characters, func(i, j int) bool {
			switch sortBy {
			case "name":
				if sortOrder == "asc" {
					return characters[i].Name < characters[j].Name
				} else {
					return characters[i].Name > characters[j].Name
				}
			case "gender":
				if sortOrder == "asc" {
					return characters[i].Gender < characters[j].Gender
				} else {
					return characters[i].Gender > characters[j].Gender
				}
			case "height":
				h1, err1 := strconv.ParseFloat(characters[i].Height, 64)
				h2, err2 := strconv.ParseFloat(characters[j].Height, 64)
				if err1 != nil || err2 != nil {
					// If there was an error parsing the height value, treat the characters as equal
					return false
				}
				if sortOrder == "asc" {
					return h1 < h2
				} else {
					return h1 > h2
				}
			default:
				// If an invalid sort field was provided, treat the characters as equal
				return false
			}
		})

		if character.Height != "unknown" {
			height, err := strconv.ParseFloat(character.Height, 64)
			if err != nil {
				utils.Dispatch500Error(w, err)
				return
			}
			character.Height = utils.CmToFeetInches(height)
			characters = append(characters, *character)
		}
	}

	response := utils.APIResponse{
		Status:  http.StatusOK,
		Message: "fetch movie character successfully",
		Data:    characters,
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		utils.Dispatch500Error(w, err)
		return
	}
	w.Write(responseJSON)
}

func getMovieByIDFromRedis(client *redis.Client, movieID string) (*Movie, error) {
	movieJSON, err := client.Get(movieID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var movie Movie
	err = json.Unmarshal([]byte(movieJSON), &movie)
	if err != nil {
		return nil, err
	}

	return &movie, nil
}

func getMovieByIDFromAPI(movieID string) (*Movie, error) {
	url := fmt.Sprintf("https://swapi.dev/api/films/%s/", movieID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Title        string   `json:"title"`
		OpeningCrawl string   `json:"opening_crawl"`
		Characters   []string `json:"characters"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	movie := &Movie{
		Title:        data.Title,
		OpeningCrawl: data.OpeningCrawl,
		Characters:   data.Characters,
	}

	return movie, nil
}

func cacheMovie(movieID string, movie *Movie, client *redis.Client) error {
	movieJSON, err := json.Marshal(movie)
	if err != nil {
		return err
	}

	err = client.Set(movieID, string(movieJSON), 0).Err()
	if err != nil {
		return err
	}

	return nil
}
