# Example SQS Localstack using aws-sdk-go-v2

## Usage 

1. Create the queue
```
aws sqs create-queue --queue-name <your_queue_name>
```

2. Run the code
```
go run main.go -q <queue_name>
```

### Available options

| name | desc | mandatory |
| --- | --- | --- |
| `q` | The name of the queue | :white_check_mark: |
| `r` | The region of localstack. Default: us-east-1 |  |
| `t` | How long, in seconds, that the message is hidden from other | |
| `u` | The Localstack url. Default: http://localhost:4566| |