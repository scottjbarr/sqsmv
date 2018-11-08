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
	srcRegion := flag.String("srcRegion", "", "source region")
	destRegion := flag.String("destRegion", "", "destination region")
	flag.Parse()

	if *src == "" || *dest == "" {
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("source queue : %v", *src)
	log.Printf("destination queue : %v", *dest)
	if *srcRegion != "" {
		log.Printf("source region : %v", *srcRegion)
	}
	if *destRegion != "" {
		log.Printf("destination region : %v", *destRegion)
	}

	// enable automatic use of AWS_PROFILE like awscli and other tools do.
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}

	sessionSrc, err := session.NewSessionWithOptions(opts)
	if err != nil {
		panic(err)
	}
	sessionDst, err := session.NewSessionWithOptions(opts)
	if err != nil {
		panic(err)
	}

	if *srcRegion == "" {
		clientSrc := sqs.New(sessionSrc)
	} else {
		clientSrc := sqs.New(sessionSrc, aws.NewConfig().WithRegion(srcRegion))
	}

	if *destRegion == "" {
		clientDest := sqs.New(sessionDst)
	} else {
		clientDest := sqs.New(sessionDst, aws.NewConfig().WithRegion(destRegion))
	}

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
		resp, err := clientSrc.ReceiveMessage(rmin)

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

				_, err := clientDest.SendMessage(&smi)

				if err != nil {
					log.Printf("ERROR sending message to destination %v", err)
					return
				}

				// message was sent, dequeue from source queue
				dmi := &sqs.DeleteMessageInput{
					QueueUrl:      src,
					ReceiptHandle: m.ReceiptHandle,
				}

				if _, err := clientSrc.DeleteMessage(dmi); err != nil {
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
