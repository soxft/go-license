package server

type License struct {
	ID      int64  `gorm:"primaryKey;autoIncrement"`
	Serial  string `gorm:"uniqueIndex"`
	Mac     string `gorm:"uniqueIndex"`
	DueTime int64  `gorm:"type:bigint"`
}
