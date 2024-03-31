package pg

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PreloadAllPlugin struct{}

func (p *PreloadAllPlugin) Name() string {
	return "PreloadAllPlugin"
}

func (p *PreloadAllPlugin) Initialize(db *gorm.DB) error {
	// Add the Preload method to all queries

	return db.Callback().Query().Before("gorm:query").Register("preload_all", preloadThreeLevel)
}

func preloadThreeLevel(db *gorm.DB) {
	// Add your preload logic here

	db.Preload(clause.Associations).
		Preload(clause.Associations).
		Preload(clause.Associations).
		Preload(clause.Associations).
		Preload(clause.Associations)
}
