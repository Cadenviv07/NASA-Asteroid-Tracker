package main

import (
	"context"
	"fmt"
	"log"

	"github.com"
	"github.com"
)


//object for all meteorites 
type asteroid struct {
	id: str
    asteroid: str
    diameter_km: float
    velocity_kph: float
    orbital_elements:Dict
	'json: "id"'
	'json: "asteorid"'
	'json: "diameter_km"'
	'json: "velocity_kph"'
	'json: "orbital_elements'
}

//process everything that comes through sqs pipeline
func worker(id int, pipe <-chan *sqs.message, queueURL string){
	for  message := range pipe {
		
		fmt.Printf("Worker %d: Processing %s\n", id, *message.Body)

		_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl: &queueURL,

			ReceiptHandle: message.ReceiptHandle,
		})

		if err != nil {
			fmt.Println("Error deleting message:", err)
		}		
	} 
}

QUEUE_URL = "https://sqs.us-east-2.amazonaws.com/574070665369/asteroidBelt"

func main(){
	ctx := context.TODO()
	//Function either creates cfg or error
	cfg, err := config.LOADDefaultConfig(ctx)

	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	sqsClient := sqs.NewFromClient(cfg)

	fmt.Println("Successfully initialized AWS SQS client.")

}