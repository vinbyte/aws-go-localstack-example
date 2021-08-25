package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func receiveSQSMessage() {
	queue := "test-queue1"
	timeout := 5
	ctx := context.Background()

	if queue == "" {
		fmt.Println("You must supply the name of a queue")
		return
	}

	if timeout < 0 {
		timeout = 0
	}

	if timeout > 12*60*60 {
		timeout = 12 * 60 * 60
	}

	endpointURL := "http://localhost:4566"
	region := "us-east-1"
	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if endpointURL != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           endpointURL,
				SigningRegion: region,
			}, nil
		}
		// returning EndpointNotFoundError will allow the service to fallback to it's default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithEndpointResolver(customResolver),
	)
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := sqs.NewFromConfig(cfg)

	gQInput := &sqs.GetQueueUrlInput{
		QueueName: &queue,
	}

	// Get URL of queue
	urlResult, err := client.GetQueueUrl(ctx, gQInput)
	if err != nil {
		fmt.Println("Got an error getting the queue URL:", err)
		return
	}

	queueURL := urlResult.QueueUrl

	gMInput := &sqs.ReceiveMessageInput{
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
		QueueUrl:            queueURL,
		MaxNumberOfMessages: 1,
		VisibilityTimeout:   int32(timeout),
	}

	msgResult, err := client.ReceiveMessage(ctx, gMInput)
	if err != nil {
		fmt.Println("Got an error receiving messages:")
		fmt.Println(err)
		return
	}

	if msgResult != nil {
		messages := msgResult.Messages
		if len(messages) > 0 {
			msg := messages[0]
			fmt.Println("Message ID:     " + *msg.MessageId)
			fmt.Println("Message Body: " + *msg.Body)

			//delete the message
			dMInput := &sqs.DeleteMessageInput{
				QueueUrl:      queueURL,
				ReceiptHandle: msg.ReceiptHandle,
			}
			_, err := client.DeleteMessage(ctx, dMInput)
			if err != nil {
				fmt.Println("failed to delete message", err)
			}
			fmt.Println("Message was deleted")
		}
	}

}

func main() {
	fmt.Println("Start listen incoming message")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	for {
		select {
		case <-quit:
			fmt.Println("app stopped")
			os.Exit(0)
		default:
			receiveSQSMessage()
		}
	}
}
