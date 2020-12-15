package main

import (
	"flag"
	"log"
	"os"
	"regexp"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	src := flag.String("src", "", "source queue")
	dest := flag.String("dest", "", "destination queue")
	flag.Parse()

	if *src == "" || *dest == "" {
		flag.Usage()
		os.Exit(1)
	}

	re := regexp.MustCompile(`[a-z]{2}-(?:north|south|east|west)[^-]*-\d`)
	srcRegion := re.FindString(*src)
	destRegion := re.FindString(*dest)

	log.Printf("source queue : %v (%v)", *src, srcRegion)
	log.Printf("destination queue : %v (%v)", *dest, destRegion)

	// enable automatic use of AWS_PROFILE like awscli and other tools do.
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}

	session, err := session.NewSessionWithOptions(opts)
	if err != nil {
		panic(err)
	}

	srcClient := sqs.New(session, aws.NewConfig().WithRegion(srcRegion))
	destClient := sqs.New(session, aws.NewConfig().WithRegion(destRegion))

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
		resp, err := srcClient.ReceiveMessage(rmin)

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
