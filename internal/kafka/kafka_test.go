package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"
	"wb-l0/internal/config"
	"wb-l0/tests"

	mock_kafka "wb-l0/internal/kafka/mocks"

	"log/slog"

	"github.com/IBM/sarama"
	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/assert"
)

func SetupTest(t *testing.T) (*KafkaConsumer, *mock_kafka.MockDB, *mock_kafka.MockConsumer, *mock_kafka.MockPartitionConsumer, *gomock.Controller) {
	ctrl := gomock.NewController(t)

	mockDB := mock_kafka.NewMockDB(ctrl)
	mockConsumer := mock_kafka.NewMockConsumer(ctrl)
	mockPartitionConsumer := mock_kafka.NewMockPartitionConsumer(ctrl)

	cfg := config.Config{
		Kafka: config.KafkaConfig{
			Topic:          "test-topic",
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	kafkaConsumer, err := NewKafkaConsumer(context.Background(), cfg, logger, &sarama.Config{}, mockConsumer, mockDB)
	assert.NoError(t, err)

	return kafkaConsumer, mockDB, mockConsumer, mockPartitionConsumer, ctrl
}
func TestKafkaConsumer_Consume_Success(t *testing.T) {
	kafkaConsumer, mockDB, mockConsumer, mockPartitionConsumer, ctrl := SetupTest(t)
	defer ctrl.Finish()

	partitions := []int32{0}
	mockConsumer.EXPECT().Partitions(kafkaConsumer.topic).Return(partitions, nil)
	mockConsumer.EXPECT().ConsumePartition(kafkaConsumer.topic, int32(0), sarama.OffsetNewest).Return(mockPartitionConsumer, nil)

	messages := make(chan *sarama.ConsumerMessage, 1)

	mockPartitionConsumer.EXPECT().Messages().Return((<-chan *sarama.ConsumerMessage)(messages)).AnyTimes()
	mockPartitionConsumer.EXPECT().Errors().Return(nil).AnyTimes()
	mockPartitionConsumer.EXPECT().Close().Return(nil)

	testOrder := tests.InstanceStruct
	orderBytes, _ := json.Marshal(testOrder)

	if err := kafkaConsumer.validateOrder(testOrder); err != nil {
		t.Logf("Validation failed")
	}
	messages <- &sarama.ConsumerMessage{Value: orderBytes}
	close(messages)

	mockDB.EXPECT().CreateOrder(gomock.Any(), testOrder).Return(1, nil)

	err := kafkaConsumer.Consume(context.Background())

	assert.NoError(t, err)
}

func TestKafkaConsumer_Consume_PartitionError(t *testing.T) {
	kafkaConsumer, _, mockConsumer, _, ctrl := SetupTest(t)
	defer ctrl.Finish()

	partitions := []int32{0}
	mockConsumer.EXPECT().Partitions(kafkaConsumer.topic).Return(partitions, nil)
	expectedErr := errors.New("test-partition-error")
	mockConsumer.EXPECT().ConsumePartition(kafkaConsumer.topic, int32(0), sarama.OffsetNewest).Return(nil, expectedErr)

	err := kafkaConsumer.Consume(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr.Error())
}

func TestKafkaConsumer_Consume_OrderProcessingError(t *testing.T) {
	kafkaConsumer, mockDB, mockConsumer, mockPartitionConsumer, ctrl := SetupTest(t)
	defer ctrl.Finish()

	partitions := []int32{0}
	mockConsumer.EXPECT().Partitions(kafkaConsumer.topic).Return(partitions, nil)
	mockConsumer.EXPECT().ConsumePartition(kafkaConsumer.topic, int32(0), sarama.OffsetNewest).Return(mockPartitionConsumer, nil)

	messages := make(chan *sarama.ConsumerMessage, 1)
	mockPartitionConsumer.EXPECT().Messages().Return((<-chan *sarama.ConsumerMessage)(messages)).AnyTimes()
	mockPartitionConsumer.EXPECT().Errors().Return(nil).AnyTimes()
	mockPartitionConsumer.EXPECT().Close().Return(nil)

	testOrder := tests.InstanceStruct
	orderBytes, _ := json.Marshal(testOrder)
	messages <- &sarama.ConsumerMessage{Value: orderBytes}
	close(messages)

	expectedErr := errors.New("test-db-error")

	mockDB.EXPECT().CreateOrder(gomock.Any(), testOrder).Return(0, expectedErr).Times(kafkaConsumer.cfg.Kafka.MaxRetries + 1)

	err := kafkaConsumer.Consume(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr.Error())
}

func TestKafkaConsumer_Close_Success(t *testing.T) {
	kafkaConsumer, _, mockConsumer, _, ctrl := SetupTest(t)
	defer ctrl.Finish()

	mockConsumer.EXPECT().Close().Return(nil)

	err := kafkaConsumer.Close()
	assert.NoError(t, err)
}

func TestKafkaConsumer_Close_Error(t *testing.T) {
	kafkaConsumer, _, mockConsumer, _, ctrl := SetupTest(t)
	defer ctrl.Finish()

	expectedErr := errors.New("test-close-error")
	mockConsumer.EXPECT().Close().Return(expectedErr)

	err := kafkaConsumer.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr.Error())
}

func TestKafkaConsumer_GetError_NoError(t *testing.T) {
	kafkaConsumer, _, _, _, ctrl := SetupTest(t)
	defer ctrl.Finish()

	err := kafkaConsumer.GetError()
	assert.NoError(t, err)
}

func TestKafkaConsumer_GetError_Error(t *testing.T) {
	kafkaConsumer, _, _, _, ctrl := SetupTest(t)
	defer ctrl.Finish()

	expectedErr := errors.New("test-error")
	kafkaConsumer.errChan <- expectedErr

	err := kafkaConsumer.GetError()
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestKafkaConsumer_Consume_ContextCancellation(t *testing.T) {
	kafkaConsumer, _, mockConsumer, mockPartitionConsumer, ctrl := SetupTest(t)
	defer ctrl.Finish()

	partitions := []int32{0}
	mockConsumer.EXPECT().Partitions(kafkaConsumer.topic).Return(partitions, nil)
	mockConsumer.EXPECT().ConsumePartition(kafkaConsumer.topic, int32(0), sarama.OffsetNewest).Return(mockPartitionConsumer, nil)

	messages := make(chan *sarama.ConsumerMessage)
	mockPartitionConsumer.EXPECT().Messages().Return((<-chan *sarama.ConsumerMessage)(messages)).AnyTimes()
	mockPartitionConsumer.EXPECT().Errors().Return(nil).AnyTimes()
	mockPartitionConsumer.EXPECT().Close().Return(nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- kafkaConsumer.Consume(ctx)
	}()

	select {
	case err := <-errChan:
		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled))
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for Consume with canceled context")
	}
}

