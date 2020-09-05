package sqsmv

import (
	"errors"
	"strings"

	"k8s.io/klog"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SQSQueue struct {
	name   string
	sqssvc *sqs.SQS
	url    string
}

func getRegion(url string) string {
	return strings.Split(url, ".")[1]
}

func getQueueName(url string) string {
	splitted := strings.Split(url, "/")
	return splitted[len(splitted)-1]
}

func getSQSClient(queueName string) *sqs.SQS {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(
			getRegion(queueName),
		),
	},
	)
	if err != nil {
		klog.Fatalf("error creating sqs client, err: %v", err)
	}

	return sqs.New(sess)
}

func NewSQSQueue(url string) *SQSQueue {
	return &SQSQueue{
		name:   getQueueName(url),
		sqssvc: getSQSClient(url),
		url:    url,
	}
}

func (s *SQSQueue) Describe() (QueueDetails, error) {
	queueDetails := QueueDetails{}

	queueURL, err := s.getQueueURL()
	if err != nil {
		return queueDetails, err
	}

	queueAttributes, err := s.getQueueAttributes(queueURL)
	if err != nil {
		return queueDetails, err
	}

	return s.buildQueueDetails(queueURL, queueAttributes), nil
}

func (s *SQSQueue) Create(queueDetails QueueDetails) (string, error) {
	createQueueInput := s.buildCreateQueueInput(s.name, queueDetails)

	createQueueOutput, err := s.sqssvc.CreateQueue(createQueueInput)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return "", errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return "", err
	}

	return aws.StringValue(createQueueOutput.QueueUrl), nil
}

func (s *SQSQueue) getQueueURL() (string, error) {
	getQueueURLInput := &sqs.GetQueueUrlInput{
		QueueName: aws.String(s.name),
	}

	getQueueURLOutput, err := s.sqssvc.GetQueueUrl(getQueueURLInput)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// AWS SQS returns a 400 if Queue is not found
				if reqErr.StatusCode() == 400 || reqErr.StatusCode() == 404 {
					return "", ErrQueueDoesNotExist
				}
			}
			return "", errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return "", err
	}

	return aws.StringValue(getQueueURLOutput.QueueUrl), nil
}

func (s *SQSQueue) getQueueAttributes(queueURL string) (map[string]string, error) {
	getQueueAttributesInput := &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: aws.StringSlice([]string{"All"}),
	}

	getQueueAttributesOutput, err := s.sqssvc.GetQueueAttributes(getQueueAttributesInput)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// AWS SQS returns a 400 if Queue is not found
				if reqErr.StatusCode() == 400 || reqErr.StatusCode() == 404 {
					return nil, ErrQueueDoesNotExist
				}
			}
			return nil, errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return nil, err
	}

	return aws.StringValueMap(getQueueAttributesOutput.Attributes), nil
}

func (s *SQSQueue) setQueueAttributes(queueURL string, queueDetails QueueDetails) error {
	setQueueAttributesInput := s.buildSetQueueAttributesInput(queueURL, queueDetails)

	_, err := s.sqssvc.SetQueueAttributes(setQueueAttributesInput)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// AWS SQS returns a 400 if Queue is not found
				if reqErr.StatusCode() == 400 || reqErr.StatusCode() == 404 {
					return ErrQueueDoesNotExist
				}
			}
			return errors.New(awsErr.Code() + ": " + awsErr.Message())
		}
		return err
	}

	return nil
}

func (s *SQSQueue) buildQueueDetails(queueURL string, attributes map[string]string) QueueDetails {
	queueDetails := QueueDetails{
		QueueURL:                      queueURL,
		DelaySeconds:                  attributes["DelaySeconds"],
		MaximumMessageSize:            attributes["MaximumMessageSize"],
		MessageRetentionPeriod:        attributes["MessageRetentionPeriod"],
		Policy:                        attributes["Policy"],
		ReceiveMessageWaitTimeSeconds: attributes["ReceiveMessageWaitTimeSeconds"],
		VisibilityTimeout:             attributes["VisibilityTimeout"],
	}

	return queueDetails
}

func (s *SQSQueue) buildCreateQueueInput(queueName string, queueDetails QueueDetails) *sqs.CreateQueueInput {
	createQueueInput := &sqs.CreateQueueInput{
		QueueName:  aws.String(queueName),
		Attributes: map[string]*string{},
	}

	if queueDetails.DelaySeconds != "" {
		createQueueInput.Attributes["DelaySeconds"] = aws.String(queueDetails.DelaySeconds)
	}

	if queueDetails.MaximumMessageSize != "" {
		createQueueInput.Attributes["MaximumMessageSize"] = aws.String(queueDetails.MaximumMessageSize)
	}

	if queueDetails.MessageRetentionPeriod != "" {
		createQueueInput.Attributes["MessageRetentionPeriod"] = aws.String(queueDetails.MessageRetentionPeriod)
	}

	if queueDetails.Policy != "" {
		createQueueInput.Attributes["Policy"] = aws.String(queueDetails.Policy)
	}

	if queueDetails.ReceiveMessageWaitTimeSeconds != "" {
		createQueueInput.Attributes["ReceiveMessageWaitTimeSeconds"] = aws.String(queueDetails.ReceiveMessageWaitTimeSeconds)
	}

	if queueDetails.VisibilityTimeout != "" {
		createQueueInput.Attributes["VisibilityTimeout"] = aws.String(queueDetails.VisibilityTimeout)
	}

	return createQueueInput
}

func (s *SQSQueue) buildSetQueueAttributesInput(queueURL string, queueDetails QueueDetails) *sqs.SetQueueAttributesInput {
	setQueueAttributesInput := &sqs.SetQueueAttributesInput{
		QueueUrl:   aws.String(queueURL),
		Attributes: map[string]*string{},
	}

	if queueDetails.DelaySeconds != "" {
		setQueueAttributesInput.Attributes["DelaySeconds"] = aws.String(queueDetails.DelaySeconds)
	}

	if queueDetails.MaximumMessageSize != "" {
		setQueueAttributesInput.Attributes["MaximumMessageSize"] = aws.String(queueDetails.MaximumMessageSize)
	}

	if queueDetails.MessageRetentionPeriod != "" {
		setQueueAttributesInput.Attributes["MessageRetentionPeriod"] = aws.String(queueDetails.MessageRetentionPeriod)
	}

	if queueDetails.Policy != "" {
		setQueueAttributesInput.Attributes["Policy"] = aws.String(queueDetails.Policy)
	}

	if queueDetails.ReceiveMessageWaitTimeSeconds != "" {
		setQueueAttributesInput.Attributes["ReceiveMessageWaitTimeSeconds"] = aws.String(queueDetails.ReceiveMessageWaitTimeSeconds)
	}

	if queueDetails.VisibilityTimeout != "" {
		setQueueAttributesInput.Attributes["VisibilityTimeout"] = aws.String(queueDetails.VisibilityTimeout)
	}

	return setQueueAttributesInput
}
