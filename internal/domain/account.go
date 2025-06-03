package domain

import "gorm.io/gorm"

type Account struct {
	  gorm.Model
	    ID      string `gorm:"primaryKey"`
	    Balance int
	    Transactions []Transaction `gorm:"foreignKey:AccountID"`
}