package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	src := flag.String("src", "", "source queue url")
	dest := flag.String("dest", "", "destination queue url")
	flag.Parse()

	if *src == "" || *dest == "" {
		flag.Usage()
		os.Exit(1)
	}

	srcClient := sqs.New(session.New(), aws.NewConfig().WithRegion(getRegionFromQueueURL(*src)))
	destClient := sqs.New(session.New(), aws.NewConfig().WithRegion(getRegionFromQueueURL(*dest)))

	maxMessages := int64(10)
	waitTime := int64(0)
	messageAttributeNames := aws.StringSlice([]string{"All"})

	rmin := &sqs.ReceiveMessageInput{
		QueueUrl:              src,
		MaxNumberOfMessages:   &maxMessages,
		WaitTimeSeconds:       &waitTime,
		MessageAttributeNames: messageAttributeNames,
	}

	// loop as long as there are messages on the queue
	for {
		resp, err := srcClient.ReceiveMessage(rmin)

		if err != nil {
			panic(err)
		}

		if len(resp.Messages) == 0 {
			log.Printf("done")
			return
		}

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

				_, err := destClient.SendMessage(&smi)

				if err != nil {
					log.Printf("ERROR sending message to destination %v", err)
					return
				}

				// message was sent, dequeue from source queue
				dmi := &sqs.DeleteMessageInput{
					QueueUrl:      src,
					ReceiptHandle: m.ReceiptHandle,
				}

				if _, err := srcClient.DeleteMessage(dmi); err != nil {
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

func getRegionFromQueueURL(url string) string {
    return strings.Split(url, ".")[1]
}
