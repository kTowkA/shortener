package app

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/kTowkA/shortener/internal/config"
	"github.com/kTowkA/shortener/internal/storage/memory"
)

func Example() {
	// создаем экземпляр сервера
	server, err := NewServer(config.DefaultConfig, slog.Default())
	if err != nil {
		log.Fatal(err)
	}

	// создаем экземпляр хранилища.
	// в примере не указываем файл - хранилище будет расположено в памяти
	st, err := memory.NewStorage("")
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	// установим контекст отмены в несколько секунд
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// запускаем сервер с заданным контексом отмены и хранилищем
	if err = server.Run(ctx, st); err != nil {
		log.Fatal(err)
	}
}
