package database

import (
	"fmt"
	"log"
	"os"

	"github.com/iamsaidovibra/blog-rest-api/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DbInstance struct {
	Db *gorm.DB
}

var Database DbInstance

func ConnectDb() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("SSL_MODE"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the DB \n", err.Error())
		// os.Exit(2)
	}
	log.Println("Connected to the DB")
	db.Logger = logger.Default.LogMode(logger.Info)
	log.Println("Running migrations")
	//TODO: Add migrations

	db.AutoMigrate(&models.Article{}, &models.Comment{}, &models.Like{}, &models.User{})
	Database = DbInstance{Db: db}
}
