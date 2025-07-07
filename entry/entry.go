package entry

//Called by main, begins the server start up sequence

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	g "forum/global"
	"forum/persistence/database"
	"forum/persistence/populate"
	"forum/server/core/config"
	"forum/server/core/handlers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// server starting sequence
func Start() {

	action, err := handleArguments()

	if err != nil {
		log.Fatal("Error with cli arguments:", err.Error())
	}

	err = g.Initialize()
	if err != nil {
		log.Fatal("Error with global config initialization:", err.Error())
	}

	switch action {
	case "populate":
		fmt.Println("Start populating")
		populate.StartProcedure()
		fmt.Println("Finished populating")
		return
	}

	dbCtx, shutDownDb := context.WithCancel(context.Background())
	g.Configs.Database.Ctx = dbCtx

	log.Println("Starting forum db")

	db, err := database.Open(g.Configs.Database)
	if err != nil {
		log.Fatal(fmt.Println("Error initializing database: ", err))
	}

	// Assign configs to handlers
	handlers.Configuration = g.Configs.Handlers

	server := g.Configs.Server

	server.Handler = handlers.SetHandlers(db)

	useHTTPS := g.Configs.Certifications.UseHTTPS
	certFile := filepath.Join(g.Configs.Certifications.File...)
	certKey := filepath.Join(g.Configs.Certifications.Key...)

	// Use HTTP if there's no SSL keys
	_, err = os.Stat(certFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("SSL certs not found in %s. Running in HTTP.\n", certFile)
			useHTTPS = false
		} else {
			log.Fatal("Something bad happened when looking for certs: ", err.Error())
		}
	}

	if useHTTPS {
		server.TLSConfig = &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
			PreferServerCipherSuites: true,
			InsecureSkipVerify:       true,
		}
	}

	config.InitOAuthConfig(g.Configs.OAuth, g.Configs.Certifications.UseHTTPS)

	go func() {
		if useHTTPS {
			log.Printf("Server running on https://%s\n", server.Addr)
			if err := server.ListenAndServeTLS(certFile, certKey); err != nil && err != http.ErrServerClosed {
				log.Fatalf("ListenAndServeTLS failed: %v", err)
			}
		} else {
			log.Printf("Server running on http://%s\n", server.Addr)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("ListenAndServe() failed: %v", err)
			}
		}
	}()

	//wait here for process termination signal to initiate graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	shutDownDb()
	if g.Configs.Database.Wal.AutoTruncate {
		database.ManualTruncate <- struct{}{}
		database.Wg.Wait()
	}
	db.Close()

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Graceful server Shutdown Failed: %v", err)
	}
	log.Println("Server stopped")
}

// handles launching arguments
func handleArguments() (string, error) {
	userArgs := os.Args[1:]

	if len(userArgs) == 1 {
		if userArgs[0] == "populate" {
			return "populate", nil
		}
		return "", errors.New("unknown argument given")
	}

	if len(userArgs) > 1 {
		return "", errors.New("too many arguments given, currently only supporting 1, which is 'populate'")
	}

	return "", nil
}
