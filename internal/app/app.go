package app

import (
	"crm-admin/config"
	"crm-admin/internal/controller"
	grpc1 "crm-admin/internal/controller/grpc"
	"crm-admin/internal/generated/products"
	"crm-admin/pkg/logger"
	"crm-admin/pkg/postgres"
	"google.golang.org/grpc"
	"log"
	"net"
)

func Run(cfg config.Config) {

	logger1 := logger.NewLogger()

	db, err := postgres.Connection(cfg)
	if err != nil {
		log.Fatal(err)
	}

	controller1 := controller.NewController(db, logger1)
	pr := grpc1.NewProductGrpc(controller1, logger1)

	listen, err := net.Listen("tcp", cfg.RUN_PORT)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	products.RegisterProductsServer(server, pr)

	log.Printf("server listening at %s", cfg.RUN_PORT)
	log.Fatal(server.Serve(listen))
}
