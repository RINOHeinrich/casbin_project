/*
	package main

import (

	"github.com/casbin/casbin"
	"github.com/labstack/echo"

)

	type Enforcer struct {
		enforcer *casbin.Enforcer
	}

	func (e *Enforcer) Enforce(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, _, _ := c.Request().BasicAuth()
			method := c.Request().Method
			path := c.Request().URL.Path

			result, _ := e.enforcer.EnforceSafe(user, path, method)

			if result {
				return next(c)
			}
			return echo.ErrForbidden
		}
	}

	/* func main() {
		e := echo.New()
		enforcer := Enforcer{enforcer: casbin.NewEnforcer("model.conf", "policy.csv")}
		e.Use(enforcer.Enforce)
		e.GET("/project", func(c echo.Context) error {
			return c.JSON(http.StatusOK, "project get allowed")
		})
		e.POST("/project", func(c echo.Context) error {
			return c.JSON(http.StatusOK, "project post allowed")
		})

		e.GET("/channel", func(c echo.Context) error {
			return c.JSON(http.StatusOK, "channel get allowed")
		})

		e.POST("/channel", func(c echo.Context) error {
			return c.JSON(http.StatusOK, "channel post allowed")
		})
		e.Logger.Fatal(e.Start("0.0.0.0:3000"))
	}
*/
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/go-redis/redis/v8" // Import the Redis package
	"github.com/labstack/echo/v4"
)

// Define the Redis client
var RedisCache *redis.Client

func Authenticate(adapter *gormadapter.Adapter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) (err error) {

			ctx := e.Request().Context()

			user, _, _ := e.Request().BasicAuth()
			method := e.Request().Method
			path := e.Request().URL.Path

			key := fmt.Sprintf("%s-%s-%s", user, path, method)

			result := RedisCache.Get(ctx, key)
			val, err := result.Result()
			if err == nil {
				boolValue, err := strconv.ParseBool(val)
				if err != nil {
					log.Fatal(err)
				}

				if !boolValue {
					return &echo.HTTPError{
						Code:    http.StatusForbidden,
						Message: "not allowed",
					}
				}
				return next(e)
			}

			// Casbin enforces policy
			ok, err := enforce(ctx, user, path, method, adapter)
			if err != nil || !ok {

				return &echo.HTTPError{
					Code:    http.StatusForbidden,
					Message: "not allowed",
				}
			}
			if !ok {
				return err
			}
			return next(e)
		}
	}
}

func enforce(ctx context.Context, sub string, obj string, act string, adapter *gormadapter.Adapter) (bool, error) {
	// Load model configuration file and policy store adapter
	enforcer, err := casbin.NewEnforcer("./model.conf", adapter)
	if err != nil {
		return false, fmt.Errorf("failed to load policy from DB: %w", err)
	}
	// Load policies from DB dynamically
	err = enforcer.LoadPolicy()
	if err != nil {
		return false, fmt.Errorf("error in policy: %w", err)
	}
	// Verify
	ok, err := enforcer.Enforce(sub, obj, act)
	if err != nil {
		return false, fmt.Errorf("error in policy: %w", err)
	}
	key := fmt.Sprintf("%s-%s-%s", sub, obj, act)
	RedisCache.Set(ctx, key, strconv.FormatBool(ok), time.Hour)
	return ok, nil
}
