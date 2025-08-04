package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"wb-l0/internal/config"
	"wb-l0/internal/domain"

	"github.com/IBM/sarama"

	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

//go:generate mockgen -source=consumer.go -destination=mocks/mock.go
//go:generate mockgen -source=internal/kafka/consumer.go -destination=internal/kafka/mocks/mock.go  KafkaConsumerInterface,DB
//go:generate mockgen -destination=internal/kafka/mocks/sarama_mock.go -package=mocks github.com/IBM/sarama Consumer,PartitionConsumer

var (
	processedMessagesCounter prometheus.Counter
	unmarshalErrorsCounter   prometheus.Counter
	processingErrorsCounter  prometheus.Counter
	retriesCounter           prometheus.Counter
)

func init() {
	processedMessagesCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kafka_consumer_processed_messages_total",
		Help: "Total number of processed Kafka messages",
	})

	unmarshalErrorsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kafka_consumer_unmarshal_errors_total",
		Help: "Total number of Kafka messages that failed to unmarshal",
	})

	processingErrorsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kafka_consumer_processing_errors_total",
		Help: "Total number of Kafka messages that failed to process",
	})

	retriesCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kafka_consumer_retries_total",
		Help: "Total number of Kafka message processing retries",
	})
}

type DB interface {
	CreateOrder(ctx context.Context, order domain.Order) (int, error)
}

type KafkaConsumerInterface interface {
	Consume(ctx context.Context) error
	Close() error
}

type KafkaConsumer struct {
	topic                         string
	cfg                           config.Config
	logger                        *slog.Logger
	wg                            sync.WaitGroup
	orderService                  DB
	consumer                      sarama.Consumer
	treatUnmarshalErrorAsCritical bool
	validator                     *validator.Validate

	processedMessagesCounter prometheus.Counter
	unmarshalErrorsCounter   prometheus.Counter
	processingErrorsCounter  prometheus.Counter
	retriesCounter           prometheus.Counter

	errChan chan error
}

func NewKafkaConsumer(ctx context.Context, cfg config.Config, logger *slog.Logger, saramaConfig *sarama.Config, consumer sarama.Consumer, orderService DB) (*KafkaConsumer, error) {
	validator := validator.New()
	errChan := make(chan error, 10)
	return &KafkaConsumer{
		topic:                         cfg.Kafka.Topic,
		cfg:                           cfg,
		logger:                        logger,
		orderService:                  orderService,
		consumer:                      consumer,
		errChan:                       errChan,
		treatUnmarshalErrorAsCritical: cfg.Kafka.TreatUnmarshalErrorAsCritical,
		processedMessagesCounter:      processedMessagesCounter,
		unmarshalErrorsCounter:        unmarshalErrorsCounter,
		processingErrorsCounter:       processingErrorsCounter,
		retriesCounter:                retriesCounter,
		validator:                     validator,
	}, nil
}

