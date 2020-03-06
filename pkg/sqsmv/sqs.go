package sqsmv

import (
	"k8s.io/klog"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func getSQSClient() *sqs.SQS {
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}
	session, err := session.NewSessionWithOptions(opts)
	if err != nil {
		klog.Fatalf("error creating sqs client, err: %v", err)
	}

	return sqs.New(session)
}

func receiveMessage(queue string) ([]*sqs.Message, error) {
	resp, err := getSQSClient().ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(queue),
		MaxNumberOfMessages:   aws.Int64(10),
		WaitTimeSeconds:       aws.Int64(0),
		MessageAttributeNames: aws.StringSlice([]string{"All"}),
	})
	if err != nil {
		return nil, err
	}

	return resp.Messages, nil
}

func longPollReceiveMessage(queue string) (int32, error) {
	result, err := getSQSClient().ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(queue),
		AttributeNames:        aws.StringSlice([]string{"SentTimestamp"}),
		VisibilityTimeout:     aws.Int64(0),
		MaxNumberOfMessages:   aws.Int64(1),
		MessageAttributeNames: aws.StringSlice([]string{"All"}),
		WaitTimeSeconds:       aws.Int64(20),
	})
	if err != nil {
		return 0, err
	}

	return int32(len(result.Messages)), nil
}

func writeMessage(m *sqs.Message, queue string) error {
	_, err := getSQSClient().SendMessage(&sqs.SendMessageInput{
		MessageAttributes: m.MessageAttributes,
		MessageBody:       m.Body,
		QueueUrl:          aws.String(queue),
	})

	return err
}

func deleteMessage(m *sqs.Message, queue string) error {
	_, err := getSQSClient().DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queue),
		ReceiptHandle: m.ReceiptHandle,
	})
	return err
}
