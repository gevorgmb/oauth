package main

import (
	"log"
	"net"
	"oauth/internal/dto"
	"oauth/internal/server"
	"oauth/internal/storage"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	_ = godotenv.Load()
	addr := getenv("GRPC_ADDR", "0.0.0.0:50051")
	dnsStr := getenv("DB_DNS", "local_root:123456@tcp(localhost:3306)/oauthdb?parseTime=true")

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("faild to listen: %v", err)
	}
	conMaxLft, err := strconv.Atoi(getenv("DB_CONNECTION_MAX_LIFETIME", "5"))
	if err != nil {
		log.Fatal("faild to get connection max lifetime from env file")
	}

	maxOpenConns, err := strconv.Atoi(getenv("DB_MAX_OPEN_CONNECTIONS", "10"))
	if err != nil {
		log.Fatal("feild to get max open connections from env file")
	}

	maxIdleConns, err := strconv.Atoi(getenv("DB_MAX_IDLE_CONNECTIONS", "5"))
	if err != nil {
		log.Fatal("feild to get max idle connections from env file")
	}

	mySQLData := dto.NewMySQLConnectionDto(
		dnsStr,
		conMaxLft,
		maxOpenConns,
		maxIdleConns,
	)

	store, err := storage.NewMySQL(mySQLData)
	if err != nil {
		log.Fatalf("failed to reate server: %v", err)
	}

	srv, err := server.NewServer(store)
	if err != nil {
		log.Fatalf("failed to reate server: %v", err)
	}

	unprotected := map[string]bool{
		"/oauth.OAuth/Register": true,
		"/oauth.OAuth/Token":    true,
		"/oauth.OAuth/Verify":   true,
		"/oauth.OAuth/Refresh":  true,
	}
	adminOnly := map[string]bool{
		"/oauth.OAuth/ListUser":   true,
		"/oauth.OAuth/DeleteUser": true,
	}

	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(server.AuthInterceptor(
			os.Getenv("JWT_SECRET"),
			unprotected,
			adminOnly,
		)),
	)
	srv.RegisterGRPC(grpcSrv)

	log.Printf("gRPC server listeningon %s", addr)

	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatalf("gRPC serve error: %v", err)
	}
}

func getenv(k string, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
