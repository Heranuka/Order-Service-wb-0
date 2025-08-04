package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	_ "wb-l0/docs"
	"wb-l0/internal/components"
	"wb-l0/internal/config"

	"golang.org/x/sync/errgroup"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Println(err.Error())
		return
	}

	logger := components.SetupLogger(*cfg)

	eg, ctx := errgroup.WithContext(context.Background())

	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, os.Interrupt, syscall.SIGTERM)

	comp, err := components.InitComponents(ctx, cfg, logger)
	if err != nil {
		logger.Error("Bad configuration", slog.String("error", err.Error()))
		return
	}

	eg.Go(func() error {
		if err := comp.HttpServer.Run(ctx); err != nil {
			logger.Error("failed to run HttpServer", slog.String("error", err.Error()))
			return err
		}
		return nil
	})

	eg.Go(func() error {
		if err := comp.KafkaConsumer.Consume(ctx); err != nil {
			logger.Error("Kafka consumer failed", "error", err.Error())
			return err
		}
		return nil
	})

	<-sigQuit
	logger.Info("The program is exiting...")

	if err := comp.Shutdown(); err != nil {
		logger.Error("Error during shutdown", slog.String("error", err.Error()))
	}

	if err := eg.Wait(); err != nil {
		logger.Error("Error waiting for goroutines", slog.String("error", err.Error()))
	}

	logger.Info("The program has exited")
}
