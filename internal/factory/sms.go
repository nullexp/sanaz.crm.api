package factory

// import (
// 	smsImpl "git.omidgolestani.ir/clinic/crm.api/internal/infrastructure/sms"
// 	infra "git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/sms"
// )

// func NewSMSSender(name string, params ...any) infra.SMSSender {

// 	if name == "" {
// 		name = Test
// 	}

// 	switch name {
// 	case Test:
// 		wantError := false
// 		if len(params) != 0 {
// 			wantError = params[0].(bool)
// 		}
// 		return smsImpl.NewSMSMock(wantError)
// 	}
// 	panic(ErrNotImplemented)
// }
