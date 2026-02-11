package dsn

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"
)

func ParseDSN(dsnString string) (*types.DSN, error) {
	if dsnString == "" {
		return nil, fmt.Errorf("empty DSN")
	}

	u, err := url.Parse(dsnString)
	if err != nil {
		return nil, fmt.Errorf("invalid DSN format: %w", err)
	}

	if u.Scheme == "" {
		return nil, fmt.Errorf("missing database scheme")
	}

	dsn := &types.DSN{
		User:     u.User.Username(),
		Host:     u.Hostname(),
		Database: strings.TrimPrefix(u.Path, "/"),
	}

	if password, ok := u.User.Password(); ok {
		dsn.Password = password
	}

	if u.Port() != "" {
		dsn.Port = u.Port()
	}

	query := u.Query()
	if serverVersion := query.Get("serverVersion"); serverVersion != "" {
		dsn.ServerVersion = serverVersion
	}

	dsn.Engine = determineEngine(u.Scheme, dsn.ServerVersion)

	if dsn.Port == "" {
		if dsn.Engine == "mysql" || dsn.Engine == "mariadb" {
			dsn.Port = "3306"
		} else if dsn.Engine == "postgresql" {
			dsn.Port = "5432"
		}
	}

	return dsn, nil
}

func determineEngine(scheme, serverVersion string) string {
	if strings.Contains(serverVersion, "mariadb") {
		return "mariadb"
	}

	switch scheme {
	case "mysql":
		return "mysql"
	case "postgresql", "postgres":
		return "postgresql"
	default:
		return scheme
	}
}
