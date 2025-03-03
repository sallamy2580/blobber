package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/0chain/blobber/code/go/0chain.net/core/common"

	"github.com/0chain/blobber/code/go/0chain.net/blobbercore/handler"
	"github.com/0chain/blobber/code/go/0chain.net/core/logging"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/reflection"
)

func startGRPCServer() {
	common.ConfigRateLimits()
	r := mux.NewRouter()
	initHandlers(r)

	grpcServer := handler.NewGRPCServerWithMiddlewares(r)
	reflection.Register(grpcServer)

	if grpcPort <= 0 {
		logging.Logger.Error("grpc port missing")
		panic(errors.New("grpc port missing"))
	}

	logging.Logger.Info("started grpc server on to grpc requests on port - " + strconv.Itoa(grpcPort))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		panic(err)
	}

	fmt.Print("> starting grpc server	[OK]\n")

	log.Fatal(grpcServer.Serve(lis))

}