func (s *KafkaConsumer) Consume(ctx context.Context) error {
	partitions, err := s.consumer.Partitions(s.topic)
	if err != nil {
		s.logger.Error("failed to get partitions", "error", err)
		return err
	}

	var mu sync.Mutex
	var allErrors []error

	for _, partition := range partitions {
		pc, err := s.consumer.ConsumePartition(s.topic, partition, sarama.OffsetNewest)
		if err != nil {
			s.logger.Error("failed to consume partition", "partition", partition, "error", err)
			mu.Lock()
			allErrors = append(allErrors, err)
			mu.Unlock()
			continue
		}
		if pc == nil {
			s.logger.Error("partition consumer is nil", "partition", partition)
			continue
		}

		s.wg.Add(1)

		go func(pc sarama.PartitionConsumer, partition int32) {
			defer s.wg.Done()
			defer func() {
				if err := pc.Close(); err != nil {
					s.logger.Error("failed to close partition consumer", "partition", partition, "error", err)
					mu.Lock()
					allErrors = append(allErrors, err)
					mu.Unlock()
				}
			}()

			for {
				select {
				case msg, ok := <-pc.Messages():
					if !ok {
						s.logger.Info("message channel closed", "partition", partition)
						return
					}

					s.logger.Info("received message", "partition", msg.Partition, "offset", msg.Offset)

					var order domain.Order
					if err := json.Unmarshal(msg.Value, &order); err != nil {
						s.logger.Error("failed to unmarshal message", "error", err)
						s.unmarshalErrorsCounter.Inc()
						if s.treatUnmarshalErrorAsCritical {
							select {
							case s.errChan <- fmt.Errorf("failed to unmarshal message at offset %d: %w", msg.Offset, err):
							case <-ctx.Done():
								return
							}
							return
						}
						continue
					}

					if err := s.validateOrder(order); err != nil {
						s.logger.Error("kafka.consumer.Consumer: invalid order", slog.String("error", err.Error()))
						continue
					}

					var processErr error
					for attempt := 0; attempt <= s.cfg.Kafka.MaxRetries; attempt++ {
						if attempt > 0 {
							s.retriesCounter.Inc()
						}
						processErr = s.processOrder(ctx, order)
						if processErr == nil {
							s.processedMessagesCounter.Inc()
							break
						}
						if attempt < s.cfg.Kafka.MaxRetries {
							s.logger.Warn("attempt to process the order failed",
								slog.Int("attempt", attempt),
								slog.Any("partition", partition),
								slog.String("error", processErr.Error()),
							)
							backOff := s.cfg.Kafka.InitialBackoff * time.Duration(1<<attempt)
							select {
							case <-time.After(backOff):
								continue
							case <-ctx.Done():
								s.logger.Info("shutting down kafka consumer due to context cancellation",
									slog.Any("partition", partition))
								return
							}
						}

						s.logger.Error("failed to process order after max retries",
							slog.String("error", processErr.Error()),
							slog.Any("partition", partition))

						s.processingErrorsCounter.Inc()
					}

					if processErr != nil {
						mu.Lock()
						allErrors = append(allErrors, processErr)
						mu.Unlock()
					}

				case err, ok := <-pc.Errors():
					if !ok {
						s.logger.Info("error channel closed", "partition", partition)
						return
					}
					s.logger.Error("partition consumer error", "error", err.Err)
					mu.Lock()
					allErrors = append(allErrors, err.Err)
					mu.Unlock()
					select {
					case s.errChan <- fmt.Errorf("partition consumer error: %w", err.Err):
					case <-ctx.Done():
						return
					}
				case <-ctx.Done():
					s.logger.Info("context canceled, shutting down partition consumer", "partition", partition)
					return
				}
			}
		}(pc, partition)
	}

	s.wg.Wait()
	close(s.errChan)

	var firstErr error
	select {
	case err := <-s.errChan:
		firstErr = err
	default:
		firstErr = nil
	}

	if len(allErrors) > 0 {
		aggErr := errors.Join(allErrors...)
		if firstErr != nil {
			return errors.Join(firstErr, aggErr)
		}
		return aggErr
	}

	if errors.Is(ctx.Err(), context.Canceled) {
		s.logger.Info("consumer context canceled, all goroutines finished")
		return ctx.Err()
	}

	return firstErr
}

func (s *KafkaConsumer) processOrder(ctx context.Context, order domain.Order) error {
	_, err := s.orderService.CreateOrder(ctx, order)
	if err != nil {
		s.logger.Error("failed to create order", "error", err, "orderUID", order.OrderUID)
		return fmt.Errorf("failed to create order: %w", err)
	}
	return nil
}

func (s *KafkaConsumer) Close() error {
	if err := s.consumer.Close(); err != nil {
		return fmt.Errorf("failed to close kafka consumer: %w", err)
	}
	return nil
}

func (s *KafkaConsumer) GetError() error {
	select {
	case err := <-s.errChan:
		return err
	default:
		return nil
	}
}

func (s *KafkaConsumer) validateOrder(order domain.Order) error {
	err := s.validator.Struct(order)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			s.logger.Error("validation error",
				"field", err.Field(),
				"tag", err.Tag(),
				"value", err.Value(),
			)
		}
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}
