package database

import (
	"log"
	"os"
	"time"

	"behavior_tree_engine/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDatabase(dsn string) {
	var err error

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color
		},
	)

	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established.")

	// Auto-migrate schemas
	// We will only migrate BehaviorTree and ExecutionLog for now.
	// The Node model is more of a schema for node types/templates or future expansion,
	// and not directly stored as individual linked records for tree structure in this iteration.
	err = DB.AutoMigrate(&models.BehaviorTree{}, &models.ExecutionLog{}, &models.Node{})
	if err != nil {
		log.Fatalf("Failed to migrate database schemas: %v", err)
	}
	log.Println("Database schemas migrated.")

	// Assign the global DB instance in the models package as well,
	// though accessing via database.DB is preferred to avoid package cycles if models need database.
	models.DB = DB
}
