package pg

import (
	"fmt"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type PgController struct {
	db *PgDatabase

	entityDefinitions []protocol.EntityBased

	initialized bool

	baseEntities []protocol.EntityBased

	config Config
}

func NewPgController(dbConfig Config, entityDefinitions, baseEntities []protocol.EntityBased) protocol.DatabaseController {
	return &PgController{entityDefinitions: entityDefinitions, baseEntities: baseEntities, config: dbConfig}
}

func (d *PgController) GetTransactionFactory() (protocol.TransactionFactory, error) {
	if !d.initialized {

		err := d.Init()
		if err != nil {
			return nil, err
		}

	}

	return NewTransactionFactory(d.db), nil
}

const (
	DBCreateDML = "CREATE DATABASE %s "

	CreateUUIdExtension = `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`

	CreateSchemaIfNotExist = `CREATE SCHEMA IF NOT EXISTS "%s";`
)

const (
	ErrorCodeDbExist = "42P04"

	ErrorCodePermissionDenied = "42501"
)

func (d *PgController) Generate() error {
	sd := NewPgDatabase(d.config)

	d.db = sd

	err := sd.OpenRaw()
	if err != nil {
		return err
	}

	tx := sd.GetGorm().Begin()

	odb, err := sd.GetGorm().DB()
	if err != nil {
		return err
	}

	defer odb.Close()

	defer tx.Rollback()

	err = sd.GetGorm().Exec(fmt.Sprintf(DBCreateDML, d.config.Name)).Error

	if err != nil {

		if pgerror, ok := err.(*pgconn.PgError); ok {

			if pgerror.Code == ErrorCodeDbExist {

				tx.Rollback()

				return sd.Open()

			}

			if pgerror.Code == ErrorCodePermissionDenied && d.config.IgnorePermissionDenied {

				tx.Rollback()

				return sd.Open()

			}

		}

		return err

	}

	return sd.Open()
}

func (d *PgController) Init() error {
	db := NewTransactionFactory(d.db)

	tx := db.New()

	err := tx.Begin()

	defer tx.RollbackUnlessCommitted()

	if err != nil {
		return err
	}

	context := tx.GetDataContext().(*gorm.DB)

	isNew := true

	tbs, err := context.Migrator().GetTables()
	if err != nil {
		return err
	}

	if len(tbs) != 0 {
		isNew = false
	}

	if isNew {

		err = context.Exec(CreateUUIdExtension).Error

		if err != nil {
			return err
		}

		err = context.Exec(fmt.Sprintf(CreateSchemaIfNotExist, d.config.Schema)).Error

		if err != nil {
			return err
		}

	}

	for _, v := range d.entityDefinitions {

		err = context.AutoMigrate(&v)

		if err != nil {

			fmt.Printf("AutoMigrate: %T\n", v)

			return err

		}

	}

	if isNew {
		for _, v := range d.baseEntities {
			if err := context.Save(v).Error; err != nil {
				return err
			}
		}
	}

	err = tx.Commit()

	if err == nil {
		d.initialized = true
	}

	return err
}
