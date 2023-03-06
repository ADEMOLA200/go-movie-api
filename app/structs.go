package app

import (
	"github.com/gorilla/mux"
	"github.com/showbaba/movies-api/data"
)

type App struct {
	Router *mux.Router
}

type CreateCommentPayload struct {
	Body         string `json:"body" validate:"required"`
	UserPublicIP string `json:"user_public_ip" validate:"required"`
}

type Movie struct {
	ID           interface{}
	Title        string   `json:"title"`
	OpeningCrawl string   `json:"opening_crawl"`
	Comments     []*data.Comment `json:"comments"`
	CommentCount int `json:"comments_count"`
	ReleaseDate  string `json:"release_date"`
	Characters []string `json:"characters"`
}

type Character struct {
	Name   string `json:"name"`
	Height string `json:"height"`
	Gender string `json:"gender"`
}

type MovieTitleWithID struct {
    Title string `json:"title"`
    ID    string `json:"id"`
}