package application

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/factory"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func DbController() protocol.DatabaseController {
	controller := factory.NewDatabaseController(factory.Sqlite,

		[]protocol.EntityBased{
			GormUser{},

			File{},

			NestedGormUser{},

			FullEntity{},
			EmbedEntity{},
		}, []protocol.EntityBased{},

		"memory", uuid.NewString(),
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

func DbPgController() protocol.DatabaseController {
	config := factory.PgConfig{}

	config.Host = "localhost"

	config.Port = 5432

	config.Username = "postgres"

	config.Password = "password"

	config.Name = "test"

	config.Driver = "postgres"

	config.Schema = "test"

	config.IgnorePermissionDenied = true

	controller := factory.NewDatabaseController(factory.Postgres,

		[]protocol.EntityBased{
			GormUser{},

			NestedGormUser{},

			File{},
		}, []protocol.EntityBased{},

		config,
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

type GormUser struct {
	Id       string `gorm:"type:uuid;default:generate_uuid_v4"`
	Number   int64
	FNumber  float64
	Username string
}

type GormUserSqlite struct {
	Id string `gorm:"type:uuid;"`

	GormUser
}

func (g GormUser) GetCreatedAt() time.Time {
	return time.Now()
}

func (g GormUser) GetUpdatedAt() *time.Time {
	return nil
}

func (g GormUser) GetUuid() string {
	return g.Id
}

func (g GormUser) IsIdEmpty() bool {
	return g.Id == ""
}

func (g GormUser) Validate(context.Context) error {
	return nil
}

func (u *GormUser) BeforeCreate(tx *gorm.DB) error {
	if u.Id == "" {
		u.Id = uuid.New().String()
	}

	return nil
}

type GormUserDto struct {
	GormUser
}

func TestBasicInsert(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	id, err := userApplication.Create(context.Background(), GormUserDto{})
	if err != nil {
		panic(err)
	}

	fmt.Println(id)
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(length int) string {
	result := make([]byte, length)

	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
}

func TestBasicGetById(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	model := GormUserDto{}

	model.Username = randomString(10)

	id, err := userApplication.Create(context.Background(), model)
	if err != nil {
		panic(err)
	}

	insertedEntity, err := userApplication.GetById(context.Background(), id)
	if err != nil {
		panic(err)
	}

	assert.EqualValues(t, model.Username, insertedEntity.Username)
}

func TestBasicUpdate(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	firstUsername := randomString(10)

	secondUsername := randomString(10)

	model := GormUserDto{}

	model.Username = firstUsername

	id, err := userApplication.Create(context.Background(), model)
	if err != nil {
		panic(err)
	}

	model.Username = secondUsername

	model.Id = id

	err = userApplication.Update(context.Background(), model)
	if err != nil {
		panic(err)
	}

	insertedEntity, err := userApplication.GetById(context.Background(), id)
	if err != nil {
		panic(err)
	}

	assert.EqualValues(t, insertedEntity.Username, secondUsername)
}

func TestBasicDelete(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	firstUsername := randomString(10)

	model := GormUserDto{}

	model.Username = firstUsername

	id, err := userApplication.Create(context.Background(), model)
	if err != nil {
		panic(err)
	}

	err = userApplication.Delete(context.Background(), QueryInfo{
		And: true,

		Queries: []BasicQuery{
			{Name: misc.Id, Op: misc.QueryOperatorEqual, Operand: misc.NewOperand(id)},
		},
	})
	if err != nil {
		panic(err)
	}

	insertedEntity, err := userApplication.GetById(context.Background(), id)
	if err != nil {
		panic(err)
	}

	assert.EqualValues(t, insertedEntity.Id, "")
}

func TestBasicList(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	model := GormUserDto{}

	model.Username = randomString(10)

	_, err := userApplication.Create(context.Background(), model)
	if err != nil {
		panic(err)
	}

	model = GormUserDto{}

	model.Username = randomString(10)

	_, err = userApplication.Create(context.Background(), model)
	if err != nil {
		panic(err)
	}

	insertedEntities, err := userApplication.List(context.Background())
	if err != nil {
		panic(err)
	}

	assert.EqualValues(t, 2, len(insertedEntities))
}

func TestBasicGetPage(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	for i := 0; i < 100; i++ {

		model := GormUserDto{}

		model.Username = strconv.Itoa(i)

		_, err := userApplication.Create(context.Background(), model)
		if err != nil {
			panic(err)
		}

	}

	insertedEntities, err := userApplication.GetPage(context.Background(), QueryPagination{Pagination: Pagination{Page: 2, Size: 10}})
	if err != nil {
		panic(err)
	}

	assert.EqualValues(t, 10, len(insertedEntities.Data))

	assert.EqualValues(t, 100, insertedEntities.TotalCount)
}

func TestBasicGetSingleQuery(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	model := GormUserDto{}

	model.Username = randomString(10)

	id, err := userApplication.Create(context.Background(), model)
	if err != nil {
		panic(err)
	}

	insertedEntity, err := userApplication.GetSingleQuery(context.Background(), QueryInfo{
		Queries: []BasicQuery{{Name: "username", Op: misc.QueryOperatorEqual, Operand: misc.NewOperand(model.Username)}},
		And:     true,
	})
	if err != nil {
		panic(err)
	}

	assert.EqualValues(t, id, insertedEntity.Id)
}

func TestBasicExist(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	model := GormUserDto{}

	model.Username = randomString(10)

	_, err := userApplication.Create(context.Background(), model)
	if err != nil {
		panic(err)
	}

	exist, err := userApplication.Exist(context.Background(), QueryInfo{
		Queries: []BasicQuery{{Name: "username", Op: misc.QueryOperatorEqual, Operand: misc.NewOperand(model.Username)}},

		And: true,
	})
	if err != nil {
		panic(err)
	}

	assert.True(t, exist)
}

func TestBasicSum(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	models := []GormUserDto{{
		GormUser{
			Id:       "4bf4eb72-6986-4dc1-a272-1af9608a7f93",
			Number:   10,
			Username: "10",
		},
	}, {
		GormUser{
			Id:       "20aa76d2-9d21-497d-a505-4d7f30762444",
			Number:   20,
			Username: "20",
		},
	}}

	for _, model := range models {
		_, err := userApplication.Create(context.Background(), model)
		if err != nil {
			panic(err)
		}
	}

	// test without specification
	sum1, err := userApplication.Sum(context.Background(), QueryFieldInfo{Field: "number"})
	assert.NoError(t, err)
	assert.EqualValues(t, float64(30), sum1)

	// test with using specification
	sum2, err := userApplication.Sum(context.Background(), QueryFieldInfo{
		QueryInfo: QueryInfo{Queries: []BasicQuery{{Name: "id", Op: misc.QueryOperatorEqual, Operand: misc.NewOperand(models[0].Id)}}},
		Field:     "number",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, float64(10), sum2)
}

func TestSumCanHandleFloat(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	models := []GormUserDto{{
		GormUser{
			Id:       "4bf4eb72-6986-4dc1-a272-1af9608a7f93",
			FNumber:  0.5,
			Username: "10",
		},
	}, {
		GormUser{
			Id:       "20aa76d2-9d21-497d-a505-4d7f30762444",
			FNumber:  1,
			Username: "20",
		},
	}}

	for _, model := range models {
		_, err := userApplication.Create(context.Background(), model)
		if err != nil {
			panic(err)
		}
	}

	// test without specification
	sum1, err := userApplication.Sum(context.Background(), QueryFieldInfo{Field: "f_number"})
	assert.NoError(t, err)
	assert.EqualValues(t, float64(1.5), sum1)

	// test with using specification
	sum2, err := userApplication.Sum(context.Background(), QueryFieldInfo{
		QueryInfo: QueryInfo{Queries: []BasicQuery{{Name: "id", Op: misc.QueryOperatorEqual, Operand: misc.NewOperand(models[0].Id)}}},
		Field:     "f_number",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, float64(0.5), sum2)
}

func TestBasicAverage(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	models := []GormUserDto{{
		GormUser{
			Id:       "4bf4eb72-6986-4dc1-a272-1af9608a7f93",
			Number:   10,
			Username: "10",
		},
	}, {
		GormUser{
			Id:       "20aa76d2-9d21-497d-a505-4d7f30762444",
			Number:   20,
			Username: "20",
		},
	}}

	for _, model := range models {
		_, err := userApplication.Create(context.Background(), model)
		if err != nil {
			panic(err)
		}
	}

	// test without specification
	average1, err := userApplication.Average(context.Background(), QueryFieldInfo{Field: "number"})
	assert.NoError(t, err)
	assert.EqualValues(t, int64(15), average1)

	qfi := QueryFieldInfo{Field: "number", QueryInfo: QueryInfo{Queries: []BasicQuery{{Name: "id", Op: misc.QueryOperatorEqual, Operand: misc.NewOperand(models[0].Id)}}, And: true}}
	// test with using specification
	average2, err := userApplication.Average(context.Background(), qfi)
	assert.NoError(t, err)
	assert.EqualValues(t, int64(10), average2)
}

func TestBasicCount(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	model := GormUserDto{}
	userName := randomString(5)
	model.Username = userName

	// insert a username
	model.Username = userName
	_, err := userApplication.Create(context.Background(), model)
	if err != nil {
		panic(err)
	}
	// insert a username inequal with the previous username 100 times
	model.Username = userName + randomString(1)
	for i := 0; i < 100; i++ {
		_, err := userApplication.Create(context.Background(), model)
		if err != nil {
			panic(err)
		}
	}

	count, err := userApplication.Count(context.Background(), QueryInfo{
		Queries: []BasicQuery{{Name: "username", Op: misc.QueryOperatorEqual, Operand: misc.NewOperand(model.Username)}},
		And:     true,
	})
	if err != nil {
		panic(err)
	}

	assert.EqualValues(t, 100, count)
}

type File struct {
	Id string `gorm:"type:uuid;"`

	FileId string

	NestedGormUserId string
}

type FileDto struct {
	Id string

	FileId string
}

func (g File) GetCreatedAt() time.Time {
	return time.Now()
}

func (g File) GetUpdatedAt() *time.Time {
	return nil
}

func (g File) GetUuid() string {
	return g.Id
}

func (g File) IsIdEmpty() bool {
	return g.Id == ""
}

func (g File) Validate(context.Context) error {
	return nil
}

func (u *File) BeforeCreate(tx *gorm.DB) error {
	if u.Id == "" {
		u.Id = uuid.New().String()
	}

	return nil
}

type NestedGormUser struct {
	Id string `gorm:"type:uuid;"`

	Username string

	Files []File `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (g NestedGormUser) GetCreatedAt() time.Time {
	return time.Now()
}

func (g NestedGormUser) GetUpdatedAt() *time.Time {
	return nil
}

func (g NestedGormUser) GetUuid() string {
	return g.Id
}

func (g NestedGormUser) IsIdEmpty() bool {
	return g.Id == ""
}

func (g NestedGormUser) Validate(context.Context) error {
	return nil
}

func (u *NestedGormUser) BeforeCreate(tx *gorm.DB) error {
	if u.Id == "" {
		u.Id = uuid.New().String()
	}

	return nil
}

type NestedGormUserDto struct {
	Id string

	Username string

	Files []FileDto
}

func (g NestedGormUserDto) GetCreatedAt() time.Time {
	return time.Now()
}

func (g NestedGormUserDto) GetUpdatedAt() *time.Time {
	return nil
}

func (g NestedGormUserDto) GetUuid() string {
	return g.Id
}

func (g NestedGormUserDto) IsIdEmpty() bool {
	return g.Id == ""
}

func (g NestedGormUserDto) Validate(context.Context) error {
	return nil
}

func TestNestedInsert(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[NestedGormUser, NestedGormUserDto](db, true)

	file := FileDto{}

	file.FileId = uuid.NewString()

	id, err := userApplication.Create(context.Background(), NestedGormUserDto{Username: "username", Files: []FileDto{file, file}})
	if err != nil {
		panic(err)
	}

	fmt.Println(id)
}

func TestNestedGet(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[NestedGormUser, NestedGormUserDto](db, true)

	file := FileDto{}

	file.FileId = uuid.NewString()

	id, err := userApplication.Create(context.Background(), NestedGormUserDto{Username: randomString(10), Files: []FileDto{file, file}})
	if err != nil {
		panic(err)
	}

	inserted, err := userApplication.GetById(context.Background(), id)
	if err != nil {
		panic(err)
	}

	assert.NotEmpty(t, inserted.Username)

	assert.NotZero(t, len(inserted.Files))
}

func TestNestedUpdate(t *testing.T) {
	// Nested Update us not supported

	db := DbController()

	userApplication := NewGormBasicAutoCrud[NestedGormUser, NestedGormUserDto](db, true)

	firstUsername := randomString(10)

	secondUsername := randomString(10)

	file := FileDto{}

	file.FileId = uuid.NewString()

	id, err := userApplication.Create(context.Background(), NestedGormUserDto{Username: firstUsername, Files: []FileDto{file, file}})
	if err != nil {
		panic(err)
	}

	err = userApplication.Update(context.Background(), NestedGormUserDto{Id: id, Username: secondUsername})
	if err != nil {
		panic(err)
	}

	inserted, err := userApplication.GetById(context.Background(), id)
	if err != nil {
		panic(err)
	}

	assert.EqualValues(t, secondUsername, inserted.Username)
}

func TestNestedSet(t *testing.T) {
	// Nested Update us not supported

	db := DbController()

	userApplication := NewGormBasicAutoCrud[NestedGormUser, NestedGormUserDto](db, true)

	firstUsername := randomString(10)

	secondUsername := randomString(10)

	file := FileDto{}

	file.FileId = uuid.NewString()

	id, err := userApplication.Create(context.Background(), NestedGormUserDto{Username: firstUsername, Files: []FileDto{file, file}})
	if err != nil {
		panic(err)
	}

	inserted, err := userApplication.GetById(context.Background(), id)
	if err != nil {
		panic(err)
	}

	assert.NotEmpty(t, inserted.GetUuid())

	nid, err := userApplication.Set(context.Background(),

		QueryDtoInfo[NestedGormUserDto]{
			QueryInfo: QueryInfo{},

			Dto: NestedGormUserDto{Id: id, Username: secondUsername, Files: []FileDto{file}},
		})
	if err != nil {
		panic(err)
	}

	ninserted, err := userApplication.GetById(context.Background(), id)
	if err != nil {
		panic(err)
	}

	assert.EqualValues(t, secondUsername, ninserted.Username)

	assert.EqualValues(t, id, nid)
}

func TestNestedDelete(t *testing.T) {
	// Nested Update us not supported

	db := DbController()

	userApplication := NewGormBasicAutoCrud[NestedGormUser, NestedGormUserDto](db, true)

	firstUsername := randomString(10)

	file := FileDto{}

	file.FileId = uuid.NewString()

	id, err := userApplication.Create(context.Background(), NestedGormUserDto{Username: firstUsername, Files: []FileDto{file, file}})
	if err != nil {
		panic(err)
	}

	err = userApplication.Delete(context.Background(),

		QueryInfo{
			And: true,

			Queries: []BasicQuery{{
				Name: misc.Id,

				Op: misc.QueryOperatorEqual,

				Operand: misc.NewOperand(id),
			}},
		})
	if err != nil {
		panic(err)
	}

	inserted, err := userApplication.GetById(context.Background(), id)
	if err != nil {
		panic(err)
	}

	assert.True(t, inserted.IsIdEmpty())
}

func TestNestedGetSingleQuery(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[NestedGormUser, NestedGormUserDto](db, true)

	file := FileDto{}

	file.FileId = uuid.NewString()

	username := randomString(10)
	_, err := userApplication.Create(context.Background(), NestedGormUserDto{Username: username, Files: []FileDto{file, file}})
	if err != nil {
		panic(err)
	}

	inserted, err := userApplication.GetSingleQuery(context.Background(), QueryInfo{
		Queries: []BasicQuery{{Name: "username", Op: misc.QueryOperatorEqual, Operand: misc.NewOperand(username)}},
	})
	if err != nil {
		panic(err)
	}

	assert.EqualValues(t, file.FileId, inserted.Files[0].FileId)
}

type EmbedEntity struct {
	Id           string `gorm:"type:uuid;"`
	FullEntityId string
	Username     string
}

func (e EmbedEntity) GetCreatedAt() time.Time {
	return time.Now()
}

func (e EmbedEntity) GetUpdatedAt() *time.Time {
	return nil
}

func (e EmbedEntity) IsIdEmpty() bool {
	return e.Id == ""
}

type FullEntity struct {
	Id string `gorm:"type:uuid;"`

	Name string

	LastName string

	EmbedEntity []EmbedEntity `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (u *FullEntity) BeforeCreate(tx *gorm.DB) error {
	if u.Id == "" {
		u.Id = uuid.New().String()
	}

	return nil
}

func (p FullEntity) GetCreatedAt() time.Time {
	return time.Now()
}

func (p FullEntity) GetUpdatedAt() *time.Time {
	return nil
}

func (p FullEntity) IsIdEmpty() bool {
	return p.Id == ""
}

func (p FullEntity) GetUuid() string {
	return p.Id
}

type EmbedEntityDto struct {
	Id           string
	FullEntityId string
	Username     string
}
type PartialFullDto struct {
	Id          string
	Name        string
	LastName    string
	EmbedEntity []EmbedEntityDto
}

type PartialDto1 struct {
	Id       string
	LastName string
}

func (p PartialDto1) Validate(ctx context.Context) error {
	return nil
}

func (p PartialDto1) IsIdEmpty() bool {
	return p.Id == ""
}

func (p PartialDto1) GetUuid() string {
	return p.Id
}

func (p PartialFullDto) Validate(ctx context.Context) error {
	return nil
}

func (p PartialFullDto) IsIdEmpty() bool {
	return p.Id == ""
}

func (p PartialFullDto) GetUuid() string {
	return p.Id
}

func TestPartial(t *testing.T) {
	// Nested Update us not supported

	db := DbController()

	userApplication := NewGormBasicAutoCrud[FullEntity, PartialFullDto](db, true)

	name := randomString(10)
	username := randomString(10)

	lastName := randomString(10)
	updateLastName := randomString(10)

	file := FileDto{}

	file.FileId = uuid.NewString()

	id, err := userApplication.Create(context.Background(), PartialFullDto{
		Name:     name,
		LastName: lastName,
		EmbedEntity: []EmbedEntityDto{
			{
				Id:       uuid.NewString(),
				Username: username,
			},
		},
	})
	if err != nil {
		panic(err)
	}

	err = userApplication.PartialUpdate(context.Background(), PartialDto1{Id: id, LastName: updateLastName})
	if err != nil {
		panic(err)
	}

	inserted, err := userApplication.GetById(context.Background(), id)
	if err != nil {
		panic(err)
	}

	assert.EqualValues(t, updateLastName, inserted.LastName)
	assert.EqualValues(t, name, inserted.Name)
}

func TestCreateMultiple(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	models := []GormUserDto{{
		GormUser{
			Id:       "4bf4eb72-6986-4dc1-a272-1af9608a7f93",
			Number:   10,
			Username: "10",
		},
	}, {
		GormUser{
			Id:       "20aa76d2-9d21-497d-a505-4d7f30762444",
			Number:   20,
			Username: "20",
		},
	}}

	ids, err := userApplication.CreateMultiple(context.Background(), models)
	assert.NoError(t, err)

	inserted1, err := userApplication.GetById(context.Background(), ids[0])
	if err != nil {
		panic(err)
	}

	inserted2, err := userApplication.GetById(context.Background(), ids[1])
	if err != nil {
		panic(err)
	}

	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(ids))
	assert.EqualValues(t, inserted1.Id, ids[0])
	assert.EqualValues(t, inserted2.Id, ids[1])
}

func TestSetMultiple(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	models := []GormUserDto{{
		GormUser{
			Id:       "4bf4eb72-6986-4dc1-a272-1af9608a7f93",
			Number:   10,
			Username: "10",
		},
	}, {
		GormUser{
			Number:   20,
			Username: "20",
		},
	}}

	ids, err := userApplication.CreateMultiple(context.Background(), models)
	assert.NoError(t, err)
	q := QueryDtoInfo[GormUserDto]{
		QueryInfo: QueryInfo{}, Dto: GormUserDto{GormUser: GormUser{
			Number:   10,
			Username: "10",
		}},
	}

	inserted1, err := userApplication.GetById(context.Background(), ids[0])
	if err != nil {
		panic(err)
	}

	inserted2, err := userApplication.GetById(context.Background(), ids[1])
	if err != nil {
		panic(err)
	}

	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(ids))
	assert.EqualValues(t, inserted1.Id, ids[0])
	assert.EqualValues(t, inserted2.Id, ids[1])

	q2 := QueryDtoInfo[GormUserDto]{
		QueryInfo: QueryInfo{}, Dto: GormUserDto{GormUser: GormUser{
			Id:       "20aa76d2-9d21-497d-a505-4d7f30762444",
			Number:   20,
			Username: "20",
		}},
	}
	var queries []QueryDtoInfo[GormUserDto]
	queries = append(queries, q)
	queries = append(queries, q2)

	ids, err = userApplication.SetMultiple(context.Background(), queries)
	assert.NoError(t, err)

	inserted1, err = userApplication.GetById(context.Background(), ids[0])
	if err != nil {
		panic(err)
	}

	inserted2, err = userApplication.GetById(context.Background(), ids[1])
	if err != nil {
		panic(err)
	}

	assert.EqualValues(t, 2, len(ids))
	assert.EqualValues(t, inserted1.Id, ids[0])
	assert.EqualValues(t, inserted2.Id, ids[1])
}

func TestNestedMultiple(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[NestedGormUser, NestedGormUserDto](db, true)

	file := FileDto{}

	file.FileId = uuid.NewString()

	ids, err := userApplication.CreateMultiple(context.Background(), []NestedGormUserDto{{Username: randomString(10), Files: []FileDto{file, file}}, {Username: randomString(10), Files: []FileDto{file, file}}})
	assert.NoError(t, err)
	inserted1, err := userApplication.GetById(context.Background(), ids[0])
	if err != nil {
		panic(err)
	}
	assert.NoError(t, err)
	inserted2, err := userApplication.GetById(context.Background(), ids[1])
	if err != nil {
		panic(err)
	}
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(ids))
	assert.EqualValues(t, inserted1.Id, ids[0])
	assert.EqualValues(t, inserted2.Id, ids[1])

	q := QueryDtoInfo[NestedGormUserDto]{
		QueryInfo: QueryInfo{}, Dto: NestedGormUserDto{Id: ids[0], Username: randomString(10), Files: []FileDto{file, file}},
	}

	q2 := QueryDtoInfo[NestedGormUserDto]{
		QueryInfo: QueryInfo{}, Dto: NestedGormUserDto{Id: ids[1], Username: randomString(10), Files: []FileDto{file, file}},
	}

	var queries []QueryDtoInfo[NestedGormUserDto]
	queries = append(queries, q)
	queries = append(queries, q2)
	ids2, err := userApplication.SetMultiple(context.Background(), queries)
	if err != nil {
		panic(err)
	}

	assert.NoError(t, err)
	inserted1, err = userApplication.GetById(context.Background(), ids[0])
	if err != nil {
		panic(err)
	}
	assert.NoError(t, err)
	inserted2, err = userApplication.GetById(context.Background(), ids[1])
	if err != nil {
		panic(err)
	}
	assert.NoError(t, err)
	assert.EqualValues(t, 2, len(ids))
	assert.EqualValues(t, inserted1.Id, ids2[0])
	assert.EqualValues(t, inserted2.Id, ids2[1])
}

func TestBasicRate(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	models := []GormUserDto{
		{
			GormUser{
				Id:       "4bf4eb72-6986-4dc1-a272-1af9608a7f93",
				Number:   10,
				Username: "aaa",
			},
		},
		{
			GormUser{
				Id:       "20aa76d2-9d21-497d-a505-4d7f30762444",
				Number:   20,
				Username: "bbb",
			},
		},
		{
			GormUser{
				Id:       "cddfe42e-d3a7-45eb-bbe3-2256bdf5dc3a",
				Number:   30,
				Username: "ccc",
			},
		},
		{
			GormUser{
				Id:       "9c417e99-d42a-441f-9aef-708eccc77704",
				Number:   40,
				Username: "ddd",
			},
		},
		{
			GormUser{
				Id:       "d98c5963-9109-4926-b333-bcc2af2e1bc7",
				Number:   50,
				Username: "eee",
			},
		},
		{
			GormUser{
				Id:       "01c03a98-7e8d-4799-989c-7ac25a7905b0",
				Number:   60,
				Username: "fff",
			},
		},
	}

	for _, model := range models {
		_, err := userApplication.Create(context.Background(), model)
		if err != nil {
			panic(err)
		}
	}

	// Nominator: 10+20+30 Denominator: 40+50+60
	rr1 := RateRequest{
		QueryFieldInfoNumerator: QueryFieldInfo{
			QueryInfo: QueryInfo{
				Queries: []BasicQuery{
					{
						Name:    "username",
						Op:      misc.QueryOperatorEqualOrLessThan,
						Operand: misc.NewOperand(models[2].Username),
					},
				},
			},
			Field: "number",
		},
		QueryFieldInfoDenominator: QueryFieldInfo{
			QueryInfo: QueryInfo{
				Queries: []BasicQuery{
					{
						Name:    "username",
						Op:      misc.QueryOperatorMoreThan,
						Operand: misc.NewOperand(models[2].Username),
					},
				},
			},
			Field: "number",
		},
	}

	rate1, err := userApplication.Rate(context.Background(), rr1)
	assert.NoError(t, err)
	assert.EqualValues(t, float64(40.0), rate1)

	// Nominator:30 Denominator: 50
	rr2 := RateRequest{
		QueryFieldInfoNumerator: QueryFieldInfo{
			QueryInfo: QueryInfo{
				Queries: []BasicQuery{
					{
						Name:    "username",
						Op:      misc.QueryOperatorEqual,
						Operand: misc.NewOperand(models[2].Username),
					},
				},
			},
			Field: "number",
		},
		QueryFieldInfoDenominator: QueryFieldInfo{
			QueryInfo: QueryInfo{
				Queries: []BasicQuery{
					{
						Name:    "username",
						Op:      misc.QueryOperatorEqual,
						Operand: misc.NewOperand(models[4].Username),
					},
				},
			},
			Field: "number",
		},
	}

	rate2, err := userApplication.Rate(context.Background(), rr2)
	assert.NoError(t, err)
	assert.EqualValues(t, float64(60.0), rate2)
}

func TestBasicDistinctSum(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	models := []GormUserDto{
		{
			GormUser{
				Id:       "4bf4eb72-6986-4dc1-a272-1af9608a7f93",
				Number:   2,
				Username: "aaa",
			},
		},
		{
			GormUser{
				Id:       "20aa76d2-9d21-497d-a505-4d7f30762444",
				Number:   3,
				Username: "bbb",
			},
		},
		{
			GormUser{
				Id:       "cddfe42e-d3a7-45eb-bbe3-2256bdf5dc3a",
				Number:   5,
				Username: "ccc",
			},
		},
		{
			GormUser{
				Id:       "9c417e99-d42a-441f-9aef-708eccc77704",
				Number:   5,
				Username: "ddd",
			},
		},
		{
			GormUser{
				Id:       "d98c5963-9109-4926-b333-bcc2af2e1bc7",
				Number:   5,
				Username: "eee",
			},
		},
		{
			GormUser{
				Id:       "01c03a98-7e8d-4799-989c-7ac25a7905b0",
				Number:   5,
				Username: "fff",
			},
		},
	}

	for _, model := range models {
		_, err := userApplication.Create(context.Background(), model)
		if err != nil {
			panic(err)
		}
	}

	qfi1 := QueryFieldInfo{
		QueryInfo: QueryInfo{
			Queries: []BasicQuery{},
		},
		Field: "number",
	}

	qfi2 := QueryFieldInfo{
		QueryInfo: QueryInfo{
			Queries: []BasicQuery{
				{
					Name:    "username",
					Op:      misc.QueryOperatorMoreThan,
					Operand: misc.NewOperand(models[2].Username),
				},
			},
		},
		Field: "number",
	}

	distinctSum1, err := userApplication.DistinctSum(context.Background(), qfi1)
	assert.NoError(t, err)
	assert.EqualValues(t, float64(10), distinctSum1)

	distinctSum2, err := userApplication.DistinctSum(context.Background(), qfi2)
	assert.NoError(t, err)
	assert.EqualValues(t, float64(5), distinctSum2)
}

func TestBasicDistinctCount(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	models := []GormUserDto{
		{
			GormUser{
				Id:       "4bf4eb72-6986-4dc1-a272-1af9608a7f93",
				Number:   2,
				Username: "aaa",
			},
		},
		{
			GormUser{
				Id:       "20aa76d2-9d21-497d-a505-4d7f30762444",
				Number:   3,
				Username: "bbb",
			},
		},
		{
			GormUser{
				Id:       "cddfe42e-d3a7-45eb-bbe3-2256bdf5dc3a",
				Number:   5,
				Username: "ccc",
			},
		},
		{
			GormUser{
				Id:       "9c417e99-d42a-441f-9aef-708eccc77704",
				Number:   5,
				Username: "ddd",
			},
		},
		{
			GormUser{
				Id:       "d98c5963-9109-4926-b333-bcc2af2e1bc7",
				Number:   5,
				Username: "eee",
			},
		},
		{
			GormUser{
				Id:       "01c03a98-7e8d-4799-989c-7ac25a7905b0",
				Number:   5,
				Username: "fff",
			},
		},
	}

	for _, model := range models {
		_, err := userApplication.Create(context.Background(), model)
		if err != nil {
			panic(err)
		}
	}

	qfi1 := QueryFieldInfo{
		QueryInfo: QueryInfo{
			Queries: []BasicQuery{},
		},
		Field: "number",
	}

	qfi2 := QueryFieldInfo{
		QueryInfo: QueryInfo{
			Queries: []BasicQuery{
				{
					Name:    "username",
					Op:      misc.QueryOperatorMoreThan,
					Operand: misc.NewOperand(models[2].Username),
				},
			},
		},
		Field: "number",
	}

	distinctCount1, err := userApplication.DistinctCount(context.Background(), qfi1)
	assert.NoError(t, err)
	assert.EqualValues(t, float64(3), distinctCount1)

	distinctCount2, err := userApplication.DistinctCount(context.Background(), qfi2)
	assert.NoError(t, err)
	assert.EqualValues(t, float64(1), distinctCount2)
}

func TestBasicUpdateField(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	models := []GormUserDto{
		{
			GormUser{
				Id:       "20aa76d2-9d21-497d-a505-4d7f30762444",
				Number:   3,
				FNumber:  3.0,
				Username: "bbb",
			},
		},
		{
			GormUser{
				Id:       "cddfe42e-d3a7-45eb-bbe3-2256bdf5dc3a",
				Number:   5,
				FNumber:  5.5,
				Username: "ccc",
			},
		},
	}

	for _, model := range models {
		_, err := userApplication.Create(context.Background(), model)
		if err != nil {
			panic(err)
		}
	}

	intendedId := "cddfe42e-d3a7-45eb-bbe3-2256bdf5dc3a"
	intendedUserName := "ccc"
	newUsername := "mojtaba"

	// get username value before update
	currentEntity, err := userApplication.GetById(context.Background(), intendedId)
	if err != nil {
		panic(err)
	}

	assert.EqualValues(t, currentEntity.Username, intendedUserName)

	// now do the update, i.e, set username to mojtaba
	queryFieldValueInfo := QueryFieldValueInfo{
		Query: NewIdQueryInfo(intendedId),
		Name:  "username",
		Value: newUsername,
	}
	err = userApplication.UpdateField(context.Background(), queryFieldValueInfo)
	assert.NoError(t, err)

	// Verify that the update was successful by retrieving the entity and checking the username field
	insertedEntity, err := userApplication.GetById(context.Background(), intendedId)

	assert.NoError(t, err)
	assert.EqualValues(t, newUsername, insertedEntity.Username)
}

func TestBasicGetWithSortPage(t *testing.T) {
	db := DbController()

	userApplication := NewGormBasicAutoCrud[GormUser, GormUserDto](db, true)

	for i := 0; i < 100; i++ {

		model := GormUserDto{}

		model.Number = int64(i)
		model.Username = strconv.Itoa(i)

		_, err := userApplication.Create(context.Background(), model)
		if err != nil {
			panic(err)
		}

	}

	insertedEntities, err := userApplication.GetPage(context.Background(), QueryPagination{Pagination: Pagination{Size: 100, Page: 0}, Sort: []misc.Sort{misc.NewSort("number", false, 0)}})
	if err != nil {
		panic(err)
	}

	assert.EqualValues(t, 100, len(insertedEntities.Data))

	assert.EqualValues(t, 99, insertedEntities.Data[0].Number)
}
