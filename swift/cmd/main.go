package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mbartnicki80/swift/internal/api"
	"github.com/mbartnicki80/swift/internal/parser"
	"github.com/mbartnicki80/swift/internal/store"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	records, err := parser.ParseFromExcel("swift_codes.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	err = store.InsertRowsToDatabase(db, records)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	v1 := router.Group("/v1")
	{
		swift := v1.Group("/swift-codes")
		{
			swift.GET("/:swiftCode", api.GetSwiftCodeHandler(db))
			swift.GET("/country/:countryISO2", api.GetSwiftCodesByCountryHandler(db))
			swift.POST("", api.CreateSwiftCodeHandler(db))
			swift.DELETE("/:swiftCode", api.DeleteSwiftCodeHandler(db))
		}
	}

	err = router.Run(":8080")
	if err != nil {
		return
	}
}
