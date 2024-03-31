package gorm

import (
	"context"
	"testing"
	"time"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/factory"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/gorm/pg"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	dbspec "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol/specification"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/misc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func DbControllerIntegerGenericRepository() protocol.DatabaseController {
	controller := factory.NewDatabaseController(factory.Sqlite,
		[]protocol.EntityBased{IntegerGenericRepositoryGormUser{}},
		[]protocol.EntityBased{},
		"memory",
		uuid.NewString(),
	)

	err := controller.Generate() // database and schema generation
	if err != nil {
		panic(err)
	}

	err = controller.Init() // Create tables

	if err != nil {
		panic(err)
	}

	return controller
}

type IntegerGenericRepositoryGormUser struct {
	Id     int64 `gorm:"primaryKey"`
	Number int64
}

func (i IntegerGenericRepositoryGormUser) GetCreatedAt() time.Time {
	return time.Time{}
}

func (i IntegerGenericRepositoryGormUser) GetUpdatedAt() *time.Time {
	return &time.Time{}
}

func (i IntegerGenericRepositoryGormUser) IsIdEmpty() bool {
	return i.Id == 0
}

func TestIntegerGenericRepositorySum(t *testing.T) {
	tx, err := DbControllerIntegerGenericRepository().GetTransactionFactory()
	assert.NoError(t, err)

	gormIntegerRepository := NewGormIntegerRepository[IntegerGenericRepositoryGormUser](tx.New(), pg.NewParser())

	entities := []IntegerGenericRepositoryGormUser{
		{Id: 1, Number: 10},
		{Id: 2, Number: 20},
	}

	for _, entity := range entities {
		err = gormIntegerRepository.Insert(context.Background(), &entity)
		assert.NoError(t, err)
	}

	// test without specification
	sum, err := gormIntegerRepository.Sum(context.Background(), "number", nil)
	assert.NoError(t, err)
	assert.Equal(t, float64(30), sum)

	// Test with using specification
	sum2, err := gormIntegerRepository.Sum(context.Background(), "number", dbspec.NewQuerySpecification(
		"id", misc.QueryOperatorEqual, misc.NewOperand(1)))
	assert.NoError(t, err)
	assert.Equal(t, float64(10), sum2)

	// Test ZeroSum
	zeroSum, err := gormIntegerRepository.Sum(context.Background(), "number", dbspec.NewQuerySpecification(
		"id", misc.QueryOperatorEqual, misc.NewOperand(30)))
	assert.NoError(t, err)
	assert.Equal(t, float64(0), zeroSum)
}

func TestIntegerGenericRepositoryAverage(t *testing.T) {
	tx, err := DbControllerIntegerGenericRepository().GetTransactionFactory()
	assert.NoError(t, err)

	gormIntegerRepository := NewGormIntegerRepository[IntegerGenericRepositoryGormUser](tx.New(), pg.NewParser())

	entities := []IntegerGenericRepositoryGormUser{
		{Id: 1, Number: 10},
		{Id: 2, Number: 20},
	}

	for _, entity := range entities {
		err = gormIntegerRepository.Insert(context.Background(), &entity)
		assert.NoError(t, err)
	}

	// Test without specification
	average1, err := gormIntegerRepository.Average(context.Background(), "number", nil)
	assert.NoError(t, err)
	assert.Equal(t, float64(15), average1)

	// Test with using specification
	average2, err := gormIntegerRepository.Average(context.Background(), "number", dbspec.NewQuerySpecification(
		"id", misc.QueryOperatorEqual, misc.NewOperand(entities[0].Id)))
	assert.NoError(t, err)
	assert.Equal(t, float64(10), average2)

	// Test Empty Result Set
	zeroSum, err := gormIntegerRepository.Average(context.Background(), "number", dbspec.NewQuerySpecification(
		"id", misc.QueryOperatorEqual, misc.NewOperand(30)))
	assert.NoError(t, err)
	assert.Equal(t, float64(0), zeroSum)
}

func TestIntegerRepositoryDistinctSum(t *testing.T) {
	tx, err := DbControllerIntegerGenericRepository().GetTransactionFactory()
	assert.NoError(t, err)

	gormUuidRepository := NewGormIntegerRepository[IntegerGenericRepositoryGormUser](tx.New(), pg.NewParser())

	entities := []IntegerGenericRepositoryGormUser{
		{Number: 10},
		{Number: 10},
		{Number: 20},
		{Number: 20},
		{Number: 20},
	}

	for _, entity := range entities {
		err = gormUuidRepository.Insert(context.Background(), &entity)
		assert.NoError(t, err)
	}

	// test without specification
	sum, err := gormUuidRepository.DistinctSum(context.Background(), "number", nil)
	assert.NoError(t, err)
	assert.Equal(t, float64(30), sum)

	// Test with using specification
	sum2, err := gormUuidRepository.DistinctSum(context.Background(), "number", dbspec.NewQuerySpecification(
		"number", misc.QueryOperatorEqual, misc.NewOperand(10)))
	assert.NoError(t, err)
	assert.Equal(t, float64(10), sum2)

	// Test ZeroSum
	zeroSum, err := gormUuidRepository.Sum(context.Background(), "number", dbspec.NewQuerySpecification(
		"id", misc.QueryOperatorEqual, misc.NewOperand(30)))
	assert.NoError(t, err)
	assert.Equal(t, float64(0), zeroSum)
}

func TestIntegerRepositoryDistinctCount(t *testing.T) {
	tx, err := DbControllerIntegerGenericRepository().GetTransactionFactory()
	assert.NoError(t, err)

	gormIntegerRepository := NewGormIntegerRepository[IntegerGenericRepositoryGormUser](tx.New(), pg.NewParser())

	entities := []IntegerGenericRepositoryGormUser{
		{Number: 10},
		{Number: 10},
		{Number: 20},
		{Number: 20},
		{Number: 20},
	}

	for _, entity := range entities {
		err = gormIntegerRepository.Insert(context.Background(), &entity)
		assert.NoError(t, err)
	}

	// test without specification
	sum, err := gormIntegerRepository.DistinctCount(context.Background(), "number", nil)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, sum)

	// Edge case: Query with no matches
	zeroMatchesQuery := dbspec.NewQuerySpecification(
		misc.Id, misc.QueryOperatorEqual, misc.NewOperand(100),
	)
	sum, err = gormIntegerRepository.DistinctCount(context.Background(), "number", zeroMatchesQuery)
	assert.NoError(t, err)
	assert.EqualValues(t, 0, sum)

	// Edge case: Query with one match
	oneMatchQuery := dbspec.NewQuerySpecification(
		misc.Id, misc.QueryOperatorEqual, misc.NewOperand(1),
	)
	sum, err = gormIntegerRepository.DistinctCount(context.Background(), "number", oneMatchQuery)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, sum)

	// Edge case: Query with multiple matches
	multipleMatchesQuery := dbspec.NewQuerySpecification(
		"number", misc.QueryOperatorEqual, misc.NewOperand(20),
	)
	sum, err = gormIntegerRepository.DistinctCount(context.Background(), "number", multipleMatchesQuery)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, sum)
}

func TestIntegerGenericRepositoryUpdateField(t *testing.T) {
	tx, err := DbControllerIntegerGenericRepository().GetTransactionFactory()
	assert.NoError(t, err)

	gormIntegerRepository := NewGormIntegerRepository[IntegerGenericRepositoryGormUser](tx.New(), pg.NewParser())

	entities := []IntegerGenericRepositoryGormUser{
		{Id: 1, Number: 10},
		{Id: 2, Number: 20},
	}

	for _, entity := range entities {
		err = gormIntegerRepository.Insert(context.Background(), &entity)
		assert.NoError(t, err)
	}

	// Test to make sure insert was successful
	res, err := gormIntegerRepository.GetById(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), res.Number)

	// Update  field using specification
	err = gormIntegerRepository.UpdateField(context.Background(), "number", 555, dbspec.NewQuerySpecification(
		"id", misc.QueryOperatorEqual, misc.NewOperand(entities[0].Id)))
	assert.NoError(t, err)

	// Test whether Update was successful
	res, err = gormIntegerRepository.GetById(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(555), res.Number)
}
