package controllers

import (
	"fmt"
	"html/template"
	"iceblog/markdown"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type Page struct {
	Title       string
	DashedTitle string
	Date        string
	Content     template.HTML
}

type Pages struct {
	P []Page
}

func titalize(str string) string {

	title := ""
	for _, word := range strings.Split(str, " ") {

		title += strings.ToUpper(string(word[0])) + word[1:] + " "

	}

	return strings.TrimSuffix(title, " ")

}

func HandleFavicon(r chi.Router) {

	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {

		http.Redirect(w,r,"/static/favicon.ico",http.StatusSeeOther)

	})

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

func HandleBlog(router chi.Router, dir string) error {

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
		if err != nil {
			log.Fatal(err)
		}

		markdownString := string(markdownBytes)

		mdarr := strings.Split(markdownString, "---")

		renderedMd := markdown.Render(mdarr[2])
		yamlMap := markdown.ParseYaml(mdarr[1])

		dashTitle := strings.TrimSuffix(file.Name(), ".md")
		title := titalize(strings.ReplaceAll(dashTitle, "-", " "))
		page := Page{Title: title, DashedTitle: dashTitle, Content: template.HTML(renderedMd), Date: yamlMap["date"]}

		pages.P = append(pages.P, page)

		tmpl, err := template.ParseFiles("./templates/blog.gohtml")
		if err != nil {

			return err

		}

		router.Get(fmt.Sprintf("/blog/%s", dashTitle), func(w http.ResponseWriter, r *http.Request) {

			if err := tmpl.Execute(w, page); err != nil {

				log.Printf(fmt.Sprintf("failed to execute template %s", file.Name()))

			}
		})

	}

	router.Get("/blog", func(w http.ResponseWriter, r *http.Request) {

		tmpl := template.Must(template.ParseFiles("./templates/blogindex.gohtml"))

		// sorts in descending order (not ascending) !!
		// cmp function basically does the opposite of what its supposed to do
		slices.SortFunc[[]Page, Page](pages.P, func(a, b Page) int {

			at, err := time.Parse("2006/01/02", a.Date)
			if err != nil {

				log.Fatal("time format in markdown is wrong", err)

			}

			bt, err := time.Parse("2006/01/02", b.Date)
			if err != nil {
				log.Fatal("time format in markdown is wrong", err)

			}

			atime := at.Unix()
			btime := bt.Unix()

			if atime > btime {

				return -1

			} else if atime < btime {

				return 1

			}

			return 0
		})

		if err := tmpl.Execute(w, pages); err != nil {

			log.Fatal(err)

		}

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

func MusicServer(r chi.Router, path string, musicfilepath string) {

	r.Get("/music", func(w http.ResponseWriter, r *http.Request) {

		tmpl := template.Must(template.ParseFiles("./templates/musicindex.gohtml"))
		if err := tmpl.Execute(w, nil); err != nil {

			log.Fatal(err)

		}

	})

	StreamRadio(r, path, musicfilepath)

}
