package storage

import (
	"time"

	ollama "github.com/YnaSolyax/godrain/pkg"
	"github.com/YnaSolyax/godrain/storage/entity"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBStorage struct {
	DB     *gorm.DB
	logger *zap.Logger
}

func NewDBStorage(db *gorm.DB, logger *zap.Logger) *DBStorage {
	return &DBStorage{
		DB:     db,
		logger: logger,
	}
}

func Conn(connStr string, logger *zap.Logger) (*gorm.DB, error) {

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		logger.Error("failed to connect database", zap.Error(err))
		return nil, err
	}

	err = db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error
	if err != nil {
		logger.Error("failed to create vector extension", zap.Error(err))
		return nil, err
	}

	err = db.AutoMigrate(
		&entity.Defect{},
		&entity.LogItem{},
	)
	if err != nil {
		logger.Error("migration failed", zap.Error(err))
		return nil, err
	}

	return db, nil
}

func (s *DBStorage) FindDefectByText(text string, threshold float64, logger *zap.Logger) (uint, []float32, error) {

	vec, err := ollama.GetVector(text)
	if err != nil {
		logger.Error("ollama err")
		return 0, nil, err
	}

	var defect entity.Defect
	pgVec := pgvector.NewVector(vec)
	query := `SELECT id FROM defects WHERE 1 - (vector <=> ?) > ? ORDER BY vector <=> ? LIMIT 1`

	err = s.DB.Raw(query, pgVec, threshold, pgVec).Scan(&defect).Error
	if err != nil {
		return 0, vec, err
	}

	return defect.ID, vec, nil
}

func (s *DBStorage) SaveLogItem(item entity.LogItem) error {
	err := s.DB.Create(&item).Error
	s.logger.Info("Create")
	if err != nil {
		s.logger.Error("GORM create error", zap.Error(err))
		return err
	}
	return nil
}

func (s *DBStorage) CreateDefect(description, solution string, vector []float32) error {
	defect := entity.Defect{
		Description: description,
		Solution:    solution,
		Vector:      pgvector.NewVector(vector),
		CreatedAt:   time.Now(),
	}
	return s.DB.Create(&defect).Error
}
