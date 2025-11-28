package repository

//
//type AppRepository interface {
//	GetByName(ctx context.Context, name string) (*model.App, error)
//	GetByKey(ctx context.Context, key string) (*model.App, error)
//}
//
//func NewAppRepository(db ...string) AppRepository {
//	return &gorm.GormApp{}
//}
//
//type AppDataRepository interface {
//	GetById(ctx context.Context, appId uint64, module string, id uint64) (*model.AppData, error)
//	Get(ctx context.Context, appId uint64, module string, params gowk.M) ([]*model.AppData, error)
//	Save(ctx context.Context, appId uint64, table string, data gowk.M) (uint64, error)
//	Delete(ctx context.Context, appId uint64, table string, id uint64) error
//}
//
//func NewAppDataRepository(db ...string) AppDataRepository {
//	return &gorm.GormAppData{}
//}
