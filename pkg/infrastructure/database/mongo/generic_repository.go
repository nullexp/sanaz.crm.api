package mongo

import (
	"context"
	"reflect"
	"time"

	"github.com/google/uuid"
	database "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository[T database.EntityBased] struct {
	Collection *mongo.Collection
}

func NewRepository[T database.EntityBased](getter database.DataContextGetter, collectionName string) Repository[T] {
	mdb, ok := getter.GetDataContext().(*mongo.Database)

	if !ok {
		panic("unknown context")
	}

	collection := mdb.Collection(collectionName)

	return Repository[T]{Collection: collection}
}

func (r *Repository[T]) Create(ctx context.Context, doc *T) error {
	timestamp := time.Now()

	docVal := reflect.ValueOf(doc).Elem()

	createdAtField := docVal.FieldByName("CreatedAt")

	updatedAtField := docVal.FieldByName("UpdatedAt")

	docVal.FieldByName("Id").Set(reflect.ValueOf(uuid.New().String()))

	if createdAtField.IsValid() && createdAtField.Type() == reflect.TypeOf(time.Time{}) {
		createdAtField.Set(reflect.ValueOf(timestamp))
	}

	if updatedAtField.IsValid() && updatedAtField.Type() == reflect.TypeOf(time.Time{}) {
		updatedAtField.Set(reflect.ValueOf(timestamp))
	}

	_, err := r.Collection.InsertOne(ctx, doc)

	return err
}

func (r Repository[T]) Read(ctx context.Context, filter DynamicFilter) ([]T, error) {
	cur, err := r.Collection.Find(ctx, filter.ToBSON())
	if err != nil {
		return nil, err
	}

	defer cur.Close(ctx)

	var results []T

	for cur.Next(ctx) {

		var doc T

		err := cur.Decode(&doc)
		if err != nil {
			return nil, err
		}

		results = append(results, doc)

	}

	return results, nil
}

func (r Repository[T]) Update(ctx context.Context, filter DynamicFilter, update UpdateField) error {
	update.Fields["$set"] = bson.M{"updated_at": time.Now()}

	_, err := r.Collection.UpdateMany(ctx, filter.ToBSON(), update.ToBSON())

	return err
}

func (r Repository[T]) Delete(ctx context.Context, filter DynamicFilter) error {
	_, err := r.Collection.DeleteMany(ctx, filter.ToBSON())

	return err
}
