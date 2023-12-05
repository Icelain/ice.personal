package controllers

import (

	"net/http"
	"html/template"
	"log"
	"strings"
	"os"
	"fmt"
	"io"
	"iceblog/markdown"	
	"github.com/go-chi/chi/v5"
)

type Page struct {
	
	Title string
	Date string
	Content template.HTML

}

type Pages struct {

	P []Page

}

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
	
	var pages Pages

	files, err := os.ReadDir("./blogs")
	if err != nil {

		return err

	}

	for _, file := range files {
		
		osfile, err := os.Open("./blogs/" + file.Name())
		if err != nil {
			
			return err

		}
	
		markdownBytes, err := io.ReadAll(osfile)
		if err != nil {log.Fatal(err)}

		markdownString := string(markdownBytes)
		
		mdarr := strings.Split(markdownString, "---")

		renderedMd := markdown.Render(mdarr[2])
		yamlMap := markdown.ParseYaml(mdarr[1])

		title := strings.TrimSuffix(file.Name(), ".md")
		page := Page {Title: title, Content: template.HTML(renderedMd), Date: yamlMap["date"]}
		
		pages.P = append(pages.P, page)

		tmpl, err := template.ParseFiles("./templates/blog.gohtml")
		if err != nil {

			return err

		}

		router.Get(fmt.Sprintf("/blog/%s", title), func(w http.ResponseWriter, r *http.Request) {

			if err := tmpl.Execute(w, page); err != nil {
			
				log.Printf(fmt.Sprintf("failed to execute template %s", file.Name()))

			}	
		})

	}
	
	router.Get("/blog", func(w http.ResponseWriter, r *http.Request) {

		tmpl := template.Must(template.ParseFiles("./templates/blogindex.gohtml"))
		tmpl.Execute(w, pages)

	})
	
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
