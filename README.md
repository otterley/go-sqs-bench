SQS benchmark tool
==================

This is a simple Amazon SQS publish benchmarking tool, written in Go.

Environment variables
---------------------
* `AWS_REGION` - AWS region to use
* `QUEUE_NAME` - SQS queue name, will be autocreated if necessary
* `CONCURRENCY` - Number of concurrent workers (defaults to `GOMAXPROCS`)
* `RATE` - Frequency at which to publish messages to SQS, in Go time.Duration
  format (default: as fast as possible)

Building
--------
```sh
go build
```

License
-------
MIT
