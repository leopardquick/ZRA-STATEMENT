// @title			Swagger Azam Super Agent API
// @version		1.0
// @description	This is a sample Azam Super Agent server.
// @termsOfService	http://swagger.io/terms/
// @contact.name	name: nadir , email: nadir.hemed@pbzbank.co.tz
// @license.name	Apache 2.0, url: http://www.apache.org/licenses/LICENSE-2.0.html
// @host			localhost:2507
// @BasePath		/
// @schemes		http
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/leopardquict/tra-statement/constant"
	"github.com/leopardquict/tra-statement/handler"
	"github.com/robfig/cron/v3"
)

func main() {

	file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)

	if err != nil {
		fmt.Println("Error creating log file")
		return
	}

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of the major browsers
	})

	ll := slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{}))

	handler := handler.NewHandler(ll)

	r := chi.NewRouter()

	c := cron.New()

	c.AddFunc("10 0 * * *", func() {

		currentTime := time.Now()

		// Subtract 24 hours to get yesterday's time
		yesterday := currentTime.Add(-24 * time.Hour)

		ll.Info("Running cron job to generate statement for " + yesterday.Format("02012006") + " for all banks accounts")

		for _, account := range constant.BANKS_ACCOUNT {
			for i := 0; i < 3; i++ {
				_, err := handler.GetStatement(yesterday.Format("02012006"), account)
				if err == nil {
					// Job succeeded, break out of the retry loop
					ll.Info("Statement", "response", "statement generated successfully for "+account+" on "+yesterday.Format("02012006"))
					break
				} else {
					// Job failed, log the error and retry after a delay
					fmt.Printf("Error: %v. Retrying...\n", err)
					time.Sleep(time.Second * 10) // Add a delay before retrying
				}
			}
			time.Sleep(time.Second * 5)
		}

		// Try the job up to 3 times on failure

	})

	c.Start()

	r.Use(middleware.RequestID)
	r.Use(corsMiddleware.Handler)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.DefaultLogger)
	r.Use(middleware.NoCache)
	r.Use(middleware.Timeout(60 * time.Second))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		ll.Error("404 page not found")
		w.Write([]byte("404 page not found"))
	})

	// r.Get("/statement", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "application/xml")

	// 	day := time.Now().Format("02012006")
	// 	rgs, err := handler.GetStatement(day, "0754716001")

	// 	if err != nil {
	// 		ll.Error("Error getting statement", "error", err)
	// 		w.WriteHeader(http.StatusInternalServerError)
	// 		w.Write([]byte("Error getting statement"))
	// 		return
	// 	}

	// 	w.Header().Set("Content-Type", "application/xml")
	// 	w.WriteHeader(http.StatusOK)
	// 	w.Write(rgs)
	// })

	r.Post("/statement", handler.StatementRequest)

	// r.Get("/statementcbs", func(w http.ResponseWriter, r *http.Request) {
	// 	statement, err := handler.GetStatementFromCbs("0404204000", "03022023", "PBZSTM"+time.Now().Format("20060102150405"))

	// 	if err != nil {
	// 		ll.Error("Error getting statement", "error", err)
	// 		w.WriteHeader(http.StatusInternalServerError)
	// 		w.Write([]byte(err.Error()))
	// 		return
	// 	}

	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.WriteHeader(http.StatusOK)

	// 	json.NewEncoder(w).Encode(statement)

	// })

	server := &http.Server{
		Addr:    ":8989",
		Handler: r,
	}

	// fmt.Println("Server running on port 3031")
	// Use a goroutine to run the server
	go func() {
		fmt.Printf("Server is listening on :%s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error: %v\n", err)
		}
	}()

	// Set up a signal channel to listen for SIGINT and SIGTERM
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	// Block until a signal is received

	<-signalChan
	fmt.Println("Received interrupt signal. Shutting down...")

	// Create a context with a timeout to allow for a graceful shutdown
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	if err := server.Shutdown(timeoutCtx); err != nil {
		fmt.Printf("Error during server shutdown: %v\n", err)
	}

	fmt.Println("Server gracefully shut down.")

}
