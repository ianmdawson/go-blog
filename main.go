package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/ianmdawson/go-blog/models"
)

const templateDir string = "tmpl"

var templates = template.Must(template.ParseGlob(templateDir + "/*.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([-a-zA-Z0-9]+)$")

// TODO:
// - routing/http handler tests
// - New Page button
// - Spruce up the page templates by making them valid HTML and adding some CSS rules. use yield to crate an application layout instead of header/footer pattern (https://www.calhoun.io/intro-to-templates-p4-v-in-mvc/)
// - Page submission form template
// - separate log/routing logic into logger/"routes"?
// - Add a handler to make the web root redirect for /
// - Implement inter-page linking by converting instances of [PageName] to
//     <a href="/view/PageName">PageName</a>. (hint: you could use regexp.ReplaceAllFunc to do this)
// - Users, permissions

type pagePaths struct {
	PageEditPath  string
	PageIndexPath string
	PageNewPath   string
	PageViewPath  string
}

var links = pagePaths{
	PageEditPath:  "/edit/",
	PageIndexPath: "/",
	PageNewPath:   "/posts/new/",
	PageViewPath:  "/view",
}

func databaseURL() string {
	var databaseURL = os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		return databaseURL
	}
	return "postgres://goblog:password@localhost:5432/blog_dev"
}

func loadPage(id string) (*models.Page, error) {
	uuid := uuid.FromStringOrNil(id)
	page := &models.Page{}
	err := page.Find(uuid)
	if err != nil {
		fmt.Println("Error finding page", err)
		return nil, err
	}
	return page, nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.URL.Path) // log request
	q := r.URL.Query()
	resultsPageParam := q["page"]
	resultsPage := 1
	if len(resultsPageParam) != 0 {
		resultsPageInt, err := strconv.Atoi(resultsPageParam[0])
		if err != nil {
			fmt.Println("An error occurred parsing the resultsPageParam", err)
			resultsPage = 1
		}
		if resultsPageInt > 0 {
			resultsPage = resultsPageInt
		}
	}

	limitParam := q["limit"]
	limit := 50
	if len(limitParam) > 0 {
		limitInt, err := strconv.Atoi(limitParam[0])
		if err != nil {
			fmt.Println("An error occurred parsing the resultsLimitParam", err)
		}
		if limitInt > 0 && limitInt < 50 {
			limit = limitInt
		}
	}

	offset := (resultsPage - 1) * limit
	pages, err := models.GetAllPages(offset, limit)
	if err != nil {
		fmt.Println("Something went wrong loading pages:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	firstPages, err := models.GetAllPages(0, 1)
	if err != nil {
		fmt.Println("Something went wrong loading the first page:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	count, err := models.CountAllPages()
	if err != nil {
		fmt.Println("Something went wrong loading pages count:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	prevPageNumber := resultsPage - 1
	if prevPageNumber < 0 {
		prevPageNumber = 0
	}
	atLastPage := ((resultsPage-1)*limit)+len(pages) >= count

	// TODO: move pagination params into their own struct to clean up indexHandler
	indexData := struct {
		Pages             []*models.Page
		Page              *models.Page
		Count             int
		ResultsPageNumber int
		Limit             int
		NextPage          int
		PreviousPage      int
		AtLastPage        bool
		Links             pagePaths
	}{
		pages,
		firstPages[0],
		count,
		resultsPage,
		limit,
		resultsPage + 1,
		resultsPage - 1,
		atLastPage,
		links,
	}
	err = templates.ExecuteTemplate(w, "index.html", indexData)
	if err != nil {
		fmt.Println("500", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.URL.Path) // log request
	renderTemplate(w, "new", &models.Page{})
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.URL.Path) // log request
	body := r.FormValue("body")
	title := r.FormValue("title")
	uuid, err := uuid.NewV4()
	if err != nil {
		fmt.Println("Failed to generate UUID", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	page := &models.Page{ID: uuid, Body: []byte(body), Title: title}
	err = page.Create()
	if err != nil {
		fmt.Println("Error saving page", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/view/%s", page.ID), http.StatusFound)
}

func savePost(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("body")
	title := r.FormValue("title")

	vars := mux.Vars(r)
	id := vars["id"]

	uuid, err := uuid.FromString(id)
	if err != nil {
		fmt.Println("Cannot parse ID for: ", id)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	p := &models.Page{ID: uuid, Title: title, Body: []byte(body)}
	err = p.Update()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+id, http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *models.Page) {
	var templateData = struct {
		Page  *models.Page
		Links pagePaths
	}{
		p,
		links,
	}
	err := templates.ExecuteTemplate(w, tmpl+".html", templateData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getIDFromRequest(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		fmt.Println(r.URL.Path)
		return "", errors.New("invalid Page")
	}
	return m[2], nil // Title is the second sub-expression
}

func viewPost(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.URL.Path) // log request

	vars := mux.Vars(r)
	id := vars["id"]

	p, err := loadPage(id)
	if err != nil {
		// TODO: make this 404 NotFound instead
		// http.NotFound(w, r)
		http.Redirect(w, r, links.PageNewPath, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editPost(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.URL.Path) // log request

	vars := mux.Vars(r)
	id := vars["id"]

	p, err := loadPage(id)
	if err != nil {
		p = &models.Page{Title: id}
	}
	renderTemplate(w, "edit", p)
}

func main() {
	err := models.InitDB(databaseURL())
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/view/{id:[a-z0-9-]+}", viewPost)
	r.HandleFunc("/edit/{id:[a-z0-9-]+}", editPost)
	r.HandleFunc("/save/{id:[a-z0-9-]+}", savePost)
	r.HandleFunc("/posts/new/", newHandler)
	r.HandleFunc("/create", createHandler)

	http.Handle("/", r)

	port := ":8080"
	fmt.Println("Setting up to listen on port ", port)
	log.Fatal(http.ListenAndServe(port, nil))
	defer models.DB.Close(context.Background())
}
