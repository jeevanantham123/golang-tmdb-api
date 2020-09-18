package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jeevanantham123/golang-tmdb-api/db"
	"github.com/jeevanantham123/golang-tmdb-api/routes"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Cannot find .env file")
		return
	}
	db, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	client := redisInit()
	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New())
	routes.SetupRoutes(db, app, client)
	routes.HomeRoutes(db, app, client)
	log.Fatal(app.Listen(":9090"))
}

func redisInit() *redis.Client {
	var cli *redis.Client
	cli = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	ctx := context.Background()
	pong, err := cli.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(pong)
	return cli
}
