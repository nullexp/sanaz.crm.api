package application

import (
	"context"

	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/gorm"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/gorm/pg"
	database "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol/specification"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"
)

type Dto interface {
	Validate(context.Context) error

	IsIdEmpty() bool

	database.UuIdIdentity
}

type PartialDto interface{}

type OperationEntity interface {
	database.UuIdIdentity
}

// TODO: provide validation for given dto

type BasicCrud[T OperationEntity, J Dto] interface {
	Create(context.Context, J) (string, error)

	Update(context.Context, J) error // Does not support nested update

	GetById(context.Context, string) (J, error)

	PartialUpdate(context.Context, Dto) error // Does not support nested update

	List(context.Context) ([]J, error)

	GetPage(context.Context, QueryPagination) (GetPageResponse[J], error)

	Delete(context.Context, QueryInfo) error

	Set(context.Context, QueryDtoInfo[J]) (string, error) // Support Nested Set

	GetQuery(context.Context, QueryInfo) ([]J, error)

	GetQueryAsMap(context.Context, QueryInfo) (map[string]J, error)

	GetSingleQuery(context.Context, QueryInfo) (J, error)

	Exist(context.Context, QueryInfo) (bool, error)

	Sum(context.Context, QueryFieldInfo) (out float64, err error)

	Average(context.Context, QueryFieldInfo) (out float64, err error)

	CreateMultiple(context.Context, []J) (ids []string, err error)

	SetMultiple(context.Context, []QueryDtoInfo[J]) (ids []string, err error)

	Count(context.Context, QueryInfo) (out int64, err error)

	// Rate calculates the rate based on the provided query field information.
	// Then, it generates queries for the numerator and denominator based on QueryInfo fields.
	// The sums of the numerator and denominator are calculated using the repository's Sum() method.
	// If the denominator is not zero, the rate is calculated by dividing the numerator by the denominator,
	// multiplying it by 100, and assigning it to the 'out' variable. Otherwise, it sets 'out' to 0.
	// Finally, the calculated rate 'out' and a nil error are returned.
	// Note that it will return in percentage as a float value!
	Rate(context.Context, RateRequest) (out float64, err error)

	// calculates number of distinct values in a field(column) of database  specified by a queryFieldInfo.
	DistinctCount(context.Context, QueryFieldInfo) (out int64, err error)

	// calculates sum of distinct values in a field(column) of database specified by a queryFieldInfo.
	DistinctSum(ctx context.Context, queryFieldInfo QueryFieldInfo) (out float64, err error)

	UpdateField(ctx context.Context, queryFieldValueInfo QueryFieldValueInfo) (err error)
}

type QueryFieldValueInfo struct {
	Query QueryInfo
	Name  string
	Value any
}

func (fs *QueryFieldValueInfo) GetName() string {
	return fs.Name
}

func (fs *QueryFieldValueInfo) GetValue() any {
	return fs.Value
}

type RateRequest struct {
	QueryFieldInfoNumerator   QueryFieldInfo `json:"queryFieldInfoNumerator"`   // Numerator
	QueryFieldInfoDenominator QueryFieldInfo `json:"queryFieldInfoDenominator"` // Denominator
}

type GetPageResponse[J Dto] struct {
	Data []J

	TotalCount int64
}

type Pagination struct {
	Page int64 `json:"page" validate:"gte=0"`

	Size int64 `json:"size" validate:"required,gte=1"`
}

func (p Pagination) GetSkip() uint {
	return uint((p.Page - 1) * p.Size)
}

func (p Pagination) GetLimit() uint {
	return uint(p.Size)
}

type QueryPagination struct {
	Pagination Pagination `json:"pagination"`

	QueryInfo *QueryInfo `json:"queryInfo"`

	Sort []misc.Sort
}

type QueryIdInfo struct {
	QueryInfo *QueryInfo `json:"queryInfo"`

	Id string `json:"id"`
}

type QueryInfo struct {
	Queries []BasicQuery `json:"queries"`

	And bool `json:"and"`

	Sort []misc.Sort
}

func NewIdQueryInfo(id string) QueryInfo {
	return QueryInfo{Queries: []BasicQuery{{Name: misc.Id, Operand: misc.NewOperand(id), Op: misc.QueryOperatorEqual}}}
}

func NewBasicEqualQueryInfo(field string, value string) QueryInfo {
	return QueryInfo{Queries: []BasicQuery{{Name: field, Operand: misc.NewOperand(value), Op: misc.QueryOperatorEqual}}}
}

func (qi QueryInfo) GetQuery() (out []misc.Query) {
	for _, v := range qi.Queries {
		val := v
		out = append(out, &val)
	}

	return out
}

func (qi QueryInfo) IsAnd() bool {
	return qi.And
}

type BasicQuery struct {
	Name string `json:"name"` // Name

	Op misc.QueryOperator `json:"op"` //  Operator

	Operand *misc.Operand `json:"operand"` // Operand
}

func (basicQuery BasicQuery) GetName() string {
	return basicQuery.Name
}

