package main

import (
	"flag"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/shomali11/parallelizer"
)

func main() {
	src := flag.String("src", "", "source queue")
	dest := flag.String("dest", "", "destination queue")
	numClients := flag.Int("numClients", 1, "number of clients")
	flag.Parse()

	if *src == "" || *dest == "" || *numClients < 1 {
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("source queue : %v", *src)
	log.Printf("destination queue : %v", *dest)
	log.Printf("number of clients : %v", *numClients)

	// enable automatic use of AWS_PROFILE like awscli and other tools do.
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}

	newSession, err := session.NewSessionWithOptions(opts)
	if err != nil {
		panic(err)
	}

	maxMessages := int64(10)
	waitTime := int64(0)
	messageAttributeNames := aws.StringSlice([]string{"All"})

	rMin := &sqs.ReceiveMessageInput{
		QueueUrl:              src,
		MaxNumberOfMessages:   &maxMessages,
		WaitTimeSeconds:       &waitTime,
		MessageAttributeNames: messageAttributeNames,
	}

	if *numClients > 1 {
		group := parallelizer.NewGroup(parallelizer.WithPoolSize(*numClients))
		defer group.Close()
		for i := 1; i <= *numClients; i++ {
			group.Add(func() {
				transferMessages(newSession, rMin, dest)
			})
		}
		err = group.Wait()
	} else {
		transferMessages(newSession, rMin, dest)
	}

	log.Println("all done")
	if err != nil {
		log.Printf("error: %v", err)
	}
}

//transferMessages loops, transferring a number of messages from the src to the dest at an interval.
func transferMessages(theSession *session.Session, rMin *sqs.ReceiveMessageInput, dest *string) {
	client := sqs.New(theSession)

	lastMessageCount := int(1)
	// loop as long as there are messages on the queue
	for {
		resp, err := client.ReceiveMessage(rMin)

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
					QueueUrl:      rMin.QueueUrl,
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
