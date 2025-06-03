package domain

import "gorm.io/gorm"

type Transaction struct {
  gorm.Model
  ID        uint   `gorm:"primaryKey"`
  FromID    string `gorm:"not null"`
  ToID      string `gorm:"not null"`
  Amount    int    `gorm:"not null"`
  Time      string `gorm:"not null"`
  AccountID string `gorm:"not null"`


}
