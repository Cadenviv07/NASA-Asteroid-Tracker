package main

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cva70/nasa-asteroid-tracker/simulation-engine/pkg/physics"
	"github.com/joho/godotenv"
)

func TestPhysics() {
	fmt.Println("--- STARTING PHYSICS CHECK ---")

	// 1. Create a Fake Asteroid (Standard Test Case)
	// Let's use Halley's Comet (very elliptical, easy to spot errors)
	// a = 17.8 AU, e = 0.967
	testAsteroid := OrbitalElements{
		SemiMajorAxis: 17.834,
		Eccentricity:  0.96714,
		Inclination:   162.26,
		AscendingNode: 58.42,
		Perihelion:    111.33,
		MeanAnomaly:   38.38, // Random spot on track
	}

	// 2. Run your Physics Engine
	// (You'll need to manually calculate M or just pass the M from the struct)
	// For this test, let's just assume we calculated M correctly and test the geometry steps:

	M := (math.Pi / 180.0) * testAsteroid.MeanAnomaly
	E := physics.CalculateEccentricAnomaly(M, testAsteroid.Eccentricity)

	xPlane, yPlane := physics.GetPlaneCoordinates(E, testAsteroid.Eccentricity, testAsteroid.SemiMajorAxis)

	pos := physics.RotatePlane(xPlane, yPlane, testAsteroid.Inclination, testAsteroid.AscendingNode, testAsteroid.Perihelion)

	// 3. Calculate the Distance from Sun (Vector Magnitude)
	// Sqrt(x^2 + y^2 + z^2)
	distance := math.Sqrt(pos.X*pos.X + pos.Y*pos.Y + pos.Z*pos.Z)

	// 4. Calculate Allowed Range
	perihelion := testAsteroid.SemiMajorAxis * (1 - testAsteroid.Eccentricity) // Min Dist
	aphelion := testAsteroid.SemiMajorAxis * (1 + testAsteroid.Eccentricity)   // Max Dist

	fmt.Printf("Calculated Position: X=%.2f Y=%.2f Z=%.2f\n", pos.X, pos.Y, pos.Z)
	fmt.Printf("Distance from Sun: %.4f AU\n", distance)
	fmt.Printf("Allowed Range:     %.4f - %.4f AU\n", perihelion, aphelion)

	if distance >= perihelion && distance <= aphelion {
		fmt.Println("✅ PASSED: Position is valid within orbit limits.")
	} else {
		fmt.Println("❌ FAILED: Position is impossible (outside the ellipse).")
	}
}

type OrbitalElements struct {
	SemiMajorAxis float64
	Eccentricity  float64
	Inclination   float64
	AscendingNode float64
	Perihelion    float64
	MeanAnomaly   float64
}

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
func worker(id int, client *sqs.Client, pipe <-chan types.Message, queueURL string) {
	for message := range pipe {

		fmt.Printf("Worker %d: Processing %s\n", id, *message.Body)

		_, err := client.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
			QueueUrl: &queueURL,

			ReceiptHandle: message.ReceiptHandle,
		})

		if err != nil {
			fmt.Println("Error deleting message:", err)
		}
	}
}

const QueueUrl = "https://sqs.us-east-2.amazonaws.com/574070665369/asteroidBelt"

func main() {

	TestPhysics()

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

	fmt.Println("Successfully initialized AWS SQS client.")
	//Create channel
	messages := make(chan types.Message)

	for i := 0; i < 5; i++ {
		go worker(i, sqsClient, messages, QueueUrl)
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
