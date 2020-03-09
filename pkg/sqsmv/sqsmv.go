package sqsmv

import (
	"fmt"
	"k8s.io/klog"
	"sync"

	"github.com/aws/aws-sdk-go/service/sqs"
)

func Run(queues []QueueConfig, stopCh <-chan struct{}) {
	klog.Info("Starting sqsmv")

	for cnt, queue := range queues {
		go sqsSync(
			fmt.Sprintf("queue-%d", cnt+1),
			queue.Source,
			queue.Destination,
			stopCh,
		)
	}

	<-stopCh
	klog.Info("Shutting down sqsmv")
}

func sqsSync(id string, srcQueue string, destQueue string, stopCh <-chan struct{}) {
	klog.Infof(
		"%s | sqsSync starting from src: %s => dest: %s",
		id,
		srcQueue,
		destQueue,
	)

	for {
		select {
		case <-stopCh:
			klog.Infof("%s | sqsSync is shutting down gracefully.", id)
			return
		default:
			sqsMv(id, srcQueue, destQueue)
		}
	}
}

func sqsMv(id string, srcQueueURL string, destQueueURL string) {
	srcQueue := NewSQSQueue(srcQueueURL)
	srcQueueDetails, err := srcQueue.Describe()
	if err != nil {
		if err != ErrQueueDoesNotExist {
			klog.Fatalf("error getting sqs queue, queue: %v, err: %v", srcQueue.url, err)
		}
		klog.Errorf("%d | source queue does not exist, queue: %v, err: %v", id, srcQueue.url, err)
		return
	}

	destQueue := NewSQSQueue(destQueueURL)
	destQueueDetails, err := destQueue.Describe()
	if err != nil {
		if err != ErrQueueDoesNotExist {
			klog.Fatalf("error getting sqs queue, queue: %v, err: %v", destQueue.url, err)
		}
		klog.Infof("%s | destination queue does not exist, queue: %v", id, destQueue.url)
		klog.Infof("%s | creating destination queue", id)
		destQueueDetails = srcQueueDetails
		destQueueDetails.QueueURL = destQueue.url
		_, err = destQueue.Create(destQueueDetails)
		if err != nil {
			klog.Fatalf("%s | error creating destination queue, err: %v", id, err)
		}
		klog.Infof("%s | created destination queue", id)
	}

	klog.Infof("%s | longPolling", id)
	messages, err := receiveMessage(srcQueue.url, 20)
	if err != nil {
		klog.Fatalf("%s | error receiving messages, err: %v", id, err)
	}
	if len(messages) == 0 {
		klog.Infof("%s | nothing in the queue", id)
		return
	}

	klog.Infof("%s | moving %v messages", id, len(messages))

	var wg sync.WaitGroup
	wg.Add(len(messages))

	for _, m := range messages {
		go func(id string, m *sqs.Message) {
			defer wg.Done()
			// write message to destination
			err := writeMessage(m, destQueue.url)
			if err != nil {
				klog.Errorf("%s | error writing message to destination, err: %v", id, err)
				return
			}

			// delete message from source which was just written
			err = deleteMessage(m, srcQueue.url)
			if err != nil {
				klog.Fatalf(
					"%s | error deleting message id %v, err: %v",
					id,
					*m.ReceiptHandle,
					err)
			}
		}(id, m)
	}

	// wait for all jobs from this batch
	wg.Wait()
}
