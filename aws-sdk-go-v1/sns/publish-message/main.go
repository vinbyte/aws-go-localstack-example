package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

type SNSLocalstack struct {
	message       string
	topicARN      string
	localstackURL string
	region        string
}

// setup the data by collecting from args parameter
func (s *SNSLocalstack) setup() {
	msg := flag.String("m", "", "The message to send to the subscribed users of the topic")
	topicARN := flag.String("t", "", "The ARN of the topic to which the user subscribes")
	localstackURL := flag.String("u", "http://localhost:4566", "The Localstack url. Default : http://localhost:4566")
	region := flag.String("r", "us-east-1", "The region of localstack. Default: us-east-1")

	flag.Parse()

	if *msg == "" || *topicARN == "" {
		fmt.Println("You must supply a message and topic ARN")
		fmt.Println("-m MESSAGE -t TOPIC-ARN")
		os.Exit(1)
	}

	s.topicARN = *topicARN
	s.message = *msg
	s.localstackURL = *localstackURL
	s.region = *region
}

func (s *SNSLocalstack) publishMessage() {
	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file. (~/.aws/credentials).
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	sess.Config.Endpoint = &s.localstackURL

	svc := sns.New(sess)

	result, err := svc.Publish(&sns.PublishInput{
		Message:  &s.message,
		TopicArn: &s.topicARN,
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(*result.MessageId)
}

func main() {
	snsLocalstack := SNSLocalstack{}

	snsLocalstack.setup()
	snsLocalstack.publishMessage()
}
