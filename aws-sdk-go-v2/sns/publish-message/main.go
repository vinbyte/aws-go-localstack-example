package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
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
		return
	}

	s.topicARN = *topicARN
	s.message = *msg
	s.localstackURL = *localstackURL
	s.region = *region
}

func (s *SNSLocalstack) publishMessage() {
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

	client := sns.NewFromConfig(cfg)

	input := &sns.PublishInput{
		Message:  &s.message,
		TopicArn: &s.topicARN,
	}

	result, err := client.Publish(ctx, input)
	if err != nil {
		fmt.Println("Got an error publishing the message:")
		fmt.Println(err)
		return
	}

	fmt.Println("Message ID: " + *result.MessageId)
}

func main() {
	snsLocalstack := SNSLocalstack{}

	snsLocalstack.setup()
	snsLocalstack.publishMessage()
}
