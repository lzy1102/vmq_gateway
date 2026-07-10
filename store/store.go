package store

import (
	"context"
	"fmt"

	"github.com/lzy1102/vmq_gateway/config"
	gormstore "github.com/lzy1102/vmq_gateway/store/gorm"
	mongostore "github.com/lzy1102/vmq_gateway/store/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	gormdriver "gorm.io/gorm"
)

func Init(ctx context.Context) error {
	switch config.DBDriver() {
	case "mongo":
		return initMongo(ctx)
	case "sqlite", "":
		return initGORM(sqlite.Open(config.DBPath()))
	case "mysql":
		return initGORM(mysql.Open(config.MySQLDSN()))
	case "postgres", "postgresql":
		return initGORM(postgres.Open(config.PostgresDSN()))
	default:
		return fmt.Errorf("unsupported db driver: %s", config.DBDriver())
	}
}

func initMongo(ctx context.Context) error {
	client, err := mongo.Connect(options.Client().ApplyURI(config.MongoURI()))
	if err != nil {
		return err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}
	DBInstance = mongostore.New(client, config.DBName)
	return nil
}

func initGORM(dialector gormdriver.Dialector) error {
	db, err := gormdriver.Open(dialector, &gormdriver.Config{})
	if err != nil {
		return err
	}
	if err := db.AutoMigrate(
		&gormstore.GormOrder{},
		&gormstore.GormUser{},
		&gormstore.GormDevice{},
		&gormstore.GormPool{},
		&gormstore.GormPoolDevice{},
		&gormstore.GormBinding{},
	); err != nil {
		return err
	}
	DBInstance = gormstore.New(db)
	return nil
}
