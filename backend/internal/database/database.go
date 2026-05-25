package database

import (
	"log"

	"hcm-backend/internal/config"
	"hcm-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), gormCfg)
	if err != nil {
		return nil, err
	}

	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		log.Printf("warning: could not create uuid-ossp extension: %v", err)
	}

	if err := db.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.EmployeeCustomField{},
		&models.Employee{},
		&models.EmployeeCustomFieldValue{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
