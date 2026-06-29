package queue

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SQSQueue struct {
	Client   *sqs.Client
	QueueURL string
}

func NewSQSQueue(client *sqs.Client, queueURL string) *SQSQueue {
	return &SQSQueue{
		Client:   client,
		QueueURL: queueURL,
	}
}

func (q *SQSQueue) Publish(ctx context.Context, body []byte, attributes map[string]string) error {
	messageAttributes := map[string]types.MessageAttributeValue{}

	for k, v := range attributes {
		messageAttributes[k] = types.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(v),
		}
	}

	_, err := q.Client.SendMessage(
		ctx,
		&sqs.SendMessageInput{
			QueueUrl:          aws.String(q.QueueURL),
			MessageBody:       aws.String(string(body)),
			MessageAttributes: messageAttributes,
		},
	)

	return err
}
