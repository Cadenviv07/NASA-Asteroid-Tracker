package alerts

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func SendCollisionAlert(client *sns.Client, topicArn string, asteroidName string, date float64, distance float64) {
	message := fmt.Sprintf(
		"CRITICAL ALERT: IMPACT DETECTED\n\nAsteroid: %s\nImpact Date (Julian): %.2f\nMiss Distance: %.2f km\n\nCheck dashboard immediately.",
		asteroidName, date, distance,
	)

	_, err := client.Publish(context.TODO(), &sns.PublishInput{
		Message:  aws.String(message),
		TopicArn: aws.String(topicArn),
	})

	if err != nil {
		fmt.Printf("FAILED TO SEND ALERT: %v\n", err)
	} else {
		fmt.Println("SNS Alert Sent successfully.")
	}
}
