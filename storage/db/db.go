package storage

import (
	"database/sql"

	"github.com/YnaSolyax/godrain/storage/entity"
	_ "github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
)

type DBStorage struct {
	Conn   *sql.DB
	logger *zap.Logger
}

func NewDBStorage(connStr string, logger *zap.Logger) *DBStorage {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("open db")
	}

	if err := db.Ping(); err != nil {
		logger.Error("conn unable")
	}

	return &DBStorage{Conn: db}
}

func (s *DBStorage) FindDefectByVector(vector pgvector.Vector, threshold float64, logger *zap.Logger) (uint, bool, error) {
	var id uint
	query := `SELECT id FROM defects WHERE 1 - (vector <=> $1) > $2 ORDER BY vector <=> $1 LIMIT 1`
	err := s.Conn.QueryRow(query, vector, threshold).Scan(&id)
	if err == sql.ErrNoRows {
		logger.Error("Find Defect QueryRow")
		return 0, false, nil
	}
	return id, true, err
}

func (s *DBStorage) CreateDefect(template string, v pgvector.Vector, logger *zap.Logger) (uint, error) {
	var id uint
	err := s.Conn.QueryRow("INSERT INTO defects (description, vector) VALUES ($1, $2) RETURNING id", template, v).Scan(&id)
	if err != nil {
		logger.Error("Create Defect QueryRow")
	}
	return id, err
}

func (s *DBStorage) CreateIncident(source string) (uint, error) {
	var id uint
	query := `INSERT INTO incidents (source_file) VALUES ($1) RETURNING id`
	err := s.Conn.QueryRow(query, source).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *DBStorage) SaveLogItem(item entity.LogItem) error {
	query := `INSERT INTO log_items (incident_id, defect_id, timestamp, level, content, cluster_id) 
              VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := s.Conn.Exec(query, item.IncidentID, item.DefectID, item.Timestamp, item.Level, item.Content, item.ClusterID)
	return err
}
