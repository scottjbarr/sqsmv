package sqsmv

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func receiveMessage(queueName string, longPollSeconds int64) ([]*sqs.Message, error) {
	resp, err := getSQSClient(queueName).ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(queueName),
		AttributeNames:        aws.StringSlice([]string{"SentTimestamp"}),
		MaxNumberOfMessages:   aws.Int64(10),
		WaitTimeSeconds:       aws.Int64(longPollSeconds),
		MessageAttributeNames: aws.StringSlice([]string{"All"}),
	})
	if err != nil {
		return nil, err
	}

	return resp.Messages, nil
}

func writeMessage(m *sqs.Message, queueName string) error {
	_, err := getSQSClient(queueName).SendMessage(&sqs.SendMessageInput{
		MessageAttributes: m.MessageAttributes,
		MessageBody:       m.Body,
		QueueUrl:          aws.String(queueName),
	})

	return err
}

func deleteMessage(m *sqs.Message, queueName string) error {
	_, err := getSQSClient(queueName).DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueName),
		ReceiptHandle: m.ReceiptHandle,
	})
	return err
}
