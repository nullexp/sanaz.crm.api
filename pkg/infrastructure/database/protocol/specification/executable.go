package specification

import "github.com/nullexp/sanaz.crm.api/pkg/infrastructure/database/protocol"

type Executable interface {
	Execute(protocol.Query) any
	And(left, right Specification) any
	Or(left, right Specification) any
	Not(spec Specification) any
	GroupAnd(left, right Specification) any
	GroupOr(left, right Specification) any
}
