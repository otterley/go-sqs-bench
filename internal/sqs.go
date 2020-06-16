package internal

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Client      *sqs.SQS
	QueueURL    string
	StopCh      chan os.Signal
	Concurrency int
	Rate        time.Duration

	logger zerolog.Logger
}

func StartParallelPublishLoop(c *Config) {
	var errs []chan error
	var stopChs []chan os.Signal

	for i := 0; i < c.Concurrency; i++ {
		errCh := make(chan error, 1)
		stopCh := make(chan os.Signal, 1)
		workerCfg := *c
		workerCfg.StopCh = stopCh
		workerCfg.logger = log.With().Int("WorkerID", i).Logger()
		go func() {
			errCh <- startPublishLoop(&workerCfg)
		}()
		errs = append(errs, errCh)
		stopChs = append(stopChs, stopCh)
	}
	// Deliver stop signal to all workers when received
	go func() {
		sig := <-c.StopCh
		for _, ch := range stopChs {
			ch <- sig
		}
	}()
	for _, ch := range errs {
		<-ch
	}
}

func startPublishLoop(c *Config) error {
	c.logger.Info().Msg("Starting publish loop with duration " + c.Rate.String())

	if c.Rate == time.Duration(0) {
		for {
			err := publish(c)
			if err != nil {
				c.logger.Error().Err(err).Msg("")
				return err
			}
			select {
			case <-c.StopCh:
				c.logger.Info().Msg("received stop request")
				return nil
			default:
			}
		}
	}

	tick := time.Tick(c.Rate)
	for {
		select {
		case <-tick:
			err := publish(c)
			if err != nil {
				c.logger.Error().Err(err).Msg("")
				return err
			}
		case <-c.StopCh:
			c.logger.Info().Msg("received stop request")
			return nil
		}
	}
}

func publish(c *Config) error {
	resp, err := c.Client.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    aws.String(c.QueueURL),
		MessageBody: aws.String(time.Now().Format(time.RFC3339)),
	})
	if err != nil {
		return err
	}
	c.logger.Info().
		Str("MessageID", aws.StringValue(resp.MessageId)).Msg("")

	return nil
}
