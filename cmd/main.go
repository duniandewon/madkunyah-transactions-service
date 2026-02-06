package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/duniandewon/madkunyah-transactions-service/api"
	"github.com/duniandewon/madkunyah-transactions-service/internal/config"
	"github.com/duniandewon/madkunyah-transactions-service/internal/features/orders"
	mw "github.com/duniandewon/madkunyah-transactions-service/internal/middleware"
	postgresql "github.com/duniandewon/madkunyah-transactions-service/internal/platform/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type application struct {
	env *config.Env
	db  *sql.DB
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := app.db.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Database Down"))
			return
		}

		w.Write([]byte("System Healthy"))
	})

	orderRepo := orders.NewService(app.db)
	menuClient := orders.NewMenuClient("http://localhost:5001")
	orderHandler := api.NewOrderHandler(orderRepo, menuClient)

	r.Route("/orders", func(r chi.Router) {
		r.Post("/", orderHandler.CreateOrderHandler)

		r.Group(func(r chi.Router) {
			r.Use(mw.IsAuth(app.env.JwtSecret))

			r.Get("/", orderHandler.GetAllOrdersHandler)
			r.Get("/{id}", orderHandler.GetUserOrderDetailsHandler)
		})
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
	env := config.NewEnv()

	db, err := postgresql.NewDatabase(env.DatabaseUrl)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	api := application{
		env: env,
		db:  db,
	}

	if err := api.run(api.mount()); err != nil {
		log.Fatal(err)
	}
}
