package main

import (
	proxy "hw/internal/pkg/proxy"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found" + err.Error())
	}
}

func main() {

	if err := run(); err != nil {
		os.Exit(1)
	}

}

func run() error {
	log.Println(os.Getenv("DATABASE_URL"))

	// db, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	// if err != nil {
	// 	log.Println("fail open postgres", err)
	// 	err = fmt.Errorf("error happened in sql.Open: %w", err)

	// 	return err
	// }
	// defer db.Close()

	// if err = db.Ping(context.Background()); err != nil {
	// 	log.Println("fail ping postgres", err)
	// 	err = fmt.Errorf("error happened in db.Ping: %w", err)

	// 	return err
	// }
	// repo := repo.NewRepo(db)

	pr := proxy.NewProxy()

	srv := http.Server{
		Addr: "127.0.0.1:8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				pr.HandleConnect(w, r)
				log.Println("connect")
			} else {
				pr.Handle(w, r)
				log.Println("http")
			}
		}),
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

	return nil
}
