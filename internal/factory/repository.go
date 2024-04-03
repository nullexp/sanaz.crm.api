package factory

// import (
// 	gormRepo "git.omidgolestani.ir/clinic/crm.api/internal/module/chat/persistence/repository/gorm"
// 	repository "git.omidgolestani.ir/clinic/crm.api/pkg/module/chat/persistence/repository"
// )

// func NewUserRepository(name string, params ...any) repository.UserFactory {

// 	if name == "" {
// 		name = Test
// 	}

// 	switch name {
// 	case Test:
// 		fallthrough
// 	case Data:

// 		if len(params) == 0 {
// 			return gormRepo.NewGormRepoFactory(false)
// 		}
// 		val, _ := params[0].(bool)
// 		return gormRepo.NewGormRepoFactory(val)

// 	}
// 	panic(ErrNotImplemented)

// }
