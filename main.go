package main

import (
	"log"
	"net/http"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize the Redis client
	RedisCache = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // Add your Redis password here if required
		DB:       0,
	})

	// Initialize the PostgreSQL database connection
	dsn := "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the Casbin adapter
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the Echo web framework
	e := echo.New()

	// Use the Authenticate middleware
	e.Use(Authenticate(adapter))

	// Define your routes and handlers here
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
