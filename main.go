package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ahmadmuzakkir/scrapenews/store/boltdb"

	"github.com/ahmadmuzakkir/scrapenews/api"
	"github.com/ahmadmuzakkir/scrapenews/store"
	"github.com/ahmadmuzakkir/scrapenews/store/mysql"
	"github.com/ahmadmuzakkir/scrapenews/store/sqlite"
	"github.com/getsentry/raven-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/kelseyhightower/envconfig"
	"github.com/robfig/cron"
)

var authorized map[string]struct{}
var env Env

type Env struct {
	SentryDsn     string   `envconfig:"SENTRY_DSN"`
	Port          int      `envconfig:"PORT" required:"true"`
	ApiKeys       []string `envconfig:"API_KEYS"`
	Database      string   `envconfig:"DATABASE"`
	MysqlAddress  string   `envconfig:"MYSQL_ADDRESS"`
	MysqlUsername string   `envconfig:"MYSQL_USERNAME"`
	MysqlPassword string   `envconfig:"MYSQL_PASSWORD"`
	MysqlDatabase string   `envconfig:"MYSQL_DATABASE"`
}

func main() {
	var err error

	env = Env{}
	envconfig.Process("", &env)

	if env.Port == 0 {
		panic("Port cannot be empty")
	}

	log.Println("Port: ", env.Port)
	log.Println("Raven: ", env.SentryDsn)

	if env.SentryDsn != "" {
		raven.SetDSN(env.SentryDsn)
	}

	setAuthorization(env.ApiKeys)

	newsStore, err := getStore()
	if err != nil {
		log.Fatalf("failed to init data store: %s", err)
	}

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 30 * time.Second,
	}
	hc := &http.Client{
		Transport: netTransport,
		Timeout:   time.Second * 30,
	}

	newsRefresher := store.NewRefresher(hc, newsStore)
	go newsRefresher.Refresh()

	newsApi := api.NewNewsHandler(newsStore)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(recoverer)
	r.Use(middleware.DefaultCompress)
	r.Use(authorization)
	r.Mount("/", newsApi.Routes())

	httpServer := &http.Server{Addr: ":" + strconv.Itoa(env.Port), Handler: r}

	go func() {
		// Run the Http server
		log.Fatal(httpServer.ListenAndServe())
	}()

	var mycron *cron.Cron
	raven.CapturePanicAndWait(func() {
		mycron = startCron(newsRefresher)
	}, map[string]string{"module": "cron"})

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)

	select {
	case <-shutdownSignal:
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err = httpServer.Shutdown(ctx)
	if err != nil {
		log.Println("Error shutting down HTTP server, ", err)
	}
}

func recoverer(next http.Handler) http.Handler {
	fn := raven.RecoveryHandler(next.ServeHTTP)

	return http.HandlerFunc(fn)
}

func setAuthorization(apiKeys []string) {
	if apiKeys == nil {
		authorized = nil
		return
	}

	authorized = make(map[string]struct{})
	for _, v := range apiKeys {
		log.Println("Api Key: ", v)
		authorized[v] = struct{}{}
	}
}

func authorization(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")

		if authorized == nil {
			next.ServeHTTP(w, r)
			return
		}

		_, exist := authorized[auth]
		if auth != "" && exist {
			next.ServeHTTP(w, r)
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
	}

	return http.HandlerFunc(fn)
}

func startCron(refresher *store.Refresher) *cron.Cron {
	loc, err := time.LoadLocation("Asia/Kuala_Lumpur")
	if err != nil {
		log.Panic(err)
	}
	c := cron.NewWithLocation(loc)

	c.AddFunc("0 0 7 * * *", func() {
		go refresher.Refresh()
	})

	c.AddFunc("0 0 10 * * *", func() {
		go refresher.Refresh()
	})

	c.AddFunc("0 0 13 * * *", func() {
		go refresher.Refresh()
	})

	c.AddFunc("0 0 16 * * *", func() {
		go refresher.Refresh()
	})

	c.AddFunc("0 0 19 * * *", func() {
		go refresher.Refresh()
	})

	c.AddFunc("0 0 21 * * *", func() {
		go refresher.Refresh()
	})

	c.AddFunc("0 0 1 * * *", func() {
		go refresher.Refresh()
	})

	c.Start()

	return c
}

func getStore() (store.NewsStore, error) {
	switch env.Database {
	case "mysql":
		return mysql.NewStore(env.MysqlAddress, env.MysqlUsername, env.MysqlPassword, env.MysqlDatabase)
	case "sqlite":
		return sqlite.NewStore()
	case "boltdb":
		return boltdb.NewStore()
	}

	return nil, fmt.Errorf("unknown store type: %s", env.Database)
}
