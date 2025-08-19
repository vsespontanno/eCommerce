package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"github.com/vsespontanno/eCommerce/products-service/internal/db"
	"github.com/vsespontanno/eCommerce/products-service/internal/repository/pg"
	serv "github.com/vsespontanno/eCommerce/products-service/internal/server"
	proto "github.com/vsespontanno/eCommerce/proto/products"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	grpcEndpoint := ":8082"

	conn, err := db.ConnectToPostgres(os.Getenv("PG_USER"), os.Getenv("PG_PASSWORD"), os.Getenv("PG_DB"), os.Getenv("PG_HOST"), os.Getenv("PG_PORT"))
	if err != nil {
		log.Fatal(err)
	}

	productStore := pg.NewProductStore(conn)

	log.Fatal(makeGRPCTransport(grpcEndpoint, productStore))

}

func makeGRPCTransport(endpoint string, productStore *pg.ProductStore) error {
	ln, err := net.Listen("tcp", endpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	server := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
	proto.RegisterProductsServer(server, serv.NewGrpcServer(productStore))
	fmt.Println("GRPC transport running on port", endpoint)
	// seedSomeProductsIntoDB(productStore)
	return server.Serve(ln)

}

// func seedSomeProductsIntoDB(pg *pg.ProductStore) {
// 	products := []*models.Product{
// 		{ID: int64(uuid.New().ID()), Name: "Chocolate", Price: 10.0, Description: "Very Tasty", Category: "Food", Brand: "Dove", Rating: 4, NumReviews: 10, CountInStock: 100},
// 		{ID: int64(uuid.New().ID()), Name: "Red Bull", Price: 20.0, Description: "Energetic drink for trainings", Category: "Cold Drinks", Brand: "Red Bull", Rating: 4, NumReviews: 5, CountInStock: 50},
// 		{ID: int64(uuid.New().ID()), Name: "Tide", Price: 30.0, Description: "For laundry", Category: "Household", Brand: "Tide", Rating: 4, NumReviews: 2, CountInStock: 0},
// 	}

// 	for _, p := range products {
// 		if err := pg.SaveProduct(context.Background(), p); err != nil {
// 			log.Printf("Error seeding product %v: %v", p.Name, err)
// 		}
// 	}
// 	fmt.Println("Products seeded successfully")
// }