func (basicQuery *BasicQuery) SetName(name string) {
	basicQuery.Name = name
}

func (basicQuery BasicQuery) GetOperator() misc.QueryOperator {
	return basicQuery.Op
}

func (basicQuery BasicQuery) GetOperand() *misc.Operand {
	return basicQuery.Operand
}

type QueryDtoInfo[J Dto] struct {
	QueryInfo QueryInfo `json:"queryInfo"` // if it is nil will set id= Dto.Id by default

	Dto J `json:"dto"`
}

type QueryFieldInfo struct {
	QueryInfo QueryInfo `json:"queryInfo"`

	Field string `json:"string"`
}

func NewGormBasicAutoCrud[T OperationEntity, J Dto](txGetter database.TransactionFactoryGetter, withTestId bool) BasicCrud[T, J] {
	bci := BasicCrudImpl[T, J]{}

	bci.TransactionFactory = txGetter

	bci.WithTestId = withTestId

	return bci
}

type BasicCrudImpl[T OperationEntity, J Dto] struct {
	TransactionFactory database.TransactionFactoryGetter

	WithTestId bool
}

func (p BasicCrudImpl[T, J]) Create(ctx context.Context, dto J) (out string, err error) {
	if err = dto.Validate(ctx); err != nil {
		return // some standard error
	}

	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	if err = repo.Insert(ctx, &dto); err != nil {
		return
	}

	return dto.GetUuid(), tx.Commit()
}

// Update does not support nested fields

func (p BasicCrudImpl[T, J]) Update(ctx context.Context, dto J) (err error) {
	if err = dto.Validate(ctx); err != nil {
		return // some standard error
	}

	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	if err = repo.Update(ctx, &dto); err != nil {
		return
	}

	return tx.Commit()
}

func (p BasicCrudImpl[T, J]) PartialUpdate(ctx context.Context, dto Dto) (err error) {
	if err = dto.Validate(ctx); err != nil {
		return // some standard error
	}

	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	if err = repo.PartialUpdate(ctx, dto); err != nil {
		return
	}

	return tx.Commit()
}

func (p BasicCrudImpl[T, J]) List(ctx context.Context) (out []J, err error) {
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	out, err = repo.Get(ctx, nil, nil, nil)

	return
}

func (p BasicCrudImpl[T, J]) GetPage(ctx context.Context, page QueryPagination) (out GetPageResponse[J], err error) {
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	query := GenerateQuery(page.QueryInfo)

	out.TotalCount, err = repo.Count(ctx, query)

	out.Data, err = repo.Get(ctx, query, page.Pagination, page.Sort)

	return
}

func (p BasicCrudImpl[T, J]) GetQuery(ctx context.Context, queryInfo QueryInfo) (out []J, err error) {
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	query := GenerateQuery(&queryInfo)

	out, err = repo.Get(ctx, query, nil, queryInfo.Sort)

	return
}

func (p BasicCrudImpl[T, J]) GetQueryAsMap(ctx context.Context, queryInfo QueryInfo) (out map[string]J, err error) {
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}
	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)
	query := GenerateQuery(&queryInfo)
	out, err = repo.GetAsMap(ctx, query, nil, queryInfo.Sort)
	return
}

func (p BasicCrudImpl[T, J]) GetSingleQuery(ctx context.Context, queryInfo QueryInfo) (out J, err error) {
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	query := GenerateQuery(&queryInfo)

	out, err = repo.GetSingle(ctx, query)

	return
}

func (p BasicCrudImpl[T, J]) GetById(ctx context.Context, id string) (out J, err error) {
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	out, err = repo.GetById(ctx, id)

	return
}

func (p BasicCrudImpl[T, J]) DeleteById(ctx context.Context, id string) (err error) {
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	err = repo.Delete(ctx, id)

	return tx.Commit()
}

func (p BasicCrudImpl[T, J]) Delete(ctx context.Context, queryInfo QueryInfo) (err error) {
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	var query specification.Specification = GenerateQueryWithId(&queryInfo, "")

	err = repo.DeleteBySpecification(ctx, query)

	return tx.Commit()
}

// By default, if query info does not have any queries, it will remove by id = dto.id

func (p BasicCrudImpl[T, J]) Set(ctx context.Context, queryDto QueryDtoInfo[J]) (out string, err error) {
	if err = queryDto.Dto.Validate(ctx); err != nil {
		return // some standard error
	}

	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	var query specification.Specification = GenerateQueryWithId(&queryDto.QueryInfo, queryDto.Dto.GetUuid())

	if err = repo.DeleteBySpecification(ctx, query); err != nil {
		return
	}

	if err = repo.Insert(ctx, &queryDto.Dto); err != nil {
		return
	}

	return queryDto.Dto.GetUuid(), tx.Commit()
}

func (p BasicCrudImpl[T, J]) Exist(ctx context.Context, queryInfo QueryInfo) (out bool, err error) {
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	query := GenerateQuery(&queryInfo)

	out, err = repo.Exist(ctx, query)

	return
}

