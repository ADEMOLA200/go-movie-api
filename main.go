package main

import (
	"log"

	"github.com/go-redis/redis"
	"github.com/showbaba/movies-api/app"
	"github.com/showbaba/movies-api/data"
	"github.com/showbaba/movies-api/utils"
)

func main() {
	// open connection to redis
	redisCLient := redis.NewClient(&redis.Options{
		Addr: app.GetConfig().RedisURL,
	})
	defer redisCLient.Close()
	// test redis connection
	_, err := redisCLient.Ping().Result()
	if err != nil {
		panic(err)
	}
	dbConn := utils.ConnectToSQLDB(
		app.GetConfig().DbHost,
		app.GetConfig().DbUser,
		app.GetConfig().DbPassword,
		app.GetConfig().DbName,
		app.GetConfig().DbPort,
	)
	defer dbConn.Close()
	models := data.New(dbConn)
	data.Migrate()
	server := app.App{}
	port := app.GetConfig().Port
	server.Initialize(&models, redisCLient)
	log.Printf("talk to me Lord your server is listening on port %s 🙏 ", port)
	server.Run(port)
}
