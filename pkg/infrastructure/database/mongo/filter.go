package mongo

import "go.mongodb.org/mongo-driver/bson"

type DynamicFilter struct {
	Fields map[string]interface{}
}

func (s *DynamicFilter) ToBSON() bson.M {
	query := bson.M{}
	for field, value := range s.Fields {
		query[field] = value
	}
	return query
}

type UpdateField struct {
	Fields map[string]interface{}
}

func (s *UpdateField) ToBSON() bson.M {
	query := bson.M{}
	for field, value := range s.Fields {
		query[field] = value
	}
	return query
}
