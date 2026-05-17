package storage

import (
	"context"
	"time"

	ollama "github.com/N0tF0und04/godrain/internal/pkg"
	"github.com/N0tF0und04/godrain/storage/entity"
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

func (s *DBStorage) FindDefectByText(ctx context.Context, text string, threshold float64) (uint, []float32, error) {

	vec, err := ollama.GetVector(ctx, text)
	if err != nil {
		s.logger.Error("ollama err")
		return 0, nil, err
	}

	var defect entity.Defect
	pgVec := pgvector.NewVector(vec)
	query := `SELECT id FROM defects WHERE 1 - (vector <=> ?) > ? ORDER BY vector <=> ? LIMIT 1`
	err = s.DB.WithContext(ctx).Raw(query, pgVec, threshold, pgVec).Scan(&defect).Error
	if err != nil {
		return 0, vec, err
	}

	return defect.ID, vec, nil
}

func (s *DBStorage) SaveLogItem(item entity.LogItem) error {
	err := s.DB.Create(&item).Error
	if err != nil {
		s.logger.Error("GORM create error", zap.Error(err))
		return err
	}
	return nil
}

func (s *DBStorage) CreateDefect(ctx context.Context, description, solution string, vector []float32) error {
	defect := entity.Defect{
		Description: description,
		Solution:    solution,
		Vector:      pgvector.NewVector(vector),
		CreatedAt:   time.Now(),
	}

	if err := s.DB.WithContext(ctx).Create(&defect).Error; err != nil {
		s.logger.Info("Create deefct error")
		return err
	}
	result := s.DB.WithContext(ctx).Exec(`
        UPDATE log_items 
        SET defect_id = ? 
        WHERE defect_id IS NULL AND 1 - ("vector" <=> ?::vector) > 0.6`,
		defect.ID, pgvector.NewVector(vector))

	if result.Error != nil {
		s.logger.Error("failed to link existing logs to new defect")
	}

	s.logger.Info("Defect created and linked",
		zap.Uint("id", defect.ID),
		zap.Int64("linked_logs", result.RowsAffected))

	return nil
}
