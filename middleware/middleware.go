package middleware

import (
	"fmt"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

//AccessDetails struct
type AccessDetails struct {
	AccessUUID string
	Username   string
}

//ExtractToken struct
func ExtractToken(c *fiber.Ctx) string {
	bearToken := c.Get("Authorization")
	//normally Authorization the_token_xxx
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

//VerifyToken struct
func VerifyToken(c *fiber.Ctx) (*jwt.Token, error) {
	tokenString := ExtractToken(c)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("ACCESS_SECRET"), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

//TokenValid func
func TokenValid(c *fiber.Ctx) error {
	token, err := VerifyToken(c)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return err
	}
	return nil
}

//ExtractTokenMetadata func
func ExtractTokenMetadata(c *fiber.Ctx) (*AccessDetails, error) {
	token, err := VerifyToken(c)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUUID, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}
		username, er := claims["user_name"].(string)
		if !er {
			return nil, err
		}
		return &AccessDetails{
			AccessUUID: accessUUID,
			Username:   username,
		}, nil
	}
	return nil, err
}

//FetchAuth func
func FetchAuth(c *fiber.Ctx, authD *AccessDetails, client *redis.Client) (string, error) {
	ctx := c.Context()
	username, err := client.Get(ctx, authD.AccessUUID).Result()
	if err != nil {
		return "", err
	}
	return username, nil
}

//DeleteAuth func
func DeleteAuth(c *fiber.Ctx, givenUUID string, client *redis.Client) (int64, error) {
	ctx := c.Context()
	deleted, err := client.Del(ctx, givenUUID).Result()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}
