package mongo

import (
	"context"

	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol"
)

type MongoController struct {
	db *MongoDatabase

	entityDefinitions []protocol.EntityBased

	initialized bool

	baseEntities []protocol.EntityBased

	config Config

	testDB bool
}

func NewMongoController(dbConfig Config, entityDefinitions, baseEntities []protocol.EntityBased, testDB bool) protocol.DatabaseController {
	return &MongoController{entityDefinitions: entityDefinitions, baseEntities: baseEntities, config: dbConfig, testDB: testDB}
}

func (d *MongoController) GetTransactionFactory() (protocol.TransactionFactory, error) {
	if !d.initialized {

		err := d.Init()
		if err != nil {
			return nil, err
		}

	}

	return NewTransactionFactory(d.db), nil
}

func (d *MongoController) Generate() error {
	md := NewMongoDatabase(d.config)

	d.db = md

	err := md.Open()
	if err != nil {
		return err
	}

	if d.testDB {
		return d.db.Database.Client().Database(d.config.Name).Drop(context.TODO())
	}

	return nil
}

func (d *MongoController) Init() error {
	d.initialized = true

	// TODO: add migration or add collection if not exist

	return nil
}
