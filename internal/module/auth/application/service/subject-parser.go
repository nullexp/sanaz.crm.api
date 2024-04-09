package application

import (
	utility "git.omidgolestani.ir/clinic/crm.api/internal/module/auth/utility"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/misc"
)

type JsonSubjectParser struct{}

func NewSubjectParser() misc.SubjectParser {
	return JsonSubjectParser{}
}

func (JsonSubjectParser) MustParseSubject(subject string) misc.Subject {
	return utility.MustGetSubject(subject)
}
