# sqsmv

Keep moving messages between a list of two SQS queues.

## Installation

- Create the env.sh and source it
```
cp .envs.sh.sample .envs.sh
```

- Update AWS secrets in `.envs.sh` and source it
```
source .env.sh
```

- Create the config map
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
I0307 01:01:46.620534   33099 config.go:40] Using config file: /etc/config/sqsmv.yaml
I0307 01:01:46.621059   33099 sqsmv.go:12] Starting sqsmv
I0307 01:01:46.621108   33099 sqsmv.go:28] queue-1 | sqsSync starting from src: https://sqs.ap-southeast-1.amazonaws.com/123/queue-1 => dest: https://sqs.ap-south-1.amazonaws.com/123/queue-1
I0307 01:01:46.621170   33099 sqsmv.go:58] queue-1 | longPolling has started
I0307 01:04:11.472369   33099 sqsmv.go:74] queue-1 | longPolling found messages in queue
I0307 01:04:11.472627   33099 sqsmv.go:78] queue-1 | longPolling is sleeping
I0307 01:04:11.472657   33099 sqsmv.go:42] queue-1 | sqsSync is triggering sqsMv
I0307 01:04:11.540273   33099 sqsmv.go:94] queue-1 | sqsMv is operating on 1 messages
I0307 01:04:11.734887   33099 sqsmv.go:44] queue-1 | sqsMv is done processing
I0307 01:04:11.734941   33099 sqsmv.go:80] queue-1 | longPolling has started again
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
