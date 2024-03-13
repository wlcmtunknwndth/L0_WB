package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/stan.go"
	"github.com/wlcmtunknwndth/L0_WB/internal/config"
	"github.com/wlcmtunknwndth/L0_WB/internal/storage"
	"github.com/wlcmtunknwndth/L0_WB/internal/storage/postgresql"
	"io"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	cfg := config.MustLoad()
	fmt.Println(cfg)
	db, err := postgresql.New(cfg.DbConfig)
	if err != nil {
		slog.Error("couldn't open db:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		slog.Error("couldn't ping db", err)
	} else {
		slog.Info("pinged db successfully")
	}

	sc, err := stan.Connect(
		"test-cluster",
		"client",
		stan.Pings(1, 3),
		stan.NatsURL(cfg.Nats.IpAddr),
	)
	if err != nil {
		slog.Error("couldn't run nats server")
	}
	defer sc.Close()

	sub, err := sc.Subscribe("order", func(m *stan.Msg) {
		var order storage.Order
		err := json.Unmarshal(m.Data, &order)
		if err != nil {
			slog.Error("couldn't unmarshall req:", err)
			return
		}
		//fmt.Printf("%+v", order)
		err = db.SaveOrder(&order)
		if err != nil {
			slog.Error("couldn't save order: ", err)
			return
		}
	})
	if err != nil {
		slog.Error("couldn't create a subscriber: ", err)
		os.Exit(1)
	}
	defer sub.Unsubscribe()

	//uuid := gofakeit.UUID()
	//orderBytes, err := json.Marshal(storage.RandomOrder(uuid))
	//if err != nil {
	//	slog.Error("err:", err)
	//}
	//if err = sc.Publish("order", orderBytes); err != nil {
	//	slog.Error("Error: ", err)
	//}
	//time.Sleep(time.Second)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(middleware.Logger)

	router.Post("/create", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		order, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("error decoding request: ", err)
		}

		//var order storage.Order
		//err = json.Unmarshal(b, &order)

		if err := sc.Publish("order", order); err != nil {
			slog.Error("Failed to save order:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write([]byte("Saved succesfully"))

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

//res, err := db.GetOrder("b563feb7b2b84b6test")
//if err != nil {
//	slog.Error("Couldn't get res: ", err)
//	os.Exit(1)
//}

//jsonRes, err := json.Marshal(res)
//if err != nil {
//	slog.Error("Couldn't marshal res to json: ", err)
//	os.Exit(1)
//}

//uuid := gofakeit.UUID()
//order := storage.RandomOrder(uuid)
//fmt.Println(order)
//err = db.SaveOrder(order)
//if err != nil {
//	slog.Error("couldn't save random:", err)
//	os.Exit(1)
//}
//
//res, err := db.GetOrder(uuid)
//if err != nil {
//	slog.Error("random save didnt work:", err)
//	os.Exit(1)
//}
//
//fmt.Println(res)
