package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ianmdawson/go-blog/handlers"
	"github.com/ianmdawson/go-blog/models"
	"github.com/joho/godotenv"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	err = models.InitDB("")
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.Use(loggingMiddleware)

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
