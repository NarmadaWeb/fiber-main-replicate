package main

import (
	"context" // Diperlukan untuk Shutdown GORM
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time" // Diperlukan untuk GORM config & shutdown timeout

	"github.com/NarmadaWeb/fiber-replicate-example/internal/config"
	"github.com/NarmadaWeb/fiber-replicate-example/internal/routers"
	"github.com/NarmadaWeb/fiber-replicate-example/internal/stores"
	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger" // Import GORM logger
)

func connectDatabase(dsn string) (*gorm.DB, error) {
	log.Println("INFO: Attempting to connect to database...")

	// Konfigurasi logger GORM yang lebih detail
	gormLogLevel := logger.Info // Default: Info (Silent, Error, Warn, Info)
	if config.AppConfig.Log.Level == "warn" {
		gormLogLevel = logger.Warn
	} else if config.AppConfig.Log.Level == "error" {
		gormLogLevel = logger.Error
	} else if config.AppConfig.Log.Level == "silent" {
		gormLogLevel = logger.Silent
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer (log ke stdout)
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormLogLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,

	})

	if err != nil {
		log.Printf("ERROR: Failed to connect to database: %v", err)
		return nil, err
	}

	log.Println("INFO: Database connection established successfully.")

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("ERROR: Failed to get underlying sql.DB for pool configuration: %v", err)
		return db, nil
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	log.Println("INFO: Database connection pool configured (MaxIdle: 10, MaxOpen: 100, MaxLifetime: 1h, MaxIdleTime: 5m).")

	return db, nil
}

func main() {
	config.LoadConfig()
	dbDSN := config.GetMainDBDSN()
	db, err := connectDatabase(dbDSN)
	if err != nil {
		log.Fatalf("FATAL: Could not connect to the database. Please check DSN and database status: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("WARN: Could not get sql.DB handle, manual closing might be needed: %v", err)
	} else {
		defer func() {
			log.Println("INFO: Closing database connection...")
			if closeErr := sqlDB.Close(); closeErr != nil {
				log.Printf("ERROR: Failed to close database connection gracefully: %v", closeErr)
			} else {
				log.Println("INFO: Database connection closed.")
			}
		}()
	}

	if err := stores.RunMigrations(db); err != nil {
		log.Fatalf("FATAL: Could not run database migrations: %v", err)
	}

	itemStore := stores.NewGormStore(db)
	log.Println("INFO: GORM item store initialized.")

	app := fiber.New(fiber.Config{
		AppName:           config.AppConfig.MainServer.AppName,
		EnablePrintRoutes: true,
		Prefork:        false,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	routers.SetupMainRoutes(app, itemStore)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%s", config.AppConfig.MainServer.Port)
		log.Printf("INFO: Starting Main Server (%s) on address %s...", config.AppConfig.MainServer.AppName, addr)
		if err := app.Listen(addr); err != nil {
			// Periksa apakah error karena server ditutup secara normal
			if !errors.Is(err, context.Canceled) && !errors.Is(err, fiber.ErrInternalServerError) {
				log.Printf("ERROR: Main Server Listen failed unexpectedly: %v", err)
				select {
				case quit <- syscall.SIGTERM:
				default:
				}
			} else {
                 log.Println("INFO: Main Server listener closed.")
            }
		}
	}()

	sig := <-quit
	log.Printf("INFO: Received signal: %s. Initiating graceful shutdown...", sig)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Println("INFO: Shutting down Fiber server...")
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("ERROR: Fiber server graceful shutdown failed: %v", err)
	} else {
		log.Println("INFO: Fiber server gracefully stopped.")
	}

	log.Println("INFO: Main Server shutdown process complete.")
}
