package sqsmv

import (
	"sync"

	"github.com/aws/aws-sdk-go/service/sqs"
	"k8s.io/klog"
)

func Run(stopCh <-chan struct{}) {
	klog.Info("starting sqsmv")
	go sqsSync(1987, "", "", stopCh)
	<-stopCh
	klog.Info("shutting down sqsmv")
}

func sqsSync(id int, srcQueue string, destQueue string, stopCh <-chan struct{}) {
	longPollCh := make(chan int32)
	longPollResumeCh := make(chan int)
	go longPoll(id, srcQueue, longPollCh, longPollResumeCh, stopCh)

	klog.Infof(
		"%d | sqsSync started src: %s, dest: %s",
		id,
		srcQueue,
		destQueue,
	)
	for {
		select {
		case <-longPollCh:
			klog.Infof("%d | sqsSync triggering sqsMv as long poll found messages", id)
			sqsMv(id, srcQueue, destQueue)
			klog.Infof("%d | sqsSync sees sqsMv has finished operation", id)
			longPollResumeCh <- 0
		case <-stopCh:
			klog.Infof("%d | sqsSync shutting down gracefully.", id)
			return
		}
	}

}

func longPoll(
	id int, queue string, longPollCh chan<- int32,
	longPollResumeCh <-chan int,
	stopCh <-chan struct{}) {
	klog.Infof("%d | longPoll started", id)

	for {
		select {
		case <-stopCh:
			klog.Infof("%d | longPoll stopping", id)
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
			klog.Infof("%d | longPoll has found messages in queue", id)
			longPollCh <- messages

			// longPolling should sleep until told to start again
			klog.Infof("%d | longPoll waiting to be started again", id)
			<-longPollResumeCh
			klog.Infof("%d | longPoll started again", id)
		}
	}
}

func sqsMv(id int, srcQueue string, destQueue string) {
	messages, err := receiveMessage(srcQueue)
	if err != nil {
		klog.Fatalf("%d | error receiving messages, err: %v", id, err)
	}

	if len(messages) == 0 {
		klog.Fatalf("%d | messages received should not be 0 since long-poll had found messages. Investigate!", id)
	}
	klog.Infof("%d | sqsMv is operating on %v messages", id, len(messages))

	var wg sync.WaitGroup
	wg.Add(len(messages))

	for _, m := range messages {
		go func(id int, m *sqs.Message) {
			defer wg.Done()
			// write message to destination
			err := writeMessage(m, srcQueue)
			if err != nil {
				klog.Errorf("%d | error writing message to destination, err: %v", id, err)
			}

			// delete message from source which was just written
			err = deleteMessage(m, destQueue)
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
