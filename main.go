package main

import (
	"flag"
	"log"
	"math"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	maxInt := int(^uint(0) >> 1)
	src := flag.String("src", "", "source queue")
	dest := flag.String("dest", "", "destination queue")
	maxMsgsToMove := flag.Int("max", maxInt, "max number of messages to move")
	flag.Parse()

	if *src == "" || *dest == "" {
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("source queue : %v", *src)
	log.Printf("destination queue : %v", *dest)
	log.Printf("max number of messages to move : %v", *maxMsgsToMove)

	if *maxMsgsToMove <= 0 {
		log.Printf("max number of message to move : %v must be greater than zero", *maxMsgsToMove)
		os.Exit(1)
	}

	// enable automatic use of AWS_PROFILE like awscli and other tools do.
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}

	session, err := session.NewSessionWithOptions(opts)
	if err != nil {
		panic(err)
	}

	client := sqs.New(session)

	maxMessages := int64(math.Min(float64(*maxMsgsToMove), float64(10)))
	waitTime := int64(0)
	messageAttributeNames := aws.StringSlice([]string{"All"})

	rmin := &sqs.ReceiveMessageInput{
		QueueUrl:              src,
		MaxNumberOfMessages:   &maxMessages,
		WaitTimeSeconds:       &waitTime,
		MessageAttributeNames: messageAttributeNames,
	}

	var mutex = &sync.Mutex{}
	var count int
	lastMessageCount := int(1)
	// loop as long as there are messages on the queue
	for {
		resp, err := client.ReceiveMessage(rmin)

		if err != nil {
			panic(err)
		}

		if count >= *maxMsgsToMove || (lastMessageCount == 0 && len(resp.Messages) == 0) {
			// no messages returned twice now, the queue is probably empty
			log.Printf("done")
			return
		}

		lastMessageCount = len(resp.Messages)
		log.Printf("received %v messages...", len(resp.Messages))

		var wg sync.WaitGroup
		wg.Add(len(resp.Messages))

		for _, m := range resp.Messages {
			if count >= *maxMsgsToMove {
				break
			}

			go func(m *sqs.Message) {
				defer wg.Done()

				mutex.Lock()
				defer mutex.Unlock()
				if count < *maxMsgsToMove {
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
					count++
				}
			}(m)
		}

		// wait for all jobs from this batch...
		wg.Wait()
	}
}
