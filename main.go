package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type QueueOperationsRequest struct {
	SourceQueue string
	DestQueue   string
	MessageID   string
	List        bool
}

type SQSClient struct {
	AWSSQSClient   sqs.SQS
	ExecutionCount int
	MessageCount   int
}

func NewSQSClient() (*SQSClient, error) {

	// enable automatic use of AWS_PROFILE like aws cli and other tools do.
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}

	session, err := session.NewSessionWithOptions(opts)
	if err != nil {
		panic(err)
	}

	return &SQSClient{
		AWSSQSClient:   *sqs.New(session),
		ExecutionCount: 0,
		MessageCount:   0,
	}, nil
}

func (c *SQSClient) ListMessages(request QueueOperationsRequest) {

	fmt.Printf("List of Messages in Queue:\t%s\n", request.SourceQueue)

	maxMessages := int64(10)
	waitTime := int64(0)
	messageAttributeNames := aws.StringSlice([]string{"All"})

	rmin := &sqs.ReceiveMessageInput{
		QueueUrl:              &request.SourceQueue,
		MaxNumberOfMessages:   &maxMessages,
		WaitTimeSeconds:       &waitTime,
		MessageAttributeNames: messageAttributeNames,
	}

	lastMessageCount := int(1)
	// loop as long as there are messages on the queue
	for {
		c.ExecutionCount = c.ExecutionCount + 1
		resp, err := c.AWSSQSClient.ReceiveMessage(rmin)

		if err != nil {
			panic(err)
		}
		c.MessageCount = c.MessageCount + len(resp.Messages)

		if lastMessageCount == 0 && len(resp.Messages) == 0 {
			// no messages returned twice now, the queue is probably empty
			//log.Printf("done")
			fmt.Printf("\nMessage Count: %d\nExecution Count:\t%d\n\n", c.MessageCount, c.ExecutionCount)
			return
		}

		lastMessageCount = len(resp.Messages)

		for _, m := range resp.Messages {
			fmt.Printf("MessageId: %s  Body: %s\n", *m.MessageId, *m.Body)
		}
	}
}

func (c *SQSClient) MoveMessage(request QueueOperationsRequest) {
	fmt.Printf("Moving Message\nFrom Queue:\t%s\nTo Queue: \t%s\nMsg ID: \t%s\n", request.SourceQueue, request.DestQueue, request.MessageID)
}

func (c *SQSClient) MoveMessages(request QueueOperationsRequest) {

	fmt.Printf("Moving Messages\nFrom Queue:\t%s\nTo Queue: \t%s\n", request.SourceQueue, request.DestQueue)

	maxMessages := int64(10)
	waitTime := int64(0)
	messageAttributeNames := aws.StringSlice([]string{"All"})

	rmin := &sqs.ReceiveMessageInput{
		QueueUrl:              &request.SourceQueue,
		MaxNumberOfMessages:   &maxMessages,
		WaitTimeSeconds:       &waitTime,
		MessageAttributeNames: messageAttributeNames,
	}

	lastMessageCount := int(1)
	// loop as long as there are messages on the queue
	for {
		c.ExecutionCount = c.ExecutionCount + 1
		resp, err := c.AWSSQSClient.ReceiveMessage(rmin)
		if err != nil {
			panic(err)
		}

		// fmt.Printf(" >Messages Fetched: %d\n", len(resp.Messages))

		if lastMessageCount == 0 && len(resp.Messages) == 0 {
			// no messages returned twice now, the queue is probably empty
			fmt.Printf("\nMessage Count: %d\nExecution Count:\t%d\n\n", c.MessageCount, c.ExecutionCount)
			return
		}

		lastMessageCount = len(resp.Messages)
		// log.Printf("received %v messages...", len(resp.Messages))

		var wg sync.WaitGroup
		wg.Add(len(resp.Messages))

		for _, m := range resp.Messages {
			go func(m *sqs.Message) {
				defer wg.Done()

				// write the message to the destination queue
				smi := sqs.SendMessageInput{
					MessageAttributes: m.MessageAttributes,
					MessageBody:       m.Body,
					QueueUrl:          &request.DestQueue,
				}

				c.ExecutionCount = c.ExecutionCount + 1
				c.MessageCount = c.MessageCount + 1

				fmt.Printf(" >> Moving - MessageId: %s  Body: %s\n", *m.MessageId, *m.Body)

				_, err := c.AWSSQSClient.SendMessage(&smi)

				if err != nil {
					log.Printf("\nERROR sending message to destination %v\n\n", err)
					return
				}

				// message was sent, dequeue from source queue
				dmi := &sqs.DeleteMessageInput{
					QueueUrl:      &request.SourceQueue,
					ReceiptHandle: m.ReceiptHandle,
				}

				if _, err := c.AWSSQSClient.DeleteMessage(dmi); err != nil {
					log.Printf("ERROR dequeueing message ID %v : %v",
						*m.ReceiptHandle,
						err)
				}
			}(m)
		}

		// wait for all jobs from this batch...
		wg.Wait()
	}
	fmt.Printf("Moving Messages Done\nMessages Transferred:\t%d\nExecution Count: %d\n", c.MessageCount, c.ExecutionCount)
}

