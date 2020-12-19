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
// - Users and permissions
// - Documentation
// - routing/http handler tests
// - Spruce up the page templates by making them valid HTML and adding some CSS rules. use yield to crate an application layout instead of header/footer pattern (https://www.calhoun.io/intro-to-templates-p4-v-in-mvc/)
// - Page edit/new shared submission form template
// - logging middleware
// - separate log/routing logic into logger/"routes"?
// - Implement inter-page linking by converting instances of [PageName] to
//     <a href="/view/PageName">PageName</a>. (hint: you could use regexp.ReplaceAllFunc to do this?)

type pagePaths struct {
	PageEditPath   string
	PageIndexPath  string
	PageNewPath    string
	PageViewPath   string
	PageCreatePath string
	PageSavePath   string
}

var links = pagePaths{
	PageEditPath:   "/pages/edit/",
	PageIndexPath:  "/",
	PageNewPath:    "/pages/new/",
	PageViewPath:   "/pages/",
	PageCreatePath: "/pages/create/",
	PageSavePath:   "/pages/save/",
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

	var firstPage *models.Page
	if firstPages == nil {
		firstPage = nil
	} else {
		firstPage = firstPages[0]
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
		firstPage,
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

func newPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.URL.Path) // log request
	renderTemplate(w, "new", &models.Page{})
}

func createPageHandler(w http.ResponseWriter, r *http.Request) {
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
	http.Redirect(w, r, fmt.Sprintf("%s%s", links.PageViewPath, page.ID), http.StatusFound)
}

func savePage(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.URL.Path) // log request
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
	http.Redirect(w, r, links.PageViewPath+id, http.StatusFound)
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

func viewPage(w http.ResponseWriter, r *http.Request) {
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

func editPage(w http.ResponseWriter, r *http.Request) {
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
	r.HandleFunc(links.PageIndexPath, indexHandler)
	r.HandleFunc(links.PageViewPath+"{id:[a-z0-9-]+}", viewPage)
	r.HandleFunc(links.PageEditPath+"{id:[a-z0-9-]+}", editPage)
	r.HandleFunc(links.PageSavePath+"{id:[a-z0-9-]+}", savePage)
	r.HandleFunc(links.PageNewPath, newPageHandler)
	r.HandleFunc(links.PageCreatePath, createPageHandler)

	http.Handle("/", r)

	port := ":8080"
	fmt.Println("Setting up to listen on port ", port)
	log.Fatal(http.ListenAndServe(port, nil))
	defer models.DB.Close(context.Background())
}
