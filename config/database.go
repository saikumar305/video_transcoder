package config

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	host	 = "localhost"
	port     = "5432"
	user	 = "ai-nexus"
	password = "topochico&lime"
	dbname   = "video_transcoder"
)


func DatabaseConnection() *gorm.DB {

	DBSource :=  fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
								host, port, user, password, dbname)

	db, err := gorm.Open(postgres.Open(DBSource), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

