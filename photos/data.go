package photos

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

type Item struct {
	gorm.Model
	ID          uint   `gorm:"primary_key" json:"id"`
	Image       string `json:"image"`
	Href        string `gorm:"index:idx_name,unique"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (i Item) ImageUrl() string {
	return fmt.Sprintf("%s/%s", i.CreatedAt.Format("2006/01"), i.Href)
}

func (i Item) ImagePath() string {
	return fmt.Sprintf("%s/%s", i.CreatedAt.Format("2006/01"), i.Image)
}

func (Item) TableName() string { return "photos" }

func NewDB(dsn string) (*gorm.DB, error) {
	var openning gorm.Dialector

	if _, err := os.Stat(dsn); os.IsNotExist(err) {
		openning = mysql.Open(dsn)
	} else {
		openning = sqlite.Open(dsn)
	}

	return gorm.Open(openning, &gorm.Config{})
}
