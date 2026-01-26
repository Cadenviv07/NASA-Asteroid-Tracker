package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cva70/nasa-asteroid-tracker/simulation-engine/pkg/alerts"
	"github.com/cva70/nasa-asteroid-tracker/simulation-engine/pkg/physics"
	"github.com/cva70/nasa-asteroid-tracker/simulation-engine/pkg/storage"
	"github.com/joho/godotenv"
)

// object for all meteorites
type Asteroid struct {
	ID       string  `json:"id"`
	Name     string  `json:"asteroid"` // Python sends "asteroid", we map to Name
	Diameter float64 `json:"diameter_km"`
	Velocity float64 `json:"velocity_kph"`

	OrbitalData OrbitalData `json:"orbital_elements"`
}

type OrbitalData struct {
	// ,string takes the string json is giving and converts it into a float
	//Shape of the orbit
	//Radius of the long part og the oval
	Eccentricity float64 `json:"eccentricity,string"`
	//How ovular it is
	SemiMajorAxis float64 `json:"semi_major_axis,string"`

	//Orientation in 3d space
	//Angle that the orbit is on
	Inclination float64 `json:"inclination,string"`
	//Draw a line from where the asteroid orbit intersects with earths orbit to the nearest star it points at
	AscendingNodeLongitude float64 `json:"ascending_node_longitude,string"`
	//Where is the low point
	PerihelionArgument float64 `json:"perihelion_argument,string"`

	//The Position in time
	//Where the asteroid would be if it moved at constant speed it is a angle
	MeanAnomaly float64 `json:"mean_anomaly,string"`
	//How fast mean anomaly is changing
	MeanMotion float64 `json:"mean_motion,string"`
	// Epoch is usually a Julian Date (number), so we can treat it as a float
	//Date when these numbers were measured
	EpochOsculation float64 `json:"epoch_osculation,string"`

	//Pre-calculated Risk (Good for checking math later)
	MinimumOrbitIntersection string `json:"minimum_orbit_intersection"`
}

// process everything that comes through sqs pipeline
func worker(id int, snsClient *sns.Client, db *sql.DB, client *sqs.Client, pipe <-chan types.Message, queueURL string, snsTopicArn string) {
	for message := range pipe {

		fmt.Printf("Worker %d: Processing %s\n", id, *message.Body)

		startDate := physics.TimetoJulian(time.Now())
		endDate := startDate + 100*365
		timeStep := 1.0 / 24.0
		minDistance := math.Inf(1)

		var asteroidData Asteroid

		err := json.Unmarshal([]byte(*message.Body), &asteroidData)
		if err != nil {
			fmt.Println("ERROR PARSING JSON: ", err)
			continue
		}

		for {

			EarthPosition := physics.GetEarthsPosition(startDate)

			positionInOrbit := physics.PositionInOrbit(startDate, asteroidData.OrbitalData.MeanMotion, asteroidData.OrbitalData.MeanAnomaly, asteroidData.OrbitalData.EpochOsculation)

			EccentricAnomaly := physics.CalculateEccentricAnomaly(positionInOrbit, asteroidData.OrbitalData.Eccentricity)

			x, y := physics.GetPlaneCoordinates(EccentricAnomaly, asteroidData.OrbitalData.Eccentricity, asteroidData.OrbitalData.SemiMajorAxis)

			AsteroidCoordinates := physics.RotatePlane(x, y, asteroidData.OrbitalData.Inclination, asteroidData.OrbitalData.AscendingNodeLongitude, asteroidData.OrbitalData.PerihelionArgument)

			distance := math.Sqrt(math.Pow((EarthPosition.X-AsteroidCoordinates.X), 2) + math.Pow((EarthPosition.Y-AsteroidCoordinates.Y), 2) + math.Pow((EarthPosition.Z-AsteroidCoordinates.Z), 2))

			distanceKm := distance * 149600000

			if distanceKm < 6000 {
				fmt.Printf("COLLISION DETECTED! Date: %f\n", startDate)

				alerts.SendCollisionAlert(snsClient, snsTopicArn, asteroidData.Name, startDate, distanceKm)

				storage.SaveTrajectory(db, asteroidData.ID, asteroidData.Name, distanceKm, startDate, true)

				break
			}

			minDistance = math.Min(minDistance, distanceKm)

			if minDistance < 50_000_000 && distanceKm > (minDistance+1_000_000) {
				break
			}

			if distanceKm < 10_000_000 {
				timeStep = 1.0 / 1440.0
			} else {
				timeStep = 1.0 / 24.0
			}

			startDate += timeStep
			if startDate > endDate {
				break
			}
		}
		if minDistance >= 6000 {
			storage.SaveTrajectory(db, asteroidData.ID, asteroidData.Name, minDistance, 0.0, false)
		}

		_, er := client.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
			QueueUrl: &queueURL,

			ReceiptHandle: message.ReceiptHandle,
		})

		if er != nil {
			fmt.Println("Error deleting message:", er)
		}
	}
}

const QueueUrl = "https://sqs.us-east-2.amazonaws.com/574070665369/asteroidBelt"

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Println("Error loading .env file. Ensure it exists in the same directory as main.go")
	}

	ctx := context.TODO()
	//Function either creates cfg or error
	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	sqsClient := sqs.NewFromConfig(cfg)

	snsClient := sns.NewFromConfig(cfg)
	snsTopicArn := os.Getenv("AWS_SNS_TOPIC_ARN")

	dbConnString := os.Getenv("DB_CONNECTION_STRING")
	db := storage.InitializeDB(dbConnString)
	defer db.Close()

	fmt.Println("Successfully initialized AWS SQS client.")
	//Create channel
	messages := make(chan types.Message)

	for i := 0; i < 5; i++ {
		go worker(i, snsClient, db, sqsClient, messages, QueueUrl, snsTopicArn)
	}

	for {
		//The & makes it a pointer to the memory location of the object instead of the object itself
		receiveInput := &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(QueueUrl),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20, // Long Polling (Wait up to 20s for data)
			VisibilityTimeout:   30, // Give workers 30s to process before retry
		}

		resp, err := sqsClient.ReceiveMessage(ctx, receiveInput)

		if err != nil {
			fmt.Println("Error receiving messages: ", err)
			continue
		}

		if len(resp.Messages) > 0 {
			for _, msg := range resp.Messages {
				// Push the messages onto the channel for the
				messages <- msg
			}
		}
	}
}
