package engines

import (
	"fmt"

	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"
)

type MySQLEngine struct {
	isMariaDB bool
}

func NewMySQLEngine(isMariaDB bool) *MySQLEngine {
	return &MySQLEngine{isMariaDB: isMariaDB}
}

func (e *MySQLEngine) BuildDumpCommand(dsn *types.DSN, tables []string) []string {
	dumpCmd := "mysqldump"
	if e.isMariaDB {
		dumpCmd = "mariadb-dump"
	}

	cmd := []string{
		dumpCmd,
		"-h", dsn.Host,
		"-u", dsn.User,
		fmt.Sprintf("-p%s", dsn.Password),
		dsn.Database,
	}

	if len(tables) > 0 {
		cmd = append(cmd, tables...)
	}

	return cmd
}

func (e *MySQLEngine) BuildCreateCommand(dsn *types.DSN, dbName string) []string {
	return []string{
		"mysql",
		"-h", dsn.Host,
		"-u", dsn.User,
		fmt.Sprintf("-p%s", dsn.Password),
		"-e", fmt.Sprintf("CREATE DATABASE `%s`", dbName),
	}
}

func (e *MySQLEngine) BuildImportCommand(dsn *types.DSN, dbName string) []string {
	return []string{
		"mysql",
		"-h", dsn.Host,
		"-u", dsn.User,
		fmt.Sprintf("-p%s", dsn.Password),
		dbName,
	}
}

func (e *MySQLEngine) BuildDropCommand(dsn *types.DSN, dbName string) []string {
	return []string{
		"mysql",
		"-h", dsn.Host,
		"-u", dsn.User,
		fmt.Sprintf("-p%s", dsn.Password),
		"-e", fmt.Sprintf("DROP DATABASE `%s`", dbName),
	}
}

func (e *MySQLEngine) BuildListCommand(dsn *types.DSN) []string {
	return []string{
		"mysql",
		"-h", dsn.Host,
		"-u", dsn.User,
		fmt.Sprintf("-p%s", dsn.Password),
		"-e", "SHOW DATABASES",
	}
}

func (e *MySQLEngine) Name() string {
	if e.isMariaDB {
		return "MariaDB"
	}
	return "MySQL"
}
