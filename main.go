package main

import (
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/otterley/go-sqs-bench/internal"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
}

func main() {
	queue := os.Getenv("QUEUE_NAME")
	if queue == "" {
		log.Fatal().Msg("QUEUE_NAME not set")
		os.Exit(1)
	}

	concurrency := int64(runtime.GOMAXPROCS(0))
	concStr := os.Getenv("CONCURRENCY")
	if concStr != "" {
		var err error
		concurrency, err = strconv.ParseInt(concStr, 0, 64)
		if err != nil {
			log.Fatal().Err(err).Msg("")
			os.Exit(1)
		}
	}

	var rate time.Duration
	rateStr := os.Getenv("RATE")
	if rateStr != "" {
		var err error
		rate, err = time.ParseDuration(rateStr)
		if err != nil {
			log.Fatal().Err(err).Msg("")
			os.Exit(1)
		}
	}

	sess := session.Must(session.NewSession())
	client := sqs.New(sess)

	resp, err := client.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String(queue),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("")
		os.Exit(1)
	}

	c := internal.Config{
		Client:      client,
		QueueURL:    aws.StringValue(resp.QueueUrl),
		Concurrency: int(concurrency),
		StopCh:      make(chan os.Signal, 1),
		Rate:        rate,
	}
	signal.Notify(c.StopCh, syscall.SIGINT, syscall.SIGTERM)

	internal.StartParallelPublishLoop(&c)
}
