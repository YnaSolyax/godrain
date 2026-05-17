package logparser

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/N0tF0und04/godrain/storage/entity"
	"github.com/jaeyo/go-drain3/pkg/drain3"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
)

type fields struct {
	content int
}

func NewFileds(content int) *fields {
	return &fields{
		content: content,
	}
}

type Parser struct {
	repo      Repository
	batchSize int
	logger    *zap.Logger
}

func NewParser(repo Repository, batchSize int, logger *zap.Logger) *Parser {
	return &Parser{
		repo:      repo,
		batchSize: batchSize,
		logger:    logger,
	}
}

type Repository interface {
	FindDefectByText(ctx context.Context, text string, threshold float64) (uint, []float32, error)
	SaveLogItem(item entity.LogItem) error
}

func GetLogFields(format string) *fields {
	field := NewFileds(-1)
	findformat := strings.Fields(format)
	for i, ff := range findformat {
		clean := strings.Trim(ff, "[]")
		switch clean {
		case "Content":
			field.content = i
		}
	}

	return field
}

func (p *Parser) sanitize(s string) string {

	return strings.ReplaceAll(s, "\x00", "")
}

func (p *Parser) processBatch(ctx context.Context, d *drain3.Drain, lines []string, field *fields) error {

	for _, l := range lines {
		if len(strings.TrimSpace(l)) == 0 {
			p.logger.Info("log empty")
			continue
		}

		found := strings.Fields(l)
		cont := ""

		if field.content != -1 && len(found) > field.content {
			cont = strings.Join(found[field.content:], " ")
		} else {
			continue
		}

		cluster, clusterType, err := d.AddLogMessage(cont)
		if err != nil {
			p.logger.Error("drain AddLog error", zap.Error(err))
			continue
		}
		if clusterType != 1 && clusterType != 2 {
			//p.logger.Info("non unic cluster")
			continue
		}

		template := cluster.GetTemplate()
		id, vector, err := p.repo.FindDefectByText(ctx, template, 0.6)
		cleanContent := p.sanitize(template)
		item := entity.LogItem{
			Content: cleanContent,
			Vector:  pgvector.NewVector(vector),
		}

		if err != nil {
			p.logger.Error("defect not found")
			continue
		}

		if id != 0 {
			item.DefectID = &id
		}

		if err := p.repo.SaveLogItem(item); err != nil {
			p.logger.Error("can't add log to db", zap.Error(err))
			return err
		}
	}
	return nil
}

func (p *Parser) ParseLog(ctx context.Context, filename string, format string) error {

	d, err := drain3.NewDrain(
		drain3.WithDepth(3),
		drain3.WithSimTh(0.5),
		drain3.WithMaxChildren(100),
	)
	if err != nil {
		return err
	}

	indexes := GetLogFields(format)

	logDir := "logs" //hardcode
	fullPath := filepath.Join(logDir, filename)
	file, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	batch := make([]string, 0, p.batchSize)

	for scanner.Scan() {
		line := scanner.Text()
		batch = append(batch, line)

		if len(batch) >= p.batchSize {
			p.processBatch(ctx, d, batch, indexes)
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		p.processBatch(ctx, d, batch, indexes)
	}

	if err := scanner.Err(); err != nil {
		p.logger.Error("can't read file", zap.Error(err))
		return err
	}

	return nil
}
