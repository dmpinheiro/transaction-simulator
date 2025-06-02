package infrastructure

import (
	"fmt"

	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewGorm(config *viper.Viper) *gorm.DB {
	file := config.GetString("database.file");
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL", file)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
  if err != nil {
    panic(fmt.Errorf("error connecting database : %+v", err.Error()))
  }
  	return db;
}