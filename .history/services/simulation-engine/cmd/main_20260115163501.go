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
	cfg, err := config.LoadDDefaultConfig(ctx)

	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	sqsClient := sqs.NewFromClient(cfg)

	fmt.Println("Successfully initialized AWS SQS client.")

	messages := make(chan types.Message)

	for i:= 0; i < 5; i++{
		go worker(i, sqsClient, messages, QUEUE_URL)
	}

	for{
		//The & makes it a pointer to the memory location of the object instead of the object itself
		receiveInput := &sqs.RecieveMessageInput{
			MaxNumberOfMessages: 10, 
            WaitTimeSeconds:     20, // Long Polling (Wait up to 20s for data)
            VisibilityTimeout:   30, // Give workers 30s to process before retry
		}

		resp, err := sqsClient.ReceiveMessage(ctx, receiveInput)

		if err != nil{
			fmt.Println("Error receiving messages: ", err)
			continue
		}

		if len(resp.Messages) > 0 {
			for _, msg := range resp.Messages{
				messages <- msg
			}
		}
	}
}