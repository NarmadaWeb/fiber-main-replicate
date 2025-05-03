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
	"github.com/NarmadaWeb/fiber-replicate-example/internal/stores" // Re-added for item store initialization
	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger" // Import GORM logger
)

// connectDatabase connects to the database using the provided DSN.
// This function is similar to the one in main_server, adapted for potential reuse.
func connectDatabase(dsn string) (*gorm.DB, error) {
	log.Println("INFO: Attempting to connect to replica database...")

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
		log.Printf("ERROR: Failed to connect to replica database: %v", err)
		return nil, err
	}

	log.Println("INFO: Replica database connection established successfully.")

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("ERROR: Failed to get underlying sql.DB for pool configuration: %v", err)
		// Still return the db instance even if pool config fails
		return db, nil
	}

	// Configure connection pool (can be adjusted based on replica needs)
	sqlDB.SetMaxIdleConns(5)  // Potentially lower than main server
	sqlDB.SetMaxOpenConns(50) // Potentially lower than main server
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	log.Println("INFO: Replica database connection pool configured (MaxIdle: 5, MaxOpen: 50, MaxLifetime: 1h, MaxIdleTime: 5m).")

	return db, nil
}

func main() {
	config.LoadConfig()
	// Use the Main DSN (as replica reads from the same DB in this setup)
	dbDSN := config.GetMainDBDSN()
	db, err := connectDatabase(dbDSN)
	if err != nil {
		log.Fatalf("FATAL: Could not connect to the replica database. Please check DSN and database status: %v", err)
	}

	// Get the underlying sql.DB handle for graceful closing
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("WARN: Could not get sql.DB handle for replica, manual closing might be needed: %v", err)
	} else {
		defer func() {
			log.Println("INFO: Closing replica database connection...")
			if closeErr := sqlDB.Close(); closeErr != nil {
				log.Printf("ERROR: Failed to close replica database connection gracefully: %v", closeErr)
			} else {
				log.Println("INFO: Replica database connection closed.")
			}
		}()
	}

	// Migrations are typically run by the main server, not the replica.
	// If needed, add migration logic here, but be cautious about running it on a replica.
	// log.Println("INFO: Skipping migrations on replica server.")

	// Initialize the item store, it will be passed to the router setup
	itemStore := stores.NewGormStore(db)
	log.Println("INFO: GORM item store initialized for replica.")

	app := fiber.New(fiber.Config{
		// Use Replica Server specific configurations
		AppName:           config.AppConfig.ReplicaServer.AppName,
		EnablePrintRoutes: true, // Useful for debugging routes
		Prefork:           false, // Prefork is generally false for simpler setups
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second, // Replica might primarily read, but keep write timeout reasonable
		IdleTimeout:       60 * time.Second,
	})

	// Setup routes specific to the replica server, passing the item store
	routers.SetupReplicaRoutes(app, itemStore) // Pass itemStore again

	// Graceful shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// Use Replica Server port
		addr := fmt.Sprintf(":%s", config.AppConfig.ReplicaServer.Port)
		log.Printf("INFO: Starting Replica Server (%s) on address %s...", config.AppConfig.ReplicaServer.AppName, addr)
		if err := app.Listen(addr); err != nil {
			// Check if the error is due to a normal shutdown
			if !errors.Is(err, context.Canceled) && !errors.Is(err, fiber.ErrInternalServerError) && err.Error() != "http: Server closed" {
				log.Printf("ERROR: Replica Server Listen failed unexpectedly: %v", err)
				// Attempt to trigger shutdown on listen error
				select {
				case quit <- syscall.SIGTERM: // Send signal to trigger shutdown
				default: // Avoid blocking if quit channel is full or closed
				}
			} else {
				log.Println("INFO: Replica Server listener closed.")
			}
		}
	}()

	// Wait for termination signal
	sig := <-quit
	log.Printf("INFO: Received signal: %s. Initiating graceful shutdown for Replica Server...", sig)

	// Create context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // 15-second timeout for shutdown
	defer cancel()

	log.Println("INFO: Shutting down Fiber server (Replica)...")
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("ERROR: Replica Fiber server graceful shutdown failed: %v", err)
	} else {
		log.Println("INFO: Replica Fiber server gracefully stopped.")
	}

	log.Println("INFO: Replica Server shutdown process complete.")
}
