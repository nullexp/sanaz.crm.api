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

func DbControllerUuidGenericRepository() protocol.DatabaseController {
	controller := factory.NewDatabaseController(factory.Sqlite,
		[]protocol.EntityBased{UuidGenericRepositoryGormUser{}},
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

type UuidGenericRepositoryGormUser struct {
	Id     string `gorm:"type:uuid;default:generate_uuid_v4"`
	Number int64
}

func (u UuidGenericRepositoryGormUser) GetCreatedAt() time.Time {
	return time.Time{}
}

func (u UuidGenericRepositoryGormUser) GetUpdatedAt() *time.Time {
	return &time.Time{}
}

func (u UuidGenericRepositoryGormUser) IsIdEmpty() bool {
	return u.Id == ""
}

func TestUUIDRepositorySum(t *testing.T) {
	tx, err := DbControllerUuidGenericRepository().GetTransactionFactory()
	assert.NoError(t, err)

	gormUuidRepository := NewGormUuidRepository[UuidGenericRepositoryGormUser](tx.New(), pg.NewParser())

	entities := []UuidGenericRepositoryGormUser{
		{Id: "c1a996eb-5a2c-4106-8c8e-116bf7716bec", Number: 10},
		{Id: "25a11c99-dce9-417b-ba00-827b2957da53", Number: 20},
	}

	for _, entity := range entities {
		err = gormUuidRepository.Insert(context.Background(), &entity)
		assert.NoError(t, err)
	}

	// test without specification
	sum, err := gormUuidRepository.Sum(context.Background(), "number", nil)
	assert.NoError(t, err)
	assert.Equal(t, float64(30), sum)

	// Test with using specification
	sum2, err := gormUuidRepository.Sum(context.Background(), "number", dbspec.NewQuerySpecification(
		"id", misc.QueryOperatorEqual, misc.NewOperand(entities[0].Id)))
	assert.NoError(t, err)
	assert.Equal(t, float64(10), sum2)

	// Test not found
	zeroSum, err := gormUuidRepository.Sum(context.Background(), "number", dbspec.NewQuerySpecification(
		misc.Id, misc.QueryOperatorEqual, misc.NewOperand(uuid.NewString())))
	assert.NoError(t, err)
	assert.Equal(t, float64(0), zeroSum)
}

func TestUuidGenericRepositoryAverage(t *testing.T) {
	tx, err := DbControllerUuidGenericRepository().GetTransactionFactory()
	assert.NoError(t, err)

	gormUuidRepository := NewGormUuidRepository[UuidGenericRepositoryGormUser](tx.New(), pg.NewParser())

	entities := []UuidGenericRepositoryGormUser{
		{Id: "c1a996eb-5a2c-4106-8c8e-116bf7716bec", Number: 10},
		{Id: "25a11c99-dce9-417b-ba00-827b2957da53", Number: 20},
	}

	for _, entity := range entities {
		err = gormUuidRepository.Insert(context.Background(), &entity)
		assert.NoError(t, err)
	}

	// test without specification
	average1, err := gormUuidRepository.Average(context.Background(), "number", nil)
	assert.NoError(t, err)
	assert.Equal(t, float64(15), average1)

	// Test with using specification
	average2, err := gormUuidRepository.Average(context.Background(), "number", dbspec.NewQuerySpecification(
		"id", misc.QueryOperatorEqual, misc.NewOperand("c1a996eb-5a2c-4106-8c8e-116bf7716bec")))
	assert.NoError(t, err)
	assert.Equal(t, float64(10), average2)

	// Test not found
	zeroAverage, err := gormUuidRepository.Average(context.Background(), "number", dbspec.NewQuerySpecification(
		misc.Id, misc.QueryOperatorEqual, misc.NewOperand(uuid.NewString())))
	assert.NoError(t, err)
	assert.Equal(t, float64(0), zeroAverage)
}

func TestUuidRepositoryDistinctSum(t *testing.T) {
	tx, err := DbControllerUuidGenericRepository().GetTransactionFactory()
	assert.NoError(t, err)

	gormUuidRepository := NewGormUuidRepository[UuidGenericRepositoryGormUser](tx.New(), pg.NewParser())

	entities := []UuidGenericRepositoryGormUser{
		{Id: "c1a996eb-5a2c-4106-8c8e-116bf7716bec", Number: 10},
		{Id: "c1a996eb-5a2c-4106-8c8e-116bf7711bec", Number: 10},
		{Id: "25a11c99-dce9-417b-ba00-827b29572a53", Number: 20},
		{Id: "25a11c99-dce9-417b-ba00-827b29573a53", Number: 20},
		{Id: "25a11c99-dce9-417b-ba00-827b29575a53", Number: 20},
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

	// Test not found
	zeroSum, err := gormUuidRepository.Sum(context.Background(), "number", dbspec.NewQuerySpecification(
		misc.Id, misc.QueryOperatorEqual, misc.NewOperand(uuid.NewString())))
	assert.NoError(t, err)
	assert.Equal(t, float64(0), zeroSum)
}

func TestUuidRepositoryDistinctCount(t *testing.T) {
	tx, err := DbControllerUuidGenericRepository().GetTransactionFactory()
	assert.NoError(t, err)

	gormUuidRepository := NewGormUuidRepository[UuidGenericRepositoryGormUser](tx.New(), pg.NewParser())

	entities := []UuidGenericRepositoryGormUser{
		{Id: "c1a996eb-5a2c-4106-8c8e-116bf7716bec", Number: 10},
		{Id: "c1a996eb-5a2c-4106-8c8e-116bf7711bec", Number: 10},
		{Id: "25a11c99-dce9-417b-ba00-827b29572a53", Number: 20},
		{Id: "25a11c99-dce9-417b-ba00-827b29573a53", Number: 20},
		{Id: "25a11c99-dce9-417b-ba00-827b29575a53", Number: 20},
	}

	for _, entity := range entities {
		err = gormUuidRepository.Insert(context.Background(), &entity)
		assert.NoError(t, err)
	}

	// test without specification
	sum, err := gormUuidRepository.DistinctCount(context.Background(), "number", nil)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, sum)

	// Edge case: Query with no matches
	zeroMatchesQuery := dbspec.NewQuerySpecification(
		misc.Id, misc.QueryOperatorEqual, misc.NewOperand(uuid.NewString()),
	)
	sum, err = gormUuidRepository.DistinctCount(context.Background(), "number", zeroMatchesQuery)
	assert.NoError(t, err)
	assert.EqualValues(t, 0, sum)

	// Edge case: Query with one match
	oneMatchQuery := dbspec.NewQuerySpecification(
		misc.Id, misc.QueryOperatorEqual, misc.NewOperand("c1a996eb-5a2c-4106-8c8e-116bf7716bec"),
	)
	sum, err = gormUuidRepository.DistinctCount(context.Background(), "number", oneMatchQuery)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, sum)

	// Edge case: Query with multiple matches
	multipleMatchesQuery := dbspec.NewQuerySpecification(
		"number", misc.QueryOperatorEqual, misc.NewOperand(20),
	)
	sum, err = gormUuidRepository.DistinctCount(context.Background(), "number", multipleMatchesQuery)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, sum)
}

func TestUuidGenericRepositoryUpdateField(t *testing.T) {
	tx, err := DbControllerUuidGenericRepository().GetTransactionFactory()
	assert.NoError(t, err)

	gormUuidRepository := NewGormUuidRepository[UuidGenericRepositoryGormUser](tx.New(), pg.NewParser())

	entities := []UuidGenericRepositoryGormUser{
		{Id: "c1a996eb-5a2c-4106-8c8e-116bf7716bec", Number: 10},
		{Id: "25a11c99-dce9-417b-ba00-827b2957da53", Number: 20},
	}

	for _, entity := range entities {
		err = gormUuidRepository.Insert(context.Background(), &entity)
		assert.NoError(t, err)
	}

	// update field
	err = gormUuidRepository.UpdateField(context.Background(), "number", 555, dbspec.NewQuerySpecification(
		"id", misc.QueryOperatorEqual, misc.NewOperand("c1a996eb-5a2c-4106-8c8e-116bf7716bec")))
	assert.NoError(t, err)

	// check wether the field updated
	getValue, err := gormUuidRepository.GetById(context.Background(), "c1a996eb-5a2c-4106-8c8e-116bf7716bec")
	assert.NoError(t, err)
	assert.Equal(t, int64(555), getValue.Number)
}
