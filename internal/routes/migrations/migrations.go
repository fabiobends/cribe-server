package migrations

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"cribeapp.com/cribe-server/internal/utils"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	repo := NewMigrationRepository()

	workingDir, _ := os.Getwd()

	migrationsPath := ""
	if os.Getenv("APP_ENV") == "test" {
		migrationsPath = filepath.Join(workingDir, "..", "..", "..", "infra", "migrations")
	} else {
		migrationsPath = filepath.Join(workingDir, "infra", "migrations")
	}

	filesReader := func() ([]MigrationFile, error) {
		files, err := os.ReadDir(migrationsPath)

		migrations := []MigrationFile{}
		for _, file := range files {
			migrations = append(migrations, file)
		}

		return migrations, err
	}

	migrationsManager := func() (MigrationExecutor, error) {
		return migrate.New("file://"+migrationsPath, os.Getenv("DATABASE_URL"))
	}

	service := NewMigrationService(MigrationService{repo: *repo, filesReader: filesReader, getCurrentTime: time.Now, migrationsManager: migrationsManager})
	handler := NewMigrationHandler(service)

	if r.Method == http.MethodGet {
		migrations := handler.GetMigrations()
		utils.EncodeResponse(w, http.StatusOK, migrations)
		return
	}

	if r.Method == http.MethodPost {
		migrations := handler.PostMigrations()
		utils.EncodeResponse(w, http.StatusCreated, migrations)
		return
	}

	utils.NotAllowed(w)
}
