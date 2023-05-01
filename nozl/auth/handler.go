package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats-server/v2/nozl/eventstream"
	"github.com/nats-io/nats-server/v2/nozl/shared"
	"github.com/nats-io/nats-server/v2/nozl/user"
	"golang.org/x/crypto/bcrypt"
)

type (
	JwtCustomClaims struct {
		Name string `json:"name"`
		jwt.RegisteredClaims
	}
)

func GetConfig() echojwt.Config {
	return echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(JwtCustomClaims)
		},
		SigningKey: []byte(shared.JWTSecret),
	}
}

func GetKeyAuthConfig(apiKey string, ctx echo.Context) (bool, error) {
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(shared.TenantKV)
	if err != nil {
		return false, ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to retreive Tenant Key Value store",
		})
	}

	allKeys, err := kv.Keys()
	if err != nil {
		return false, ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to retreive keys from the Key Value store",
		})
	}

	for _, key := range allKeys {
		value, err := kv.Get(key)
		if err != nil {
			return false, ctx.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Unable to retreive value corresponding to a key from the Key Value store",
			})
		}

		payload := make(map[string]string)
		if err := json.Unmarshal(value.Value(), &payload); err != nil {
			shared.Logger.Error(err.Error())
		}

		if payload["api_key"] == apiKey {
			return true, nil
		}
	}

	return false, nil
}

func LoginHandler(ctx echo.Context) error {
	newUser := &user.User{}

	if err := ctx.Bind(newUser); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to parse request's body",
		})
	}

	kvStore, err := eventstream.Eventstream.RetreiveKeyValStore(shared.UserKV)

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to retrieve key value store",
		})
	}

	keys, err := kvStore.Keys()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to retrieve keys from key value store",
		})
	}

	for _, key := range keys {
		userKvEntry, _ := kvStore.Get(key)
		userInfo := &user.User{}
		_ = json.Unmarshal(userKvEntry.Value(), userInfo)
		err := bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(newUser.Password))

		if newUser.UserName == userInfo.UserName && err == nil {
			// Set custom claims
			claims := &JwtCustomClaims{
				newUser.UserName,
				jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
				},
			}

			// Create token with claims
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

			// Generate encoded token and send it as response.
			t, err := token.SignedString([]byte(shared.JWTSecret))
			if err != nil {
				return err
			}

			return ctx.JSON(http.StatusOK, echo.Map{
				"token": t,
			})
		}
	}

	return ctx.JSON(http.StatusNotFound, echo.Map{
		"message": "username not found",
	})
}

func RestrictedHandler(ctx echo.Context) error {
	u := ctx.Get("user").(*jwt.Token)
	claims := u.Claims.(*JwtCustomClaims)
	name := claims.Name

	return ctx.JSON(http.StatusOK, echo.Map{
		"message": "Welcome " + name + "!",
	})
}

func SignUpHandler(ctx echo.Context) error {
	newUser := &user.User{}

	if err := ctx.Bind(newUser); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to parse request's body",
		})
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newUser.Password), 8)
	newUser.Password = string((hashedPassword))
	jsonPayload, _ := json.Marshal(newUser)
	kvStore, err := eventstream.Eventstream.RetreiveKeyValStore(shared.UserKV)

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to retrieve key value store",
		})
	}

	_, err = kvStore.Put(uuid.New().String(), jsonPayload)

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unable to save user information",
		})
	}

	claims := &JwtCustomClaims{
		newUser.UserName,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(shared.JWTSecret))

	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"token": t,
	})
}
