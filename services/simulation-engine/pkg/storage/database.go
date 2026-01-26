package storage

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func InitializeDB(connString string) *sql.DB {
	db, err := sql.Open("postgres", connString)

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Database unreachable: %v", err)
	}

	fmt.Println("Connected to PostgreSQL Database")
	return db
}

func SaveTrajectory(db *sql.DB, asteroidID string, asteroidName string, closestDistKm float64, impactDate float64, isDangerous bool) {

	query := `
        INSERT INTO simulation_results (asteroid_id, name, closest_distance_km, impact_date, is_dangerous)
        VALUES ($1, $2, $3, $4, $5)
    `

	_, err := db.Exec(query, asteroidID, asteroidName, closestDistKm, impactDate, isDangerous)
	if err != nil {
		fmt.Printf("Failed to save to DB: %v\n", err)
	}
}