func TestKafkaConsumer_Consume_UnmarshalError_Critical(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_kafka.NewMockDB(ctrl)
	mockConsumer := mock_kafka.NewMockConsumer(ctrl)
	mockPartitionConsumer := mock_kafka.NewMockPartitionConsumer(ctrl)

	cfg := config.Config{
		Kafka: config.KafkaConfig{
			Topic:                         "test-topic",
			MaxRetries:                    3,
			InitialBackoff:                10 * time.Millisecond,
			TreatUnmarshalErrorAsCritical: true,
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	kafkaConsumer, err := NewKafkaConsumer(context.Background(), cfg, logger, &sarama.Config{}, mockConsumer, mockDB)
	assert.NoError(t, err)

	partitions := []int32{0}
	mockConsumer.EXPECT().Partitions(kafkaConsumer.topic).Return(partitions, nil)
	mockConsumer.EXPECT().ConsumePartition(kafkaConsumer.topic, int32(0), sarama.OffsetNewest).Return(mockPartitionConsumer, nil)

	messages := make(chan *sarama.ConsumerMessage, 1)
	mockPartitionConsumer.EXPECT().Messages().Return(messages).AnyTimes()
	mockPartitionConsumer.EXPECT().Errors().Return(nil).AnyTimes()
	mockPartitionConsumer.EXPECT().Close().Return(nil)

	messages <- &sarama.ConsumerMessage{Value: []byte("this is not json")}
	close(messages)

	err = kafkaConsumer.Consume(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal message")
}
