package database

import (
	"fmt"
	"net/http"
	"os"
	"sme-backend/src/core/config"
	"sme-backend/src/enums/context_keys"
	"sme-backend/src/enums/dev_modes"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	ADMIN_DB    *gorm.DB
	API_USER_DB *gorm.DB
)

// GetRlsContextDB retrieves the database connection from the request context.
func GetRlsContextDB(request *http.Request) *gorm.DB {
	db, ok := request.Context().Value(context_keys.DB).(*gorm.DB)
	if !ok {
		panic("Failed to retrieve database connection from context")
	}
	return db
}

// ConnectDB initializes the database connection securely.
func ConnectDB() {
	// Securely retrieve environment variables
	portStr := config.Config("DB_PORT")
	host := config.Config("DB_HOST")
	dbname := config.Config("DB_NAME")
	api_user := config.Config("DB_USERNAME")
	api_password := config.Config("DB_PASSWORD")
	dev_mode := config.Config("DEV_MODE")

	// Validate environment variables
	if portStr == "" || host == "" || dbname == "" || api_user == "" || api_password == "" {
		panic("Missing required database environment variables")
	}

	port, err := strconv.ParseUint(portStr, 10, 32)
	if err != nil {
		panic("Invalid database port")
	}

	var sslmode string
	if dev_mode == dev_modes.PROD {
		sslmode = "verify-full"
	} else if dev_mode == dev_modes.DEV {
		sslmode = "disable"
	} else {
		panic("Invalid development mode for database ssl connection")
	}

	// Construct DSN (PostgreSQL connection string)
	api_user_dsn := fmt.Sprintf(`user=%s password=%s host=%s port=%d dbname=%s sslmode=%s`, api_user, api_password, host, port, dbname, sslmode)

	// Apply SSL configuration for production mode
	if dev_mode == dev_modes.PROD {
		db_certificate_path := config.Config("DB_SSL_CERTIFICATE")

		if db_certificate_path == "" {
			panic("DB_SSL_CERTIFICATE is not set in production")
		}

		db_certificate_path = strings.ReplaceAll(db_certificate_path, `\n`, "\n")
		filePath := "supabase-ssl-certificate.cer"

		// Securely write SSL certificate to a temporary file
		err := os.WriteFile(filePath, []byte(db_certificate_path), 0600)
		if err != nil {
			panic("Failed to write SSL certificate file securely")
		}

		api_user_dsn = fmt.Sprintf(`%s sslrootcert=%s`, api_user_dsn, filePath)
	}

	// Open database connection securely
	api_db, api_user_err := gorm.Open(postgres.Open(api_user_dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Error), // Set to `Error` to avoid logging sensitive data
		TranslateError: true,
		PrepareStmt:    false,
	})

	if api_user_err != nil {
		panic("Failed to establish secure database connection")
	}

	// Configure connection pooling
	sqlDB, err := api_db.DB()
	if err != nil {
		panic("Database connection pool initialization failed")
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(200)
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // Lower lifetime to avoid long-lived connections

	// Assign to global variables
	ADMIN_DB = api_db
	API_USER_DB = api_db

	fmt.Println("✅ Secure database connection established")
}
