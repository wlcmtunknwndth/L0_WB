package main

import (
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/stan.go"
	"github.com/wlcmtunknwndth/L0_WB/internal/cacher"
	"github.com/wlcmtunknwndth/L0_WB/internal/config"
	natsServer "github.com/wlcmtunknwndth/L0_WB/internal/nats-server"
	"github.com/wlcmtunknwndth/L0_WB/internal/storage"
	"github.com/wlcmtunknwndth/L0_WB/internal/storage/postgresql"
	"io"
	"log/slog"
	"net/http"
	"time"
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

	sc := natsServer.New(cfg, db)

	cach := cacher.New(db, 1*time.Minute, 3*time.Minute)
	err = cach.Restore()
	if err = cach.Restore(); err != nil {
		slog.Error("couldn't restore cache: ", err)
	} else {
		slog.Info("cache successfully restored")
	}
	ticker := time.NewTicker(20 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := cach.SaveCache(); err != nil {
					continue
				}
				slog.Info("made a cache backup")
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	defer close(quit)

	//defer func(cach *cacher.Cacher) {
	//	err := cach.SaveCache()
	//	if err != nil {
	//		slog.Error("couldn't backup cache: ", err)
	//	}
	//}(cach)

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

		//cache
		var order storage.Order
		if err = json.Unmarshal(body, &order); err != nil {
			slog.Error("couldn't unmarshall order: ", err)
		}
		cach.CacheOrder(order)

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

		order := storage.RandomOrder(uuid)
		orderBytes, err := json.Marshal(order)
		if err != nil {
			slog.Error("couldn't encode random order: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		//cache
		cach.CacheOrder(*order)

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
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}(r.Body)

		var searchReq storage.SearchRequest
		if err = json.Unmarshal(req, &searchReq); err != nil {
			slog.Error("couldn't unmarshall search request: ", err)
			return
		}

		if order, found := cach.GetOrder(searchReq.Uuid); found {
			err = SendOrderAsJson(order, w)
			if err != nil {
				slog.Error("couldn't send cached back: ", err)
			} else {
				slog.Info("sent cached order")
			}
		}

		ch := make(chan bool)
		sub, err := sc.OrderGetter(searchReq.Uuid, w, &ch, cach)
		if err != nil {
			slog.Error("couldn't run receiver: ", err)
			return
		}
		defer func(sub stan.Subscription) {
			if err := sub.Close(); err != nil {
				slog.Error("couldn't close connection: ", err)
				return
			}
		}(sub)

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
	//slog.Error("application finished")
}

func SendOrderAsJson(order *storage.Order, w http.ResponseWriter) error {
	answer, err := json.Marshal(*order)
	if err != nil {
		slog.Error("couldn't marshal order: ", err)
		return err
	}

	if _, err = w.Write(answer); err != nil {
		slog.Error("couldn't send answer: ", err)
		return err
	}
	return nil
}
