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

type SimulationResult struct {
	ID                int     `json:"id"`
	AsteroidID        string  `json:"asteroid_id"`
	Name              string  `json:"name"`
	ClosestDistanceKm float64 `json:"closest_distance_km"`
	ImpactDate        float64 `json:"impact_date"`
	IsDangerous       bool    `json:"is_dangerous"`
}

func GetDangerousAsteroids(db *sql.DB) ([]SimulationResult, error) {
	query := `
        SELECT id, asteroid_id, name, closest_distance_km, impact_date, is_dangerous 
        FROM simulation_results 
        WHERE is_dangerous = true 
        ORDER BY closest_distance_km ASC 
        LIMIT 10
    `

	rows, err := db.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var results []SimulationResult

	for rows.Next() {
		var r SimulationResult

		if err := rows.Scan(&r.ID, &r.AsteroidID, &r.Name, &r.ClosestDistanceKm, &r.ImpactDate, &r.IsDangerous); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}
