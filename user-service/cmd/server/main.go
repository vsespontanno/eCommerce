package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/joho/godotenv"
	proto "github.com/vsespontanno/eCommerce/proto/user"
	"github.com/vsespontanno/eCommerce/user-service/internal/auth"
	"github.com/vsespontanno/eCommerce/user-service/internal/db"
	"github.com/vsespontanno/eCommerce/user-service/internal/repository/pg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	err := godotenv.Load("user-service/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	grpcEndpoint := ":8081"

	conn, err := db.ConnectToPostgres(os.Getenv("PG_USER"), os.Getenv("PG_PASSWORD"), os.Getenv("PG_DB"), os.Getenv("PG_HOST"), os.Getenv("PG_PORT"))
	if err != nil {
		log.Fatal(err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	jwtService, err := auth.NewJwtService(jwtSecret)
	if err != nil {
		log.Fatal(err)
	}
	userStore := pg.NewUserStore(conn)

	authService := auth.NewAuthService(userStore, jwtService, time.Duration(1)*time.Hour)
	log.Fatal(makeGRPCTransport(grpcEndpoint, authService, jwtService))

}

func makeGRPCTransport(endpoint string, authService *auth.AuthService, jwtService *auth.JwtService) error {
	ln, err := net.Listen("tcp", endpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	server := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
	proto.RegisterAuthServer(server, authService)
	fmt.Println("GRPC transport running on port", endpoint)
	return server.Serve(ln)

}
