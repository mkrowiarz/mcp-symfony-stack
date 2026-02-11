package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mkrowiarz/mcp-symfony-stack/internal/core"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/config"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/dsn"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/executor"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/executor/engines"
)

func Dump(projectRoot, dbName string, tables []string) (*types.DumpResult, error) {
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return nil, err
	}

	if cfg.Database == nil {
		return nil, &types.CommandError{
			Code:    types.ErrConfigMissing,
			Message: "database configuration is required for dump operations",
		}
	}

	if err := core.IsDatabaseAllowed(dbName, cfg.Database.Allowed); err != nil {
		return nil, err
	}

	parsedDSN, err := dsn.ParseDSN(cfg.Database.DSN)
	if err != nil {
		return nil, err
	}

	engine := getEngine(parsedDSN.Engine)

	dbExecutor := executor.NewDockerDatabaseExecutor(engine, cfg.Docker.ComposeFile)

	destPath := filepath.Join(cfg.Database.DumpsPath,
		fmt.Sprintf("%s_%s.sql", dbName, time.Now().Format("2006-01-02T15-04")))

	return dbExecutor.Dump(cfg.Database.Service, parsedDSN, destPath, tables)
}

func CreateDB(projectRoot, dbName string) (*types.CreateResult, error) {
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return nil, err
	}

	if cfg.Database == nil {
		return nil, &types.CommandError{
			Code:    types.ErrConfigMissing,
			Message: "database configuration is required for create operations",
		}
	}

	if err := core.IsDatabaseAllowed(dbName, cfg.Database.Allowed); err != nil {
		return nil, err
	}

	parsedDSN, err := dsn.ParseDSN(cfg.Database.DSN)
	if err != nil {
		return nil, err
	}

	engine := getEngine(parsedDSN.Engine)

	dbExecutor := executor.NewDockerDatabaseExecutor(engine, cfg.Docker.ComposeFile)

	return dbExecutor.Create(cfg.Database.Service, parsedDSN, dbName)
}

func ImportDB(projectRoot, dbName, sourcePath string) (*types.ImportResult, error) {
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return nil, err
	}

	if cfg.Database == nil {
		return nil, &types.CommandError{
			Code:    types.ErrConfigMissing,
			Message: "database configuration is required for import operations",
		}
	}

	if err := core.IsDatabaseAllowed(dbName, cfg.Database.Allowed); err != nil {
		return nil, err
	}

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil, &types.CommandError{
			Code:    types.ErrFileNotFound,
			Message: fmt.Sprintf("SQL file not found: %s", sourcePath),
		}
	}

	parsedDSN, err := dsn.ParseDSN(cfg.Database.DSN)
	if err != nil {
		return nil, err
	}

	engine := getEngine(parsedDSN.Engine)

	dbExecutor := executor.NewDockerDatabaseExecutor(engine, cfg.Docker.ComposeFile)

	return dbExecutor.Import(cfg.Database.Service, parsedDSN, sourcePath, dbName)
}

func DropDB(projectRoot, dbName string) (*types.DropResult, error) {
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return nil, err
	}

	if cfg.Database == nil {
		return nil, &types.CommandError{
			Code:    types.ErrConfigMissing,
			Message: "database configuration is required for drop operations",
		}
	}

	if err := core.IsDatabaseAllowed(dbName, cfg.Database.Allowed); err != nil {
		return nil, err
	}

	parsedDSN, err := dsn.ParseDSN(cfg.Database.DSN)
	if err != nil {
		return nil, err
	}

	if err := core.IsNotDefaultDB(dbName, parsedDSN.Database); err != nil {
		return nil, err
	}

	engine := getEngine(parsedDSN.Engine)

	dbExecutor := executor.NewDockerDatabaseExecutor(engine, cfg.Docker.ComposeFile)

	return dbExecutor.Drop(cfg.Database.Service, parsedDSN, dbName)
}

func getEngine(engineType string) engines.DatabaseEngine {
	if engineType == "mariadb" {
		return engines.NewMySQLEngine(true)
	}
	return engines.NewMySQLEngine(false)
}
