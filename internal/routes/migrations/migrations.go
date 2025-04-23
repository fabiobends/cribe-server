package migrations

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"cribeapp.com/cribe-server/internal/utils"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func getMigrationsPath() (string, error) {
	// Get the path of the current file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", os.ErrNotExist
	}

	// Get the project root directory by going up from the current file
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))

	// Construct the migrations path relative to the project root
	return filepath.Join(projectRoot, "..", "infra", "migrations"), nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	repo := NewMigrationRepository()

	migrationsPath, err := getMigrationsPath()
	if err != nil {
		utils.EncodeResponse(w, http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to get migrations path",
		})
		return
	}

	filesReader := func() ([]MigrationFile, error) {
		files, err := os.ReadDir(migrationsPath)
		if err != nil {
			return nil, err
		}

		migrations := []MigrationFile{}
		for _, file := range files {
			migrations = append(migrations, file)
		}

		return migrations, nil
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
		if len(migrations) > 0 {
			utils.EncodeResponse(w, http.StatusCreated, migrations)
		} else {
			utils.EncodeResponse(w, http.StatusOK, migrations)
		}
		return
	}

	utils.NotAllowed(w)
}
