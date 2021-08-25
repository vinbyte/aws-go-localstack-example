package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
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
	// Create a session that gets credential values from ~/.aws/credentials
	// and the default region from ~/.aws/config
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	sess.Config.Endpoint = &s.localstackURL

	client := sqs.New(sess)

	urlResult, err := client.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &s.queueName,
	})
	if err != nil {
		fmt.Println("Got an error getting the queue URL:", err)
		return err
	}

	queueURL := urlResult.QueueUrl
	timeout := int64(s.timeoutSeconds)

	msgResult, err := client.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            queueURL,
		MaxNumberOfMessages: aws.Int64(1),
		VisibilityTimeout:   &timeout,
	})
	if err != nil {
		fmt.Println("Got an error receiving messages :", err)
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
			_, err := client.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      queueURL,
				ReceiptHandle: msg.ReceiptHandle,
			})
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
