package multipart

import http "gitlab.espadev.ir/espad-go/infrastructure/http/protocol"

type DataDefinition struct {
	Name     string
	Object   http.Verifier
	Optional bool
	Single   bool
}

func (f *DataDefinition) GetPartName() string {
	return f.Name
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

func (f *DataDefinition) Verify() error {
	return f.Object.Verify()
}

const UnknownData = "unknown data"
