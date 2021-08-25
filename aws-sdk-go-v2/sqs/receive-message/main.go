package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SQSLocalstack struct {
	region         string
	queueName      string
	timeoutSeconds int
	localstackURL  string
}

// setup the data by collecting from args parameter
func (s *SQSLocalstack) setup() {
	region := flag.String("r", "us-east-1", "The region of localstack. Default: us-east-1")
	queue := flag.String("q", "", "The name of the queue")
	timeout := flag.Int("t", 5, "How long, in seconds, that the message is hidden from others")
	localstackURL := flag.String("u", "http://localhost:4566", "The Localstack url. Default : http://localhost:4566")
	flag.Parse()

	if *queue == "" {
		fmt.Println("You must supply the name of a queue (-q QUEUE)")
		return
	}

	if *timeout < 0 {
		*timeout = 0
	}

	if *timeout > 12*60*60 {
		*timeout = 12 * 60 * 60
	}
	s.region = *region
	s.queueName = *queue
	s.timeoutSeconds = *timeout
	s.localstackURL = *localstackURL
}

func (s *SQSLocalstack) receiveSQSMessage() error {
	ctx := context.TODO()

	// customResolver used to handle the localstack url we given to aws-sdk-go-v2
	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if s.localstackURL != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           s.localstackURL,
				SigningRegion: s.region,
			}, nil
		}
		// returning EndpointNotFoundError will allow the service to fallback to it's default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(s.region),
		config.WithEndpointResolver(customResolver),
	)
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := sqs.NewFromConfig(cfg)

	gQInput := &sqs.GetQueueUrlInput{
		QueueName: &s.queueName,
	}

	// Get URL of queue
	urlResult, err := client.GetQueueUrl(ctx, gQInput)
	if err != nil {
		fmt.Println("Got an error getting the queue URL:", err)
		return err
	}

	queueURL := urlResult.QueueUrl

	gMInput := &sqs.ReceiveMessageInput{
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
		QueueUrl:            queueURL,
		MaxNumberOfMessages: 1,
		VisibilityTimeout:   int32(s.timeoutSeconds),
	}

	msgResult, err := client.ReceiveMessage(ctx, gMInput)
	if err != nil {
		fmt.Println("Got an error receiving messages:")
		fmt.Println(err)
		return err
	}

	if msgResult != nil {
		messages := msgResult.Messages
		if len(messages) > 0 {
			msg := messages[0]
			fmt.Println("Message received")
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
				return err
			}
			fmt.Println("Message succesfully deleted")
		}
	}
	return nil
}

func main() {
	sqsLocalstack := SQSLocalstack{}

	sqsLocalstack.setup()

	fmt.Println("Start listen incoming message")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	for {
		select {
		case <-quit:
			fmt.Println("app stopped")
			os.Exit(0)
		default:
			err := sqsLocalstack.receiveSQSMessage()
			if err != nil {
				os.Exit(1)
			}
		}
	}
}
