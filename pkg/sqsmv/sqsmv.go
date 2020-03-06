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

	longPollCh := make(chan int32)
	longPollResumeCh := make(chan int)
	go longPoll(id, srcQueue, longPollCh, longPollResumeCh, stopCh)

	for {
		select {
		case <-longPollCh:
			klog.Infof("%s | sqsSync is triggering sqsMv", id)
			sqsMv(id, srcQueue, destQueue)
			klog.Infof("%s | sqsMv is done processing", id)
			longPollResumeCh <- 0
		case <-stopCh:
			klog.Infof("%s | sqsSync is shutting down gracefully.", id)
			return
		}
	}

}

func longPoll(
	id string, queue string, longPollCh chan<- int32,
	longPollResumeCh <-chan int,
	stopCh <-chan struct{}) {
	klog.Infof("%s | longPolling has started", id)

	for {
		select {
		case <-stopCh:
			klog.Infof("%s | longPolling is shutting down gracefully.", id)
			return
		default:
			messages, err := longPollReceiveMessage(queue)
			if err != nil {
				klog.Fatalf("%s | error in longPolling, err: %v", id, err)
			}
			if messages == 0 {
				continue
			}
			// trigger sqsmv to start moving messages
			klog.Infof("%s | longPolling found messages in queue", id)
			longPollCh <- messages

			// longPolling should sleep until told to start again
			klog.Infof("%s | longPolling is sleeping", id)
			<-longPollResumeCh
			klog.Infof("%s | longPolling has started again", id)
		}
	}
}

func sqsMv(id string, srcQueue string, destQueue string) {
	messages, err := receiveMessage(srcQueue)
	if err != nil {
		klog.Fatalf("%s | error receiving messages, err: %v", id, err)
	}

	if len(messages) == 0 {
		klog.Fatalf("%s | messages received should not be 0 since longPolling had found messages. Investigate!", id)
	}
	klog.Infof("%s | sqsMv is operating on %v messages", id, len(messages))

	var wg sync.WaitGroup
	wg.Add(len(messages))

	for _, m := range messages {
		go func(id string, m *sqs.Message) {
			defer wg.Done()
			// write message to destination
			err := writeMessage(m, destQueue)
			if err != nil {
				klog.Errorf("%s | error writing message to destination, err: %v", id, err)
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
