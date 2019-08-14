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

func main() {
	src := flag.String("src", "", "source queue")
	dest := flag.String("dest", "", "destination queue")
	limit := flag.Int("limit", -1, "limit number of messages moved (defaults to all)")
	flag.Parse()

	if *src == "" || *dest == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *limit < -1 {
		flag.Usage()
		os.Exit(1)
	}

	limitDescription := "no limit"

	if *limit != -1 {
		limitDescription = fmt.Sprintf("%d messages", *limit)
	}

	log.Printf("source queue : %v", *src)
	log.Printf("destination queue : %v", *dest)
	log.Printf("limit : %v", limitDescription)

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

	movedMessageCount := 0
	lastMessageCount := int(1)

	// loop as long as there are messages on the queue, or we've reached the limit
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

		for _, m := range resp.Messages {
			go func(m *sqs.Message) {
				if (*limit == -1 || movedMessageCount <= *limit) {
					wg.Add(1)

					movedMessageCount++

					defer wg.Done()

					// write the message to the destination queue
					smi := sqs.SendMessageInput{
						MessageAttributes: m.MessageAttributes,
						MessageBody:	   m.Body,
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
				}
			}(m)
		}

		// wait for all jobs from this batch...
		wg.Wait()

		if (*limit != -1 && movedMessageCount >= *limit) {
			break;
		}
	}
}
