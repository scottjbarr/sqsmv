# sqsmv

Move all messages from one SQS queue, to another.


## Installation

### Source

    go get github.com/scottjbarr/sqsmv


### Binaries

Download the appropriate binary from the
[Releases](https://github.com/scottjbarr/sqsmv/releases) page.


## Configuration

The `AWS_SECRET_ACCESS_KEY`, `AWS_ACCESS_KEY_ID`, and ,`AWS_REGION`
environment variables must be set.


## Usage

Supply source and destination URL endpoints.

    sqsmv -src https://region.queue.amazonaws.com/123/queue-a -dest https://region.queue.amazonaws.com/123/queue-b

Optionally supply also the awscli [named profile](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html) to use.

    sqsmv -src https://region.queue.amazonaws.com/123/queue-a -dest https://region.queue.amazonaws.com/123/queue-b -aws-profile myprofile

## Seeing is believing :)

Create some SQS messages to play with using the AWS CLI.

    for i in {0..24..1}; do
        aws sqs send-message \
            --queue-url https://ap-southeast-2.queue.amazonaws.com/123/wat-a
            --message-body "{\"id\": $i}"
    done


## License

The MIT License (MIT)

Copyright (c) 2016-2018 Scott Barr

See [LICENSE.md](LICENSE.md)
