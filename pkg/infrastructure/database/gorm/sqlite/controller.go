package sqlite

import (
	"errors"
	"os"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	"gorm.io/gorm"
)

type SqliteController struct {
	dataDir string

	db *SqliteDatabase

	entityDefinitions []protocol.EntityBased

	initialized bool

	baseEntities []protocol.EntityBased

	dbname string
}

func CreateDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {

		err = os.MkdirAll(dir, 0o755)

		return err

	}

	return nil
}

func NewSqliteController(dataDir string, entityDefinitions, baseEntities []protocol.EntityBased, dbname string) protocol.DatabaseController {
	err := CreateDir(dataDir)
	if err != nil {
		panic(err)
	}

	return &SqliteController{dataDir: dataDir, entityDefinitions: entityDefinitions, baseEntities: baseEntities, dbname: dbname}
}

func NewMemorySqliteController(entityDefinitions []protocol.EntityBased, dbname string) protocol.DatabaseController {
	return &SqliteController{dataDir: ":memory:", entityDefinitions: entityDefinitions, dbname: dbname}
}

var ErrDbNotFound = errors.New("database not found")

func (d *SqliteController) GetTransactionFactory() (protocol.TransactionFactory, error) {
	if !d.initialized {

		err := d.Init()
		if err != nil {
			return nil, err
		}

	}

	return NewTransactionFactory(d.db), nil
}

func (d *SqliteController) Generate() error {
	dir := d.dataDir

	sd := NewSqliteDatabase(dir == ":memory:", dir, d.dbname)

	d.db = sd

	err := sd.Open()
	if err != nil {
		return err
	}

	return err
}

func (d *SqliteController) Init() error {
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

	for _, v := range d.entityDefinitions {

		err = context.AutoMigrate(&v)

		if err != nil {
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
