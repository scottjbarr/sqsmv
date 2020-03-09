package sqsmv

import (
	"fmt"
	"k8s.io/klog"
	"sync"

	"github.com/aws/aws-sdk-go/service/sqs"
)

func Run(queues []Queue, stopCh <-chan struct{}) {
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

func sqsMv(id string, srcQueue string, destQueue string) {
	klog.Infof("%s | longPolling", id)
	messages, err := receiveMessage(srcQueue, 20)
	if err != nil {
		klog.Fatalf("%s | error receiving messages, err: %v", id, err)
	}
	if len(messages) == 0 {
		klog.Infof("%s | nothing in the queue", id)
		return
	}

	klog.Infof("%s | operating on %v messages", id, len(messages))

	var wg sync.WaitGroup
	wg.Add(len(messages))

	for _, m := range messages {
		go func(id string, m *sqs.Message) {
			defer wg.Done()
			// write message to destination
			err := writeMessage(m, destQueue)
			if err != nil {
				klog.Errorf("%s | error writing message to destination, err: %v", id, err)
				return
			}

			// delete message from source which was just written
			err = deleteMessage(m, srcQueue)
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
