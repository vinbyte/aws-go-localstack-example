# Example SNS Localstack using aws-sdk-go-v2

## Usage 

1. Create the SNS topic and copy the `TopicArn` as a result from this command

```
aws sns create-topic --name <your_topic>
```

2. Run the code

```
go run main.go -t <the_topic_arn> -m <your_message>
```

### Available options

| name | desc | mandatory |
| --- | --- | --- |
| `t` | The ARN of the topic to which the user subscribes | :white_check_mark: |
| `m` | The message to send to the subscribed users of the topic | :white_check_mark: |
| `r` | The region of localstack. Default: us-east-1 |  |
| `u` | The Localstack url. Default: http://localhost:4566| |