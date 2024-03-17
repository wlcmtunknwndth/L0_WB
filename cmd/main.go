package main

import (
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/stan.go"
	"github.com/wlcmtunknwndth/L0_WB/internal/config"
	nats_server "github.com/wlcmtunknwndth/L0_WB/internal/nats-server"
	"github.com/wlcmtunknwndth/L0_WB/internal/storage"
	"github.com/wlcmtunknwndth/L0_WB/internal/storage/postgresql"
	"io"
	"log/slog"
	"net/http"
)

type Database interface {
	SaveOrder(order *storage.Order) error
	GetOrder(uuid string) (*storage.Order, error)
	Close() error
}

func main() {
	cfg := config.MustLoad()

	db, err := postgresql.New(cfg.DbConfig)
	if err != nil {
		slog.Error("couldn't open db:", err)
	}
	defer func(db Database) {
		err := db.Close()
		if err != nil {
			slog.Error("wasn't able to close db connection: ", err)
		}
	}(db)

	//order, err := db.GetOrder("b563feb7b2b84b6test")
	//if err != nil {
	//	slog.Error("couldn't get order: ", err)
	//	return
	//}
	//fmt.Printf("%+v", order)

	sc := nats_server.New(cfg, db)

	saverSub, err := sc.Saver()
	defer func(sub stan.Subscription) {
		if err := sub.Close(); err != nil {
			slog.Error("couldn't close saver: ", err)
			return
		}
	}(saverSub)
	if err != nil {
		slog.Error("couldn't run saver: ", err)
		return
	}

	getterSub, err := sc.GetHandler()
	if err != nil {
		slog.Error("couldn't start get handler: ", err)
		return
	}
	defer func(sub stan.Subscription) {
		if err := sub.Close(); err != nil {
			slog.Error("couldn't close connection: ", err)
			return
		}
	}(getterSub)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(middleware.Logger)

	router.Post("/save", func(w http.ResponseWriter, r *http.Request) {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}(r.Body)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("error decoding request: ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err = sc.PublishOrder(body); err != nil {
			slog.Error("couldn't publish order: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err = w.Write([]byte("saved")); err != nil {
			slog.Error("couldn't write body: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	router.Post("/save_random", func(w http.ResponseWriter, r *http.Request) {
		uuid := gofakeit.UUID()

		orderBytes, err := json.Marshal(storage.RandomOrder(uuid))
		if err != nil {
			slog.Error("couldn't encode random order: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = sc.PublishOrder(orderBytes); err != nil {
			slog.Error("couldn't publish order: ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if _, err = w.Write([]byte(uuid)); err != nil {
			slog.Error("Couldn't write head")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	router.Get("/get", func(w http.ResponseWriter, r *http.Request) {
		req, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			slog.Error("couldn't get body: ", err)
			return
		}
		var searchReq storage.SearchRequest
		if err = json.Unmarshal(req, &searchReq); err != nil {
			slog.Error("couldn't unmarshall search request: ", err)
			return
		}

		ch := make(chan bool)
		sub, err := sc.OrderGetter(searchReq.Uuid, w, &ch)
		defer func(sub stan.Subscription) {
			if err := sub.Close(); err != nil {
				slog.Error("couldn't close connection: ", err)
				return
			}
		}(sub)
		if err != nil {
			slog.Error("couldn't run receiver: ", err)
			return
		}

		if err = sc.PublishUUID([]byte(searchReq.Uuid)); err != nil {
			slog.Error("couldn't publish uuid: ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		<-ch
	})

	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  cfg.Server.Timeout,
		WriteTimeout: cfg.Server.Timeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		slog.Error("failed to start server")
	}
	slog.Error("application finished")
}
