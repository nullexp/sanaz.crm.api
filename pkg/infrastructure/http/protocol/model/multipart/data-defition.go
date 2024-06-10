package multipart

import (
	"context"

	http "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/http/protocol"
)

type DataDefinition struct {
	Name     string
	Object   http.Verifier
	Optional bool
	Single   bool
}

func (f *DataDefinition) GetPartName() string {
	return f.Name
}

func (f *DataDefinition) GetSupportedTypes() []string {
	return []string{}
}

func (f *DataDefinition) GetObject() interface{} {
	return f.Object
}

func (f *DataDefinition) IsSingle() bool {
	return f.Single
}

func (f *DataDefinition) IsOptional() bool {
	return f.Optional
}

func (f *DataDefinition) Verify(context.Context) error {
	return f.Object.Verify(context.Background())
}

const UnknownData = "unknown data"
