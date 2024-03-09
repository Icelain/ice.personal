package cmd

import (
	"flag"
	"fmt"
	"iceblog/controllers"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Execute() {

	port := flag.Int("port", 8080, "./server -port 5000")
	flag.Parse()

	mux := chi.NewRouter()

	// handle controllers

	controllers.HandleHTML(mux, []string{"/", "/index"}, "./templates/index.html")

	// handle static files

	controllers.FileServer(mux, "/static/", http.Dir("static"))
	if err := controllers.HandleBlog(mux, "./blogs"); err != nil {

		log.Fatal(err)

	}

	controllers.MusicServer(mux, "/music/stream", "./audio/result.aac")

	log.Printf("iceblog listening on port %d", *port)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), mux)

}
