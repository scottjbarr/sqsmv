package main

import (
	"flag"
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
	awsProfile := flag.String("aws-profile", "", "AWS profile")
	flag.Parse()

	if *src == "" || *dest == "" {
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("source queue: %v", *src)
	log.Printf("destination queue: %v", *dest)
	log.Printf("AWS profile: %v", *awsProfile)

	// enable automatic use of AWS_PROFILE like awscli and other tools do.
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}

	// enable AWS_PROFILE selected by name (useful if you configured multiple profiles in awscli)
	if *awsProfile != "" {
		opts.Profile = *awsProfile
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
