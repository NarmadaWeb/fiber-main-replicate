package config

import (
	"fmt"
	"log"
	"strings" // Added for environment variable key replacer
	"time"

	"github.com/spf13/viper"
)


type Config struct {
	AppName       string         `mapstructure:"appName"`
	Log           LogConfig      `mapstructure:"log"`
	MainServer    ServerConfig   `mapstructure:"mainServer"`
	ReplicaServer ReplicaConfig  `mapstructure:"replicaServer"`
	Database      DatabaseConfig `mapstructure:"database"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type ServerConfig struct {
	Port    string `mapstructure:"port"`
	AppName string `mapstructure:"appName"`
}

type ReplicaConfig struct {
	Port              	string `mapstructure:"port"`
	AppName           	string `mapstructure:"appName"`
	MainServerAddress 	string `mapstructure:"mainServerAddress"`
	FetchTimeoutSeconds int    `mapstructure:"fetchTimeoutSeconds"`
}

type DatabaseConfig struct {
	Main DBConnection `mapstructure:"main"`
}

type DBConnection struct {
	DSN string `mapstructure:"dsn"`
}


var AppConfig Config

func LoadConfig() {
	// --- Environment Variable Setup ---
	viper.SetEnvPrefix("APP") // Prefix for environment variables (e.g., APP_DATABASE_MAIN_DSN)
	viper.AutomaticEnv()      // Read environment variables
	// Replace dots with underscores for nested env vars (e.g., database.main.dsn -> APP_DATABASE_MAIN_DSN)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// --- Set Default Config --- //
	viper.SetDefault("appName", "DefaultApp") // Use lowercase keys consistent with mapstructure tags
	viper.SetDefault("log.level", "info")
	viper.SetDefault("mainServer.port", "8000")
	viper.SetDefault("mainServer.appName", "Main")
	viper.SetDefault("replicaServer.port", "8001")
	viper.SetDefault("replicaServer.appName", "replica")
	viper.SetDefault("replicaServer.mainServerAddress", "http://localhost:8000")
	viper.SetDefault("replicaServer.fetchTimeoutSecond", 5)
	viper.SetDefault("database.main.dsn", "postgres://user:password@localhost:5432/appdb?sslmode=disable") // Default DSN

	// --- Config File Setup --- //
	viper.SetConfigName("config") // Name of config file (without extension)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		if _,ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("INFO: Configuration file (config.yaml) not found.")
		} else {
			// Error reading config file, but not critical if env vars are set
			log.Printf("WARN: Error reading configuration file (%s): %v", viper.ConfigFileUsed(), err)
		}
	} else {
		log.Printf("INFO: Using configuration file: %s", viper.ConfigFileUsed())
	}

	// --- Unmarshal Config (File values overridden by Env Vars) ---
	err := viper.Unmarshal(&AppConfig)
	if err != nil {
		log.Fatalf("FATAL: Unable to decode config into struct: %v", err)
	}

	// --- Validation and Post-processing ---
	// DSN validation (check after unmarshal)
	if AppConfig.Database.Main.DSN == "" {
		log.Fatalf("FATAL: Database DSN is not configured! Set APP_DATABASE_MAIN_DSN or database.main.dsn in config.yaml.")
	}

	// Validasi alamat main server untuk replica
	if AppConfig.ReplicaServer.MainServerAddress == "" && AppConfig.MainServer.Port != "" {
		log.Println("WARN: replicaServer.mainServerAddress is empty, attempting to construct from mainServer.port")
		AppConfig.ReplicaServer.MainServerAddress = fmt.Sprintf("http://localhost:%s", AppConfig.MainServer.Port)
	}
	// Validasi timeout replica
	if AppConfig.ReplicaServer.FetchTimeoutSeconds <= 0 {
		log.Printf("WARN: replicaServer.fetchTimeoutSeconds is invalid (%d), setting to default 5s", AppConfig.ReplicaServer.FetchTimeoutSeconds)
		AppConfig.ReplicaServer.FetchTimeoutSeconds = 5
	}

	log.Println("INFO: Configuration loaded successfully.")
	log.Printf("INFO: Main Server Port: %s", AppConfig.MainServer.Port)
	log.Printf("INFO: Replica Server Port: %s", AppConfig.ReplicaServer.Port)
	log.Printf("INFO: Replica will fetch from: %s", AppConfig.ReplicaServer.MainServerAddress)
	// Log the final DSN being used (could be from file or env var)
	// Avoid logging the full DSN if it contains sensitive info in production
	log.Printf("INFO: Main Database DSN configured (check logs for connection errors if this is incorrect)")
}

func GetReplicaFetchTimeout() time.Duration {
	// Ensure AppConfig is populated before accessing
	if AppConfig.ReplicaServer.FetchTimeoutSeconds <= 0 {
		log.Println("WARN: GetReplicaFetchTimeout called before config loaded or invalid value, returning default 5s")
		return 5 * time.Second
	}
	return time.Duration(AppConfig.ReplicaServer.FetchTimeoutSeconds) * time.Second
}

func GetMainDBDSN() string {
	// Ensure AppConfig is populated
	if AppConfig.Database.Main.DSN == "" {
		log.Fatal("FATAL: GetMainDBDSN called before config loaded or DSN is empty")
	}
	return AppConfig.Database.Main.DSN
}
