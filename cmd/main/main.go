package main

import (
	"context"
	"file_share/cmd"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

const (
	codeError = 1
)

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {

	_ = godotenv.Load("./deploy/.debug.env")

	container, closer, err := cmd.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize container: %v\n", err)
		os.Exit(codeError)
	}

	defer closer()

	log := container.GetLogger()
	ctx := container.GetContext()

	ctxFields := map[string]string{
		"path": "cmd/service/main.go",
		"name": "main",
	}

	ctx = log.WithFields(ctx, ctxFields)
	log.Info(ctx, "application starting")

	migrator := container.GetMigrator()
	err = migrator.MigrateUp()
	if err != nil {
		log.Error(ctx, errors.Wrap(err, "error during migration"))
		os.Exit(codeError)
	}

	log.Info(ctx, "successful migration")

	scanService := container.GetScanService()
	scanDuration := time.Second * 5

	scanService.StartProcessScan(ctx, scanDuration)

	server := container.GetServer()
	go func() {
		log.Info(ctx, fmt.Sprintf("starting server on %s", container.GetServerAddress()))
		server.Run(ctx)
	}()

	// Настраиваем graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info(ctx, "shutting down server...")

	// Даем серверу время на завершение активных соединений
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Выполняем graceful shutdown для HTTP сервера
	if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
		log.Error(ctx, fmt.Errorf("server forced to shutdown: %w", shutdownErr))
	}

	log.Info(ctx, "server stopped")
}
