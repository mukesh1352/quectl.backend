package config

import "gorm.io/gorm"

type Config struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

// Repository wraps access to the config table.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Set(key, value string) error {
	c := Config{Key: key, Value: value}
	return r.db.Save(&c).Error
}

func (r *Repository) Get(key string) (string, error) {
	var c Config
	if err := r.db.First(&c, "key = ?", key).Error; err != nil {
		return "", err
	}
	return c.Value, nil
}

func (r *Repository) All() ([]Config, error) {
	var items []Config
	if err := r.db.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
