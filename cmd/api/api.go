package main

import (
	"flag"
	"fmt"
	"log"
	"myapp/internal/driver"
	"myapp/internal/models"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
	stripe struct {
		secret string
		key    string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
	}
}

type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
	DB       models.DBModel
}

func (app *application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.infoLog.Println(fmt.Sprintf("Starting  BackEnd Server in %s mode on port %d", app.config.env, app.config.port))

	return srv.ListenAndServe()
}

func main() {

	var cfg config
	flag.IntVar(&cfg.port, "port", 4001, "Server port to listen on ")
	flag.StringVar(&cfg.env, "env", "development", "Application Environment {development|production|maintenance}")
	flag.StringVar(&cfg.db.dsn, "dsn", "cpaez:12345678@tcp(localhost:3306)/widgets?parseTime=true&tls=false", "DSN")
	flag.StringVar(&cfg.smtp.host, "smpthost", "sandbox.smtp.mailtrap.io", "smtp host")
	flag.StringVar(&cfg.smtp.username, "smptuser", "f622baa29fef81", "smtp user")
	flag.StringVar(&cfg.smtp.password, "smptpass", "8ac6ea947efcaf", "smtp password")
	flag.IntVar(&cfg.smtp.port, "smtpport", 587, "smtp port")

	flag.Parse()

	log.Println("hello6")

	cfg.stripe.key = os.Getenv("STRIPE_KEY")
	cfg.stripe.secret = os.Getenv("STRIPE_SECRET")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	conn, err := driver.OpenDB(cfg.db.dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer conn.Close()

	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  version,
		DB:       models.DBModel{DB: conn},
	}

	err = app.serve()
	if err != nil {
		log.Fatal(err)
	}

}
