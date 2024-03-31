package gorm

import (
	"context"
	"fmt"

	database "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol"
	dbspec "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/database/protocol/specification"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/misc"
	"gorm.io/gorm"
)

type IntegerRepository[T database.Identity] struct {
	db *gorm.DB

	parser QueryParser
}

func NewGormIntegerRepository[T database.Identity](getter database.DataContextGetter, parser QueryParser) IntegerRepository[T] {
	gdb, ok := getter.GetDataContext().(*gorm.DB)

	if !ok {
		panic("unknown context")
	}

	return IntegerRepository[T]{db: gdb, parser: parser}
}

func NewGormIntegerRepositoryWithDB[T database.Identity](gdb *gorm.DB, parser QueryParser) IntegerRepository[T] {
	return IntegerRepository[T]{db: gdb}
}

func (r IntegerRepository[T]) Get(ctx context.Context, query dbspec.Specification, page misc.Pagination, sort []misc.Sort) (out []T, err error) {
	r.db = r.db.Model(new(T))
	r.db = r.parser.ParseSpecification(r.db, query)
	r.db = r.parser.ParseSort(r.db, sort...)
	r.db = r.parser.ParsePage(r.db, page)
	err = r.db.WithContext(ctx).Find(&out).Error

	return
}

func (r *IntegerRepository[T]) Insert(ctx context.Context, entity *T) (err error) {
	err = r.db.WithContext(ctx).Create(entity).Error

	return
}

func (r IntegerRepository[T]) GetById(ctx context.Context, id int64) (out T, err error) {
	r.db = r.db.Model(new(T))
	r.db = r.parser.ParseSpecification(r.db, dbspec.NewQuerySpecification(misc.Id, misc.QueryOperatorEqual, misc.NewOperand(id)))
	err = r.db.WithContext(ctx).First(&out).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
		return
	}

	return
}

func (r IntegerRepository[T]) Exist(ctx context.Context, query dbspec.Specification) (exist bool, err error) {
	out := new(T)
	r.db = r.db.Model(new(T))
	r.db = r.parser.ParseSpecification(r.db, query)
	err = r.db.WithContext(ctx).Select(misc.Id).First(&out).Error

	if err == gorm.ErrRecordNotFound {
		err = nil
		return
	}

	exist = true
	return
}

func (r IntegerRepository[T]) Count(ctx context.Context, query dbspec.Specification) (count int64, err error) {
	r.db = r.db.Model(new(T))
	r.db = r.parser.ParseSpecification(r.db, query)
	err = r.db.WithContext(ctx).Count(&count).Error

	return
}

func (r IntegerRepository[T]) Update(ctx context.Context, entity *T) (err error) {
	v := *entity

	if v.IsIdEmpty() {
		panic("save require id")
	}

	err = r.db.WithContext(ctx).Save(entity).Error
	return
}

func (r IntegerRepository[T]) PartialUpdate(ctx context.Context, entity database.IdIdentity) (err error) {
	if entity.IsIdEmpty() {
		panic("partial update require id")
	}

	err = r.db.WithContext(ctx).Model(new(T)).Where(misc.Id, entity.GetId()).UpdateColumns(entity).Error
	return
}

func (r IntegerRepository[T]) Delete(ctx context.Context, id int64) (err error) {
	model := new(T)

	r.db = r.db.Model(model)
	r.db = r.parser.ParseSpecification(r.db, dbspec.NewQuerySpecification(misc.Id, misc.QueryOperatorEqual, misc.NewOperand(id)))
	err = r.db.WithContext(ctx).Delete(model).Error

	return
}

func (r IntegerRepository[T]) DeleteBySpecification(ctx context.Context, query dbspec.Specification) (err error) {
	model := new(T)

	r.db = r.db.Model(model)
	r.db = r.parser.ParseSpecification(r.db, query)
	err = r.db.WithContext(ctx).Delete(model).Error

	return
}

func (r IntegerRepository[T]) GetSingle(ctx context.Context, query dbspec.Specification) (out T, err error) {
	r.db = r.db.Model(new(T))

	r.db = r.parser.ParseSpecification(r.db, query)
	err = r.db.WithContext(ctx).First(&out).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
		return
	}

	return
}

func (r IntegerRepository[T]) Sum(ctx context.Context, column string, query dbspec.Specification) (sum float64, err error) {
	// Note: when we query a data and the query return empty result set,
	// if we do not provide gorm with pointer, it will return error
	var tempSum *float64
	r.db = r.db.Model(new(T))
	sumQuery := fmt.Sprintf("SUM(%s)", column)
	r.db = r.parser.ParseSpecification(r.db, query)
	err = r.db.WithContext(ctx).Select(sumQuery).Scan(&tempSum).Error
	if tempSum != nil {
		sum = *tempSum
	}
	return
}

func (r IntegerRepository[T]) Average(ctx context.Context, column string, query dbspec.Specification) (average float64, err error) {
	// Note: when we query a data and the query return empty result set,
	// if we do not provide gorm with pointer, it will return error
	var tempAverage *float64
	r.db = r.db.Model(new(T))
	averageQuery := fmt.Sprintf("AVG(%s)", column)
	r.db = r.parser.ParseSpecification(r.db, query)
	err = r.db.WithContext(ctx).Select(averageQuery).Scan(&tempAverage).Error
	if tempAverage != nil {
		average = *tempAverage
	}
	return
}

func (r *IntegerRepository[T]) SetDB(db *gorm.DB) {
	r.db = db
}

func (r IntegerRepository[T]) DistinctSum(ctx context.Context, column string, query dbspec.Specification) (distinctSum float64, err error) {
	// Note: when we query a data and the query return empty result set,
	// if we do not provide gorm with pointer, it will return error
	var tempSum *float64
	r.db = r.db.Model(new(T))
	sumQuery := fmt.Sprintf("Sum(DISTINCT %s)", column)
	r.db = r.parser.ParseSpecification(r.db, query)
	err = r.db.WithContext(ctx).Select(sumQuery).Scan(&tempSum).Error
	if tempSum != nil {
		distinctSum = *tempSum
	}
	return
}

func (r IntegerRepository[T]) DistinctCount(ctx context.Context, column string, query dbspec.Specification) (distinctCount int64, err error) {
	r.db = r.db.Model(new(T))
	r.db = r.parser.ParseSpecification(r.db, query)
	err = r.db.WithContext(ctx).Distinct(column).Count(&distinctCount).Error
	return
}

func (r IntegerRepository[T]) UpdateField(ctx context.Context, fieldName string, fieldValue any, query dbspec.Specification) (err error) {
	r.db = r.db.Model(new(T))
	r.db = r.parser.ParseSpecification(r.db, query)
	err = r.db.WithContext(ctx).Update(fieldName, fieldValue).Error
	return
}
