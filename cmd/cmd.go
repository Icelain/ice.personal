package cmd

import (
	
	"net/http"
	"iceblog/controllers"
	"flag"
	"fmt"
	"log"

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
	controllers.HandleBlog(mux, "./blogs")

	log.Printf("iceblog listening on port %d", *port)
	http.ListenAndServe(fmt.Sprintf(":%d",*port), mux)

}
