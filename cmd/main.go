package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/duniandewon/madkunyah-transactions-service/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type application struct {
	env *config.Env
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		// if err := app.db.PingContext(ctx); err != nil {
		// 	w.WriteHeader(http.StatusServiceUnavailable)
		// 	w.Write([]byte("Database Down"))
		// 	return
		// }

		w.Write([]byte("System Healthy"))
	})

	return r
}

func (app *application) run(h http.Handler) error {
	srv := &http.Server{
		Handler:      h,
		Addr:         fmt.Sprintf(":%v", app.env.Port),
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Printf("Server started port: %s", app.env.Port)

	return srv.ListenAndServe()
}

func main() {
	api := application{
		env: config.NewEnv(),
	}

	if err := api.run(api.mount()); err != nil {
		log.Fatal(err)
	}
}
