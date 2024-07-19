package migrations

import (
	"embed"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var fs embed.FS

// MigrationsUP проведение начальной инициализации БД типа postgres.
// connString строка-подключение.
// возвращает возможную ошибку или nil
func MigrationsUP(connString string) error {
	log.Println(0)
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return fmt.Errorf("создание драйвера для считывания миграций. %w", err)
	}
	log.Println(111)
	connString = strings.TrimPrefix(connString, "postgres://")
	connString = strings.TrimPrefix(connString, "postgresql://")
	connString = "pgx5://" + connString
	log.Println(connString)
	log.Println(222)
	m, err := migrate.NewWithSourceInstance("iofs", d, connString)
	if err != nil {
		return fmt.Errorf("создание экземпляра миграций. %w", err)
	}
	log.Println(333)
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("применение миграций. %w", err)
	}
	log.Println(444)
	return nil
}