func (p BasicCrudImpl[T, J]) Sum(ctx context.Context, queryFieldInfo QueryFieldInfo) (out float64, err error) {
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	query := GenerateQuery(&queryFieldInfo.QueryInfo)

	out, err = repo.Sum(ctx, queryFieldInfo.Field, query)

	return
}

func (p BasicCrudImpl[T, J]) Average(ctx context.Context, queryFieldInfo QueryFieldInfo) (out float64, err error) {
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	query := GenerateQuery(&queryFieldInfo.QueryInfo)

	out, err = repo.Average(ctx, queryFieldInfo.Field, query)

	return
}

func (p BasicCrudImpl[T, J]) Rate(ctx context.Context, queryFieldInfoStruct RateRequest) (out float64, err error) {
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	queryForNumerator := GenerateQuery(&queryFieldInfoStruct.QueryFieldInfoNumerator.QueryInfo)
	queryForDenominator := GenerateQuery(&queryFieldInfoStruct.QueryFieldInfoDenominator.QueryInfo)

	numerator, err := repo.Sum(ctx, queryFieldInfoStruct.QueryFieldInfoNumerator.Field, queryForNumerator)
	if err != nil {
		return
	}
	denominator, err := repo.Sum(ctx, queryFieldInfoStruct.QueryFieldInfoDenominator.Field, queryForDenominator)
	if err != nil {
		return
	}
	if denominator != 0 {
		out = 100 * float64(numerator) / float64(denominator)
	} else {
		out = 0
	}

	return out, nil
}

func (p BasicCrudImpl[T, J]) CreateMultiple(ctx context.Context, dtos []J) (ids []string, err error) {
	for _, dto := range dtos {
		if err = dto.Validate(ctx); err != nil {
			return
		}
	}
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)
	for _, dto := range dtos {
		if err = repo.Insert(ctx, &dto); err != nil {
			return
		}
		ids = append(ids, dto.GetUuid())
	}

	return ids, tx.Commit()
}

func (p BasicCrudImpl[T, J]) SetMultiple(ctx context.Context, queries []QueryDtoInfo[J]) (ids []string, err error) {
	for _, queryDto := range queries {
		if err = queryDto.Dto.Validate(ctx); err != nil {
			return // some standard error
		}
	}
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	for _, queryDto := range queries {
		query := GenerateQueryWithId(&queryDto.QueryInfo, queryDto.Dto.GetUuid())
		if err = repo.DeleteBySpecification(ctx, query); err != nil {
			return
		}

	}

	for _, queryDto := range queries {
		if err = repo.Insert(ctx, &queryDto.Dto); err != nil {
			return
		}
		ids = append(ids, queryDto.Dto.GetUuid())
	}

	return ids, tx.Commit()
}

func (p BasicCrudImpl[T, J]) Count(ctx context.Context, queryInfo QueryInfo) (out int64, err error) {
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()
	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()
	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)
	query := GenerateQuery(&queryInfo)
	out, err = repo.Count(ctx, query)

	return
}

func (p BasicCrudImpl[T, J]) DistinctSum(ctx context.Context, queryFieldInfo QueryFieldInfo) (out float64, err error) {
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	query := GenerateQuery(&queryFieldInfo.QueryInfo)

	out, err = repo.DistinctSum(ctx, queryFieldInfo.Field, query)

	return
}

func (p BasicCrudImpl[T, J]) DistinctCount(ctx context.Context, queryFieldInfo QueryFieldInfo) (out int64, err error) {
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	query := GenerateQuery(&queryFieldInfo.QueryInfo)

	out, err = repo.DistinctCount(ctx, queryFieldInfo.Field, query)

	return
}

func (p BasicCrudImpl[T, J]) UpdateField(ctx context.Context, queryFieldValueInfo QueryFieldValueInfo) (err error) {
	factory, err := p.TransactionFactory.GetTransactionFactory()
	if err != nil {
		return
	}

	tx := factory.New()

	err = tx.Begin()
	if err != nil {
		return
	}

	defer tx.RollbackUnlessCommitted()

	repo := gorm.NewGormUuidCompoundRepository[T, J](tx, pg.NewParser(), database.NewDefaultMapper[T, J](), p.WithTestId, misc.Id)

	query := GenerateQuery(&queryFieldValueInfo.Query)

	err = repo.UpdateField(ctx, queryFieldValueInfo.GetName(), queryFieldValueInfo.GetValue(), query)
	if err != nil {
		return
	}

	return tx.Commit()
}

func GenerateQueryWithId(queryInfo *QueryInfo, id string) (query specification.Specification) {
	if queryInfo == nil || len(queryInfo.Queries) == 0 {
		query = specification.GetIdExistSpecification(id)
	} else {
		dbQuery := database.MapQueryInfo(queryInfo)
		query = specification.ToSpecification(dbQuery)
	}
	return
}

func GenerateQuery(queryInfo *QueryInfo) (query specification.Specification) {
	if queryInfo == nil {
		return nil
	} else {
		dbQuery := database.MapQueryInfo(queryInfo)
		query = specification.ToSpecification(dbQuery)
	}
	return
}
