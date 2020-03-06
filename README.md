# sqsmv

Move all the messages from one SQS queue, to another. This sqsmv fork, runs as a daemon. It has all the automation to run on Kubernetes as a worker Deployment.

This worker operates on a Configuration (config map)[./artifcats/configmap-template.yaml]. You can keep updating the configuration from an external process and delete the sqsmv pod. The new pod will then start moving the messages based on the new configuration.

It runs a separate go routine for every item specified in the Configuration.

We are using this service to have a downtime free migration of SQS across region for multiple products across the organization.

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

## License

The MIT License (MIT)

Copyright (c) 2016-2018 Scott Barr

See [LICENSE.md](LICENSE.md)

## Thanks

https://github.com/scottjbarr/sqsmv
