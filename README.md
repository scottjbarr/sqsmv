# sqsmv

Continuously sync messages between SQS queues.

## Installation

- Create the env.sh and source it
```
cp .envs.sh.sample .envs.sh
# vim .envs.sh (and add values)
source .env.sh
```

- Create the first time configuration. Later you can modify the configmap from some external process or manually.
```
export KUBE_CONTEXT=k8sx.domain
make create-firstime-config
```

- Install `sqsmv` in your Kubernetes cluster.
```
make deploy
```

- Check the logs
```
kubectl get pods -n central | grep sqsmv | awk '{print $1}' | xargs -I {} kubectl logs -f {} -n central
```

## Seeing is believing :)
```
for i in {0..24..1}; do
    aws sqs send-message \
        --queue-url https://ap-south-1.queue.amazonaws.com/123/wat-a
        --message-body "{\"id\": $i}"
done
```

```
$ bin/darwin_amd64/sqsmv run
I0306 22:52:12.696796   23706 config.go:40] Using config file: /etc/config/sqsmv.yaml
I0306 22:52:12.702603   23706 sqsSync.go:11] Starting sqsmv
I0306 22:52:12.703253   23706 sqsSync.go:49] uuid_1 | longPolling has started
I0306 22:52:12.703275   23706 sqsSync.go:24] uuid_1 | sqsSync starting from src: https://ap-south-1.queue.amazonaws.com/123/wat-a => dest: https://ap-southeast-1.queue.amazonaws.com/123/wat-a
I0306 22:52:24.731105   23706 sqsSync.go:65] uuid_1 | longPolling found messages in queue
I0306 22:52:24.731136   23706 sqsSync.go:69] uuid_1 | longPolling is sleeping
I0306 22:52:24.731150   23706 sqsSync.go:33] uuid_1 | sqsSync is triggering sqsMv
I0306 22:52:24.786064   23706 sqsSync.go:85] uuid_1 | sqsMv is operating on 10 messages
I0306 22:52:24.998162   23706 sqsSync.go:35] uuid_1 | sqsMv is done processing
I0306 22:52:24.998189   23706 sqsSync.go:71] uuid_1 | longPolling has started again
```
**Note:** Logs are shown with `uuid` so that it easier to debug what is happening with each queue pair configuration.

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

## License

The MIT License (MIT)

Copyright (c) 2016-2018 Scott Barr

See [LICENSE.md](LICENSE.md)

## Thanks

https://github.com/scottjbarr/sqsmv
