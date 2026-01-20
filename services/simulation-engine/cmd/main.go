package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
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
	Eccentricity  float64 `json:"eccentricity,string"`
	SemiMajorAxis float64 `json:"semi_major_axis,string"`

	//Orientation in 3d space
	Inclination            float64 `json:"inclination,string"`
	AscendingNodeLongitude float64 `json:"ascending_node_longitude,string"`
	PerihelionArgument     float64 `json:"perihelion_argument,string"`

	//The Position in time
	MeanAnomaly float64 `json:"mean_anomaly,string"`
	MeanMotion  float64 `json:"mean_motion,string"`
	// Epoch is usually a Julian Date (number), so we can treat it as a float
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
