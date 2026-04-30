package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"

	kafkainfra "dlq_processor_ms/internal/infrastructure/kafka"
	"dlq_processor_ms/internal/infrastructure/storage"
	"dlq_processor_ms/internal/repository/postgres"
	"dlq_processor_ms/internal/usecase"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env file not found, falling back to environment variables")
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// --- Postgres ---
	dsn := mustEnv("POSTGRES_DSN")
	pool, err := storage.NewPostgresPool(ctx, dsn)
	if err != nil {
		logger.Error("postgres init failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// --- Wire layers ---
	repo := postgres.NewDLQRepository(pool)
	uc := usecase.NewDLQUseCase(repo, logger)
	handler := kafkainfra.NewDLQConsumerHandler(uc, logger)

	// --- Kafka ---
	brokers := strings.Split(mustEnv("KAFKA_BROKERS"), ",")
	cfg := sarama.NewConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	cfg.Version = sarama.V2_6_0_0

	cg, err := sarama.NewConsumerGroup(brokers, kafkainfra.ConsumerGroup, cfg)
	if err != nil {
		logger.Error("kafka consumer group init failed", "error", err)
		os.Exit(1)
	}
	defer cg.Close()

	logger.Info("dlq-processor started", "topic", kafkainfra.Topic, "group", kafkainfra.ConsumerGroup)

	for {
		if err := cg.Consume(ctx, []string{kafkainfra.Topic}, handler); err != nil {
			logger.Error("consumer group error", "error", err)
		}
		if ctx.Err() != nil {
			logger.Info("shutting down")
			return
		}
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("missing required env var", "key", key)
		os.Exit(1)
	}
	return v
}
