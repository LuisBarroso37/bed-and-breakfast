package dbrepository

import (
	"database/sql"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/config"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/repository"
)

type postgresDBRepository struct {
	App *config.AppConfig
	DB *sql.DB
}

type testDBRepository struct {
	App *config.AppConfig
	DB *sql.DB
}

func NewPostgresRepository(conn_pool *sql.DB, app *config.AppConfig) repository.DatabaseRepository {
	return &postgresDBRepository{
		App: app,
		DB: conn_pool,
	}
}

func NewTestRepository(app *config.AppConfig) repository.DatabaseRepository {
	return &testDBRepository{
		App: app,
	}
}