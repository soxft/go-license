package server

type License struct {
	ID      int64  `gorm:"primaryKey;autoIncrement"`
	Serial  string `gorm:"uniqueIndex,size:128"`
	DueTime int64  `gorm:"type:bigint"`
}
