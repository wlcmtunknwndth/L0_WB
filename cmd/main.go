package main

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/wlcmtunknwndth/L0_WB/internal/config"
	"github.com/wlcmtunknwndth/L0_WB/internal/storage"
	"github.com/wlcmtunknwndth/L0_WB/internal/storage/postgresql"
	"log/slog"
	"os"
)

func main() {
	cfg := config.MustLoad()
	db, err := postgresql.New(cfg.DbConfig)
	if err != nil {
		slog.Error("couldn't open db:", err)
	}
	if err = db.Ping(); err != nil {
		slog.Error("couldn't ping db", err)
	} else {
		slog.Info("pinged db successfully")
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

	uuid := gofakeit.UUID()
	order := storage.RandomOrder(uuid)
	fmt.Println(order)
	err = db.SaveOrder(order)
	if err != nil {
		slog.Error("couldn't save random:", err)
		os.Exit(1)
	}

	res, err := db.GetOrder(uuid)
	if err != nil {
		slog.Error("random save didnt work:", err)
		os.Exit(1)
	}

	fmt.Println(res)

}
