package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	// Import your storage package
	"github.com/cva70/nasa-asteroid-tracker/simulation-engine/pkg/storage"
	_ "github.com/lib/pq"
)

func main() {

	godotenv.Load()

	dbConnString := os.Getenv("DB_CONNECTION_STRING")
	db := storage.InitializeDB(dbConnString)

	r := gin.Default()

	r.GET("/asteroids/dangerous", func(c *gin.Context) {
		results, err := storage.GetDangerousAsteroids(db)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"count": len(results),
			"data":  results,
		})
	})
	log.Println("API Server running on http://localhost:8080")
	r.Run(":8080")
}
