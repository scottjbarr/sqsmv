# sqsmv

Move all messages from one SQS queue to another.

## Installation

### Source

    go get github.com/vercel/sqsmv

## Configuration

The `AWS_SECRET_ACCESS_KEY`, `AWS_ACCESS_KEY_ID`, and `AWS_REGION`
environment variables must be set.

## Usage

Supply source and destination URL endpoints.

    sqsmv -src https://region.queue.amazonaws.com/123/queue-a -dest https://region.queue.amazonaws.com/123/queue-b

## License

The MIT License (MIT)

Copyright (c) 2016-2018 Scott Barr

See [LICENSE.md](LICENSE.md)
