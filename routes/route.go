package routes

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/jeevanantham123/golang-tmdb-api/controllers"
	"github.com/jeevanantham123/golang-tmdb-api/middleware"
	"github.com/jeevanantham123/golang-tmdb-api/model"
	"github.com/jinzhu/gorm"
	"github.com/twinj/uuid"
)

//SetupRoutes func
func SetupRoutes(db *gorm.DB, app *fiber.App, client *redis.Client) {

	app.Get("/signup", func(c *fiber.Ctx) error {
		user := new(model.User)
		if err := c.QueryParser(user); err != nil {
			return err
		}
		var _, err = controllers.Signup(db, user)
		if err != nil {
			return c.JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return setToken(c, user.Username, client)
	})

	app.Get("/login", func(c *fiber.Ctx) error {
		user := new(model.User)
		if err := c.QueryParser(user); err != nil {
			return err
		}
		var _, err = controllers.Login(db, user)
		if err != nil {
			return c.JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return setToken(c, user.Username, client)
	})

	app.Get("/logout", func(c *fiber.Ctx) error {
		c.ClearCookie("token")
		au, err := middleware.ExtractTokenMetadata(c)
		if err != nil {
			return c.SendString("unAuthorized")
		}
		deleted, delErr := middleware.DeleteAuth(c, au.AccessUUID, client)
		if delErr != nil || deleted == 0 { //if any goes wrong
			return c.SendString("unAuth")
		}
		return c.SendString("Successfully signed out")
	})

	app.Post("/token/refresh", func(c *fiber.Ctx) error {
		return Refresh(c, client)
	})

}

func setToken(c *fiber.Ctx, username string, client *redis.Client) error {
	// Create token
	td := &model.TokenDetails{}
	td.AtExpires = time.Now().Add(time.Second * 30).Unix()
	td.AccessUUID = uuid.NewV4().String()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUUID = uuid.NewV4().String()

	var err error
	//Creating Access Token
	// os.Setenv("ACCESS_SECRET", "jdnfksdmfksd") //this should be in an env file
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUUID
	atClaims["user_name"] = username
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte("ACCESS_SECRET"))
	if err != nil {
		return c.JSON(fiber.Map{
			"error": err,
		})
	}

	//Creating Refresh Token
	//os.Setenv("REFRESH_SECRET", "mcmvmkmsdnfsdmfdsjf") //this should be in an env file
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUUID
	rtClaims["user_name"] = username
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte("REFRESH_SECRET"))
	if err != nil {
		return c.JSON(fiber.Map{
			"error": err,
		})
	}

	// Create cookie
	cookie := new(fiber.Cookie)
	cookie.Name = "token"
	cookie.Value = td.AccessToken
	cookie.Expires = time.Now().Add(time.Minute * 15)

	// Set cookie
	c.Cookie(cookie)
	e := CreateAuth(c, username, td, client)
	if e != nil {
		return c.JSON(fiber.Map{
			"error": e,
		})
	}
	return c.JSON(fiber.Map{
		"success":      "successfull",
		"jwttoken":     td.AccessToken,
		"refreshtoken": td.RefreshToken,
	})
}

//CreateAuth func
func CreateAuth(c *fiber.Ctx, username string, td *model.TokenDetails, client *redis.Client) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()
	ctx := c.Context()
	errAccess := client.Set(ctx, td.AccessUUID, username, at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}
	errRefresh := client.Set(ctx, td.RefreshUUID, username, rt.Sub(now)).Err()
	if errRefresh != nil {
		return errRefresh
	}
	return nil
}

//Refresh func
func Refresh(c *fiber.Ctx, client *redis.Client) error {
	type Rtoken struct {
		RefreshToken string `json:"refresh_token"`
	}
	r := new(Rtoken)
	if err := c.BodyParser(r); err != nil {
		return err
	}

	refreshToken := r.RefreshToken
	//verify the token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("REFRESH_SECRET"), nil
	})
	//if there is an error, the token must have expired
	if err != nil {
		return c.JSON(fiber.Map{
			"Error": err,
		})
	}
	//is token valid?
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return c.JSON(fiber.Map{
			"Error": err,
		})
	}
	//Since token is valid, get the uuid:
	claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims
	if ok && token.Valid {
		refreshUUID, ok := claims["refresh_uuid"].(string) //convert the interface to string
		if !ok {
			return c.JSON(fiber.Map{
				"Error": err,
			})
		}
		username := claims["user_name"].(string)

		//Delete the previous Refresh Token
		deleted, delErr := middleware.DeleteAuth(c, refreshUUID, client)
		if delErr != nil || deleted == 0 { //if any goes wrong
			return c.JSON(fiber.Map{
				"Error": "un auth",
			})
		}
		//Create new pairs of refresh and access tokens
		return setToken(c, username, client)

	}

	return c.JSON(fiber.Map{
		"Error": "tok exp",
	})
}
