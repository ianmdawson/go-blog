package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/ianmdawson/go-blog/handlers"
	"github.com/ianmdawson/go-blog/models"
)

const templateDir string = "tmpl"

// Temporary until all handlers are moved to the handlers package
var templates = handlers.Templates

// TODO:
// - Users and permissions
// - Documentation

// - routing/http handler tests

// - Spruce up the page templates by making them valid HTML and adding some CSS rules. use yield to crate an application layout instead of header/footer pattern (https://www.calhoun.io/intro-to-templates-p4-v-in-mvc/)
// - Page edit/new shared submission form template
// - logging middleware
// - separate log/routing logic into logger/"routes"?
// - Implement inter-page linking by converting instances of [PageName] to
//     <a href="/view/PageName">PageName</a>. (hint: you could use regexp.ReplaceAllFunc to do this?)

var links = handlers.PagePaths

func databaseURL() string {
	var databaseURL = os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		return databaseURL
	}
	return "postgres://goblog:password@localhost:5432/blog_dev"
}

func main() {
	err := models.InitDB(databaseURL())
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	pagePaths := handlers.PagePaths
	r.HandleFunc(pagePaths.PageIndexPath, handlers.IndexHandler)
	r.HandleFunc(pagePaths.PageViewPath+"{id:[a-z0-9-]+}", handlers.ViewPage)
	r.HandleFunc(pagePaths.PageEditPath+"{id:[a-z0-9-]+}", handlers.EditPage)
	r.HandleFunc(pagePaths.PageSavePath+"{id:[a-z0-9-]+}", handlers.SavePage)
	r.HandleFunc(pagePaths.PageNewPath, handlers.NewPage)
	r.HandleFunc(pagePaths.PageCreatePath, handlers.CreatePageHandler)

	http.Handle("/", r)

	port := ":8080"
	fmt.Println("Setting up to listen on port ", port)
	log.Fatal(http.ListenAndServe(port, nil))
	defer models.DB.Close(context.Background())
}
