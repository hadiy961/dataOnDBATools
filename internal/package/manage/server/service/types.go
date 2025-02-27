package service

import (
	"database/sql"
	"dbaTools/internal/logger"
)

type ServerService struct {
	DB     *sql.DB
	Logger *logger.Logger
}

func NewServerService(db *sql.DB, log *logger.Logger) *ServerService {
	return &ServerService{
		DB:     db,
		Logger: log,
	}
}