func main_old() {
	src := flag.String("src", "", "source queue")
	dest := flag.String("dest", "", "destination queue")
	flag.Parse()

	if *src == "" || *dest == "" {
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("source queue : %v", *src)
	log.Printf("destination queue : %v", *dest)

	// enable automatic use of AWS_PROFILE like awscli and other tools do.
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}

	session, err := session.NewSessionWithOptions(opts)
	if err != nil {
		panic(err)
	}

	client := sqs.New(session)

	maxMessages := int64(10)
	waitTime := int64(0)
	messageAttributeNames := aws.StringSlice([]string{"All"})

	rmin := &sqs.ReceiveMessageInput{
		QueueUrl:              src,
		MaxNumberOfMessages:   &maxMessages,
		WaitTimeSeconds:       &waitTime,
		MessageAttributeNames: messageAttributeNames,
	}

	lastMessageCount := int(1)
	// loop as long as there are messages on the queue
	for {
		resp, err := client.ReceiveMessage(rmin)

		if err != nil {
			panic(err)
		}

		if lastMessageCount == 0 && len(resp.Messages) == 0 {
			// no messages returned twice now, the queue is probably empty
			log.Printf("done")
			return
		}

		lastMessageCount = len(resp.Messages)
		log.Printf("received %v messages...", len(resp.Messages))

		var wg sync.WaitGroup
		wg.Add(len(resp.Messages))

		for _, m := range resp.Messages {
			go func(m *sqs.Message) {
				defer wg.Done()

				// write the message to the destination queue
				smi := sqs.SendMessageInput{
					MessageAttributes: m.MessageAttributes,
					MessageBody:       m.Body,
					QueueUrl:          dest,
				}

				_, err := client.SendMessage(&smi)

				if err != nil {
					log.Printf("ERROR sending message to destination %v", err)
					return
				}

				// message was sent, dequeue from source queue
				dmi := &sqs.DeleteMessageInput{
					QueueUrl:      src,
					ReceiptHandle: m.ReceiptHandle,
				}

				if _, err := client.DeleteMessage(dmi); err != nil {
					log.Printf("ERROR dequeueing message ID %v : %v",
						*m.ReceiptHandle,
						err)
				}
			}(m)
		}

		// wait for all jobs from this batch...
		wg.Wait()
	}
}

func main() {

	client, _ := NewSQSClient()
	request := getCmdArguments()
	//client.ListMessages(request)
	routeRequest(request, client)

}

func getCmdArguments() QueueOperationsRequest {

	sourceQueue := flag.String("src", "-BLANK-", "-src >queue>")
	destQueue := flag.String("dest", "-BLANK-", "-dest <queue>")
	messageId := flag.String("id", "-BLANK-", "-id <message id>")
	list := flag.Bool("l", false, "-l")
	flag.Parse()

	return QueueOperationsRequest{
		SourceQueue: *sourceQueue,
		DestQueue:   *destQueue,
		MessageID:   *messageId,
		List:        *list,
	}
}

func routeRequest(req QueueOperationsRequest, client *SQSClient) bool {

	ran := false

	if req.List && len(req.SourceQueue) > 15 {
		// List it
		ran = true
		client.ListMessages(req)
	}

	if !req.List && len(req.SourceQueue) > 15 && len(req.DestQueue) > 15 && len(req.MessageID) > 10 {
		ran = true
		client.MoveMessage(req)
	}

	if !req.List && len(req.SourceQueue) > 15 && len(req.DestQueue) > 15 {
		ran = true
		client.MoveMessages(req)
	}

	return ran
}
