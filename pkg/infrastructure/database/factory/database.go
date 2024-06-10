package factory

import (
	"errors"

	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/gorm/pg"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/gorm/sqlite"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/mongo"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol"
)

type DatabaseType string

const (
	Sqlite DatabaseType = "Gorm-Sqlite"

	Postgres DatabaseType = "Gorm-Pg"

	Mongo DatabaseType = "mongo"

	MongoTest DatabaseType = "mongo-test"

	Test DatabaseType = "Test"
)

var (
	ErrMissingParameter = errors.New("missing parameter")

	ErrNotImplemented = errors.New("not implemented")
)

func NewDatabaseController(name DatabaseType, definitions, baseEntities []protocol.EntityBased, param ...any) protocol.DatabaseController {
	if name == "" {
		name = Test
	}

	switch name {

	case Test:

		fallthrough

	case Sqlite:

		if len(param) != 0 {
			if param[0] == true || param[0] == "memory" {
				return sqlite.NewMemorySqliteController(definitions, param[1].(string))
			} else {
				return sqlite.NewSqliteController(param[0].(string), definitions, baseEntities, param[1].(string))
			}
		}

		panic(ErrMissingParameter)

	case Postgres:

		cfg := param[0].(PgConfig)

		return pg.NewPgController(pg.Config{
			Host: cfg.Host,

			Port: cfg.Port,

			Username: cfg.Username,

			Password: cfg.Password,

			Name: cfg.Name,

			Driver: cfg.Driver,

			Schema: cfg.Schema,

			IgnorePermissionDenied: cfg.IgnorePermissionDenied,
		}, definitions, baseEntities)

	case Mongo:

		cfg := param[0].(MongoConfig)

		return mongo.NewMongoController(mongo.Config{
			Host: cfg.Host,

			Port: cfg.Port,

			Username: cfg.Username,

			Password: cfg.Password,

			Name: cfg.Name,

			Driver: cfg.Driver,

			IgnorePermissionDenied: cfg.IgnorePermissionDenied,
		}, definitions, baseEntities, false)

	case MongoTest:

		cfg := param[0].(MongoConfig)

		return mongo.NewMongoController(mongo.Config{
			Host: cfg.Host,

			Port: cfg.Port,

			Username: cfg.Username,

			Password: cfg.Password,

			Name: cfg.Name,

			Driver: cfg.Driver,

			IgnorePermissionDenied: cfg.IgnorePermissionDenied,
		}, definitions, baseEntities, true)

	}

	panic(ErrNotImplemented)
}

type PgConfig struct {
	pg.Config
}

type MongoConfig struct {
	mongo.Config
}
