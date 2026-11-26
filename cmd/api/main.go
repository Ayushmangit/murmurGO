package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ayushmangit/systemDesignGO/internal/feed"
	"github.com/Ayushmangit/systemDesignGO/internal/platform/postgres"
	"github.com/Ayushmangit/systemDesignGO/internal/posts"
	"github.com/Ayushmangit/systemDesignGO/internal/users"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("murmur stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return errors.New("DATABASE_URL is not set")
	}
	addr := cmp.Or(os.Getenv("ADDR"), ":8081")

	pool, err := postgres.Connect(ctx, url)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pool.Close()
	if err := postgres.Migrate(url); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	mux := http.NewServeMux()

	u := users.New(pool)
	mux.HandleFunc("POST /users", u.Create)
	mux.HandleFunc("POST /users/{handle}/follow", u.Auth(u.Follow))
	mux.HandleFunc("DELETE /users/{handle}/unfollow", u.Auth(u.UnFollow))

	p := posts.New(pool)
	mux.HandleFunc("POST /posts", u.Auth(p.Create))
	mux.HandleFunc("GET /users/{handle}/posts", p.List)

	f := feed.New(pool)
	mux.HandleFunc("GET /feed", u.Auth(f.List))

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       1 * time.Minute,
	}

	errc := make(chan error, 1)
	go func() {
		errc <- srv.ListenAndServe()
	}()
	slog.Info("listenting", "addr", addr)

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		stop()
	}
	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return nil
}
