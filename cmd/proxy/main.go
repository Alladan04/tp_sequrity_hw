package main

import (
	proxyHttp "hw/internal/pkg/proxy/http"
	"log"
	"net/http"
)

func main() {

	// db, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	// if err != nil {
	// 	log.Println("error connecting to postgres: " + err.Error())
	// 	return
	// }
	// defer db.Close()

	http.HandleFunc("/", proxyHttp.Handler)
	log.Println("Proxy server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
