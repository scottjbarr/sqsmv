package sqsmv

import (
	"sync"

	"github.com/aws/aws-sdk-go/service/sqs"
	"k8s.io/klog"
)

func Run(stopCh <-chan struct{}) {
	klog.Info("Starting sqsmv")
	go sqsSync(1987, "https://sqs.ap-southeast-1.amazonaws.com/754922593538/practodevsingapore-accounts-loginotp-latest",
		"https://sqs.ap-south-1.amazonaws.com/754922593538/practodevmumbai-accounts-loginotp-latest", stopCh)

	<-stopCh
	klog.Info("Shutting down sqsmv")
}

func sqsSync(id int, srcQueue string, destQueue string, stopCh <-chan struct{}) {
	longPollCh := make(chan int32)
	longPollResumeCh := make(chan int)
	go longPoll(id, srcQueue, longPollCh, longPollResumeCh, stopCh)

	klog.Infof(
		"%d | sqsSync starting from src: %s => dest: %s",
		id,
		srcQueue,
		destQueue,
	)
	for {
		select {
		case <-longPollCh:
			klog.Infof("%d | sqsSync is triggering sqsMv", id)
			sqsMv(id, srcQueue, destQueue)
			klog.Infof("%d | sqsMv is done processing", id)
			longPollResumeCh <- 0
		case <-stopCh:
			klog.Infof("%d | sqsSync is shutting down gracefully.", id)
			return
		}
	}

}

func longPoll(
	id int, queue string, longPollCh chan<- int32,
	longPollResumeCh <-chan int,
	stopCh <-chan struct{}) {
	klog.Infof("%d | longPolling has started", id)

	for {
		select {
		case <-stopCh:
			klog.Infof("%d | longPolling is shutting down gracefully.", id)
			return
		default:
			messages, err := longPollReceiveMessage(queue)
			if err != nil {
				klog.Fatalf("%d | error in longPolling, err: %v", id, err)
			}
			if messages == 0 {
				continue
			}
			// trigger sqsmv to start moving messages
			klog.Infof("%d | longPolling found messages in queue", id)
			longPollCh <- messages

			// longPolling should sleep until told to start again
			klog.Infof("%d | longPolling is sleeping", id)
			<-longPollResumeCh
			klog.Infof("%d | longPolling has started again", id)
		}
	}
}

func sqsMv(id int, srcQueue string, destQueue string) {
	messages, err := receiveMessage(srcQueue)
	if err != nil {
		klog.Fatalf("%d | error receiving messages, err: %v", id, err)
	}

	if len(messages) == 0 {
		klog.Fatalf("%d | messages received should not be 0 since longPolling had found messages. Investigate!", id)
	}
	klog.Infof("%d | sqsMv is operating on %v messages", id, len(messages))

	var wg sync.WaitGroup
	wg.Add(len(messages))

	for _, m := range messages {
		go func(id int, m *sqs.Message) {
			defer wg.Done()
			// write message to destination
			err := writeMessage(m, destQueue)
			if err != nil {
				klog.Errorf("%d | error writing message to destination, err: %v", id, err)
			}

			// delete message from source which was just written
			err = deleteMessage(m, srcQueue)
			if err != nil {
				klog.Fatalf(
					"%d | error deleting message id %v, err: %v",
					id,
					*m.ReceiptHandle,
					err)
			}
		}(id, m)
	}

	// wait for all jobs from this batch
	wg.Wait()
}
