package engines

import "github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"

type DatabaseEngine interface {
	BuildDumpCommand(dsn *types.DSN, tables []string) []string
	BuildCreateCommand(dsn *types.DSN, dbName string) []string
	BuildImportCommand(dsn *types.DSN, dbName string) []string
	BuildDropCommand(dsn *types.DSN, dbName string) []string
	BuildListCommand(dsn *types.DSN) []string
	Name() string
}
