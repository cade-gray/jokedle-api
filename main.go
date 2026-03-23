package main

import (
	"fmt"
	"jokedle-api/routes"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	router := gin.Default()
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	jokedle_db_name := os.Getenv("JOKEDLE_DB_NAME")
	auth_db_name := os.Getenv("AUTH_DB_NAME")
	port := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")

	jokedle_dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, jokedle_db_name, port, sslmode,
	)
	auth_dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, auth_db_name, port, sslmode,
	)

	jokedle_db, err := gorm.Open(postgres.Open(jokedle_dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect jokedle database")
	}

	auth_db, err := gorm.Open(postgres.Open(auth_dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect auth database")
	}

	routes.RegisterJokeRoutes(router, jokedle_db, auth_db)

	router.Run(":4269")
}
