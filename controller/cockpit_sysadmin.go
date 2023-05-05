package controller

import "gorm.io/gorm"

type SysAdmin struct {
	gorm.Model
	AdminCredential AdminCredential `gorm:"not null"`
}
