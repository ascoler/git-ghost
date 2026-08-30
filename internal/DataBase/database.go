package database

import (
	"time"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type RepoState struct {
	ID         uint `gorm:"primarykey"`
	Path       string `gorm:"uniqueIndex;not null"`
	Hash string `gorm:"column:hash"`
	LastBackup *time.Time
	Status     string `gorm:"default:'never'"` 
	LastError  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type DB struct {
	*gorm.DB
}

func Open(dbPath string) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&RepoState{})
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}


func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
func (db *DB) GetHashRepo(path string,hash string) (bool,error) {
	var state RepoState
	result := db.Where("path =  ?",path).First(&state)
	if result.Error != nil{
		return false,result.Error
	}
	if hash == state.Hash {
		return false,nil
	}
	return true,nil
}

func (db *DB) UpdateRepoState(path string, status string, lastError string,hash ...string) error {
	now := time.Now()
	
	state := RepoState{
		Path:       path,
		Hash: 		hash[0],
		LastBackup: &now,
		Status:     status,
		LastError:  lastError,
		UpdatedAt: now,
	}
	if status == "ok"{
		state.LastBackup = &now
	}
	
	return db.Where("path = ?", path).Assign(state).FirstOrCreate(&state).Error
}


func (db *DB) GetRepoState(path string) (*RepoState, error) {
	var state RepoState
	err := db.Where("path = ?", path).First(&state).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil 
	}
	return &state, err
}


func (db *DB) GetAllRepoStates() ([]RepoState, error) {
	var states []RepoState
	err := db.Find(&states).Error
	return states, err
}


