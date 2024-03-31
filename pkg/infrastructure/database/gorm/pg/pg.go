package pg

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Config struct {
	Host string

	Port int

	Username string

	Password string

	Name string

	Driver string

	Schema string

	IgnorePermissionDenied bool
}

const Public = "public"

func (config *Config) DSN() string {
	if config.Schema == "" {
		config.Schema = Public
	}

	return fmt.Sprintf("host=%s  port=%d user=%s password=%s dbname=%s sslmode=disable",

		config.Host, config.Port, config.Username, config.Password, config.Name)
}

func (config *Config) RawDSN() string {
	if config.Schema == "" {
		config.Schema = Public
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable",

		config.Host, config.Port, config.Username, config.Password)
}

func (config *Config) GetTableSchema() string {
	if config.Schema == "" {
		config.Schema = Public
	}

	return config.Schema + "."
}

type PgDatabase struct {
	Database *gorm.DB

	DBConfig Config
}

func NewPgDatabase(DBConfig Config) *PgDatabase {
	md := PgDatabase{}

	md.DBConfig = DBConfig

	return &md
}

func (md *PgDatabase) GetGorm() *gorm.DB {
	return md.Database
}

func (md *PgDatabase) open() error {
	db, err := gorm.Open(postgres.Open(md.DBConfig.DSN()), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: md.DBConfig.GetTableSchema(), // schema name

			SingularTable: false,
		},

		SkipDefaultTransaction: true,
	})
	if err != nil {
		return err
	}

	preloadPlugin := &PreloadAllPlugin{}

	err = preloadPlugin.Initialize(db)

	if err != nil {
		return err
	}

	md.Database = db

	return nil
}

func (md *PgDatabase) Open() error {
	return md.open()
}

func (md *PgDatabase) OpenRaw() error {
	db, err := gorm.Open(postgres.Open(md.DBConfig.RawDSN()), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return err
	}

	preloadPlugin := &PreloadAllPlugin{}

	err = preloadPlugin.Initialize(db)

	if err != nil {
		return err
	}

	md.Database = db

	return nil
}
