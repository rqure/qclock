package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	qmq "github.com/rqure/qmq/src"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NameProvider struct{}

func (np *NameProvider) Get() string {
	return "clock"
}

type TimestampToAnyTransformer struct {
	logger qmq.Logger
}

func NewTimestampToAnyTransformer(logger qmq.Logger) qmq.Transformer {
	return &TimestampToAnyTransformer{
		logger: logger,
	}
}

func (t *TimestampToAnyTransformer) Transform(i interface{}) interface{} {
	d, ok := i.(*qmq.Timestamp)
	if !ok {
		t.logger.Error(fmt.Sprintf("TimestampToAnyTransformer: invalid input %T", i))
		return nil
	}

	a, err := anypb.New(d)
	if err != nil {
		t.logger.Error(fmt.Sprintf("TimestampToAnyTransformer: failed to marshal timestamp into anypb: %v", err))
		return nil
	}

	return a
}

type TransformerProviderFactory struct{}

func (t *TransformerProviderFactory) Create(components qmq.EngineComponentProvider) qmq.TransformerProvider {
	transformerProvider := qmq.NewDefaultTransformerProvider()
	transformerProvider.Set("producer:clock:event:new-timestamp", []qmq.Transformer{
		NewTimestampToAnyTransformer(components.WithLogger()),
		qmq.NewAnyToMessageTransformer(components.WithLogger(), qmq.AnyToMessageTransformerConfig{
			SourceProvider: &NameProvider{},
		}),
		qmq.NewTracePushTransformer(components.WithLogger()),
	})
	return transformerProvider
}

type ClockEngineProcessor struct{}

func (c *ClockEngineProcessor) Process(e qmq.EngineComponentProvider) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(500 * time.Millisecond)

	for {
		select {
		case <-quit:
			return
		case <-ticker.C:
			timestamp := &qmq.Timestamp{Value: timestamppb.Now()}
			e.WithProducer("clock:event:new-timestamp").Push(timestamp)
		}
	}
}

func main() {
	engine := qmq.NewDefaultEngine(qmq.DefaultEngineConfig{
		NameProvider:               &NameProvider{},
		TransformerProviderFactory: &TransformerProviderFactory{},
		EngineProcessor:            &ClockEngineProcessor{},
	})
	engine.Run()
}
