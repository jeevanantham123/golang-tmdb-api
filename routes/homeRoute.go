package routes

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/jeevanantham123/golang-tmdb-api/middleware"
	"github.com/jinzhu/gorm"
)

//HomeRoutes func
func HomeRoutes(db *gorm.DB, app *fiber.App, client *redis.Client) {
	app.Get("/home", authRequired(), func(c *fiber.Ctx) error {
		tokenAuth, err := middleware.ExtractTokenMetadata(c)
		if err != nil {
			return c.JSON(fiber.Map{
				"error": err,
			})
		}
		username, err := middleware.FetchAuth(c, tokenAuth, client)
		if err != nil {
			return c.JSON(fiber.Map{
				"error": err,
			})
		}
		s, err := http.Get("https://api.themoviedb.org/3/search/movie?api_key=d272326e467344029e68e3c4ff0b4059&language=en-US&query=spiderman")
		if err != nil {
			return c.SendString("error")
		}

		res, err := ioutil.ReadAll(s.Body)
		if err != nil {
			log.Fatal(err)
		}

		return c.JSON(fiber.Map{
			"username": username,
			"value":    string(res),
		})
	})
}

func authRequired() func(ctx *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		err := middleware.TokenValid(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": err,
			})
		}
		return c.Next()
	}
}
