package controllers

import (

	"net/http"
	"html/template"
	"log"
	"strings"
	"os"
	"fmt"

//	"github.com/russross/blackfriday/v2"	
	"github.com/go-chi/chi/v5"
)

func HandleHTML(r chi.Router, paths []string, htmlpath string) {
	
	for _, path := range paths {

		r.Get(path, func(w http.ResponseWriter, r *http.Request) {

			tmpl, err := template.ParseFiles(htmlpath)
			if err != nil {

				log.Fatal("Could not find the template lol")
				return

			}

			if err := tmpl.Execute(w, nil); err != nil {
				
				log.Printf("failed to execute template index.html")	

			}	

		})		

	}

}

func HandleBlog(router chi.Router, dir string) error{

	files, err := os.ReadDir("./blogs")
	if err != nil {

		return err

	}

	for _, file := range files {

		tmpl := template.Must(template.ParseFiles(dir + "/" + file.Name()))
		postname := strings.TrimSuffix(tmpl.Name(), ".html")

		router.Get(fmt.Sprintf("/blog/%s", postname), func(w http.ResponseWriter, r *http.Request) {

			if err := tmpl.Execute(w, nil); err != nil {
			
				log.Printf(fmt.Sprintf("failed to execute template %s", file.Name()))

			}	
		})

	}
	
	return nil
}


func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
