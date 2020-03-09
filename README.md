# sqsmv

Keep moving messages between a list of two SQS queues.

## Installation

- Create the env.sh and source it
```
cp .envs.sh.sample .envs.sh
```

- Update AWS secrets in `.envs.sh` and source it
```
source .envs.sh
```

- Create the config map. Update the queues manually for the first time.
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: sqsmv
  namespace: central
data:
  sqsmv.yaml: |
    queues:
    - source: https://ap-southeast-1.queue.amazonaws.com/123/wat-a
      destination: https://ap-south-1.queue.amazonaws.com/123/wat-a
```
```
kubectl create -f ./artifacts/configmap-example.yaml
```

- Install `sqsmv` in your Kubernetes cluster.
```
make deploy
```

- Check the logs
```
kubectl get pods -n central | grep sqsmv | awk '{print $1}' | xargs -I {} kubectl logs -f {} -n central
```

- **Reloading** `sqsmv` with the new set of queues. Just edit config map and delete the pod. We are updating the `sqsmv` queues list by updating the config map by an external process and re-creating the sqsmv pod. This helps us in having a downtime free sqs migration across region.
```
kubectl edit cm -n central sqsmv
make deploy
```

## Seeing is believing :)
```
for i in {0..24..1}; do
    aws sqs send-message \
        --queue-url https://ap-southeast-1.queue.amazonaws.com/123/queue-1
        --message-body "{\"id\": $i}"
done
```

```
$ bin/darwin_amd64/sqsmv run
I0309 18:51:02.231771   66861 config.go:26] Using config file: /etc/config/sqsmv.yaml
I0309 18:51:02.232248   66861 sqsmv.go:12] Starting sqsmv
I0309 18:51:02.232300   66861 sqsmv.go:28] queue-1 | sqsSync starting from src: https://sqs.ap-southeast-1.amazonaws.com/123/queue-1 => dest: https://sqs.ap-south-1.amazonaws.com/123/queue-1
I0309 18:51:02.939755   66861 sqsmv.go:63] queue-1 | destination queue does not exist, queue: https://sqs.ap-south-1.amazonaws.com/123/queue-1
I0309 18:51:02.939785   66861 sqsmv.go:64] queue-1 | creating destination queue
I0309 18:51:03.020776   66861 sqsmv.go:71] queue-1 | created destination queue
I0309 18:51:03.020823   66861 sqsmv.go:74] queue-1 | longPolling
I0309 18:51:13.679031   66861 sqsmv.go:84] queue-1 | moving 1 messages
I0309 18:51:13.953909   66861 sqsmv.go:74] queue-1 | longPolling
I0309 18:51:25.261180   66861 sqsmv.go:84] queue-1 | moving 1 messages
I0309 18:51:25.541479   66861 sqsmv.go:74] queue-1 | longPolling
I0309 18:51:27.705795   66861 sqsmv.go:84] queue-1 | moving 1 messages
I0309 18:51:28.013757   66861 sqsmv.go:74] queue-1 | longPolling
I0309 18:51:33.846544   66861 sqsmv.go:84] queue-1 | moving 1 messages
I0309 18:51:34.129340   66861 sqsmv.go:74] queue-1 | longPolling
```
**Note:** Logs are shown with ids of the queue (`queue-1` is the `id` this example). So that it easier to debug(using `grep`) each queue source-destination configuration.

## Contributing
```
$ make build
making bin/darwin_amd64/sqsmv

$ sudo bin/darwin_amd64/sqsmv run
I0305 19:39:30.652967   89271 config.go:42] Using config file: /etc/config/sqsmv.yaml
I0305 17:32:50.128648   82452 sqsmv.go:8] Starting sqsmv

$ make push
$ make deploy
```
**Note:** For PRs, `git push` and not `make push` :)

## Troubleshoot

Find the queue id you want to troubleshoot and grep for that queue_id
```
kubectl logs -n central sqsmv-xxx |  grep queue_name | grep "sqsSync starting from src"
kubectl logs -n central sqsmv-xx | grep queue_id
```


## License

The MIT License (MIT)

Copyright (c) 2016-2018 Scott Barr

See [LICENSE.md](LICENSE.md)

## Thanks

https://github.com/scottjbarr/sqsmv
