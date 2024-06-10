package gorm

import (
	dbspec "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol/specification"
	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/misc"
	"gorm.io/gorm"
)

type (
	SpecificationParser interface {
		ParseSpecification(db *gorm.DB, spec dbspec.Specification) *gorm.DB
	}

	SortParser interface {
		ParseSort(db *gorm.DB, sorts ...misc.Sort) *gorm.DB
	}

	PageParser interface {
		ParsePage(db *gorm.DB, page misc.Pagination) *gorm.DB
	}

	QueryParser interface {
		SpecificationParser

		SortParser

		PageParser
	}
)
