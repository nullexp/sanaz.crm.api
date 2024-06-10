package factory

import (
	dbimpl "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/gorm/sqlite"
	db "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol"
)

const Sqlite = "sqlite"

func NewDatabaseController(name string, baseEntities []db.EntityBased, param ...any) db.DatabaseController {
	if name == "" {
		name = Test
	}
	switch name {
	case Test:
		fallthrough
	case Sqlite:
		if len(param) != 0 {
			if param[0] == true || param[0] == "memory" {
				return dbimpl.NewMemorySqliteController(baseEntities, "memory")
			} else {
				dir, ok := param[0].(string)
				if !ok {
					panic(ErrMissingParameter)
				}
				return dbimpl.NewSqliteController(dir, baseEntities, []db.EntityBased{}, param[1].(string))
			}
		}
		panic(ErrMissingParameter)

	}
	panic(ErrNotImplemented)
}
