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

// TODO:
// Users, permissions, settings
// Better documentation
// template improvments
// 	- Spruce up the page templates by making them valid HTML and adding some CSS rules. use yield to crate an application layout instead of header/footer pattern (https://www.calhoun.io/intro-to-templates-p4-v-in-mvc/)
// 	- Implement inter-page linking by converting instances of [PageName] to
//     <a href="/view/PageName">PageName</a>. (hint: you could use regexp.ReplaceAllFunc to do this?)
// testing
//	- dockerize tests and test setup
//	- make database reset more efficient
//	- avoid requiring the database setup at all via mocking: https://github.com/jackc/pgx/issues/616#issuecomment-535749087
// 	- More routing/http handler tests

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// fmt.Printf("%s %s\n", , r.URL.Path) // log request
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
	r.Use(handlers.SessionMiddleware)

	pagePaths := handlers.PagePaths
	r.HandleFunc(pagePaths.PageIndexPath, handlers.IndexHandler)
	r.HandleFunc(pagePaths.PageViewPath+"{id:[a-z0-9-]+}", handlers.ViewPage)
	r.HandleFunc(pagePaths.PageEditPath+"{id:[a-z0-9-]+}", handlers.EditPage)
	r.HandleFunc(pagePaths.PageSavePath+"{id:[a-z0-9-]+}", handlers.SavePage)
	r.HandleFunc(pagePaths.PageNewPath, handlers.NewPage)
	r.HandleFunc(pagePaths.PageCreatePath, handlers.CreatePageHandler)

	r.HandleFunc("/signup/", handlers.SignUpHandler)
	r.HandleFunc(handlers.UserPaths.UserCreatePath, handlers.CreateUserHandler)
	r.HandleFunc(handlers.UserPaths.UserLogInPath, handlers.LogInHandler)
	r.HandleFunc(handlers.UserPaths.UserAuthenticatePath, handlers.AuthenticateUserHandler)

	http.Handle("/", r)

	port := ":8080"
	fmt.Println("Setting up to listen on port ", port)
	log.Fatal(http.ListenAndServe(port, nil))
	defer models.DB.Close(context.Background())
}
