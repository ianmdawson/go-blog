package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/ianmdawson/go-blog/models"
)

var (
	_, b, _, _ = runtime.Caller(0)
	// basepath is the package root directory
	basepath = filepath.Dir(b)
)

var templateDir string = basepath + "/../tmpl"

// Templates are the html templates for the blog
var Templates = template.Must(template.ParseGlob(templateDir + "/*.html"))

// PagePatterns dictates which paths are available for the interacting with the Page model
type PagePatterns struct {
	PageEditPath   string
	PageIndexPath  string
	PageNewPath    string
	PageViewPath   string
	PageCreatePath string
	PageSavePath   string
}

// PagePaths Returns all page URI pattern prefixes
// PagePaths page paths for routing and linking
var PagePaths = PagePatterns{
	PageEditPath:   "/pages/edit/",
	PageIndexPath:  "/",
	PageNewPath:    "/pages/new/",
	PageViewPath:   "/pages/",
	PageCreatePath: "/pages/create/",
	PageSavePath:   "/pages/save/",
}

// NewPage renders the new page template for users to create a new Page
func NewPage(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.URL.Path) // log request
	renderTemplate(w, "new", &models.Page{})
}

// ViewPage renders the Page if the given ID exists, otherwise it redirects to NewPage
func ViewPage(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.URL.Path) // log request

	vars := mux.Vars(r)
	id := vars["id"]

	p, err := LoadPage(id)
	if err != nil {
		// TODO: make this 404 NotFound instead
		// http.NotFound(w, r)
		http.Redirect(w, r, PagePaths.PageNewPath, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

// EditPage renders the edit template for a given :id in query the params
func EditPage(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.URL.Path) // log request

	vars := mux.Vars(r)
	id := vars["id"]

	p, err := LoadPage(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "edit", p)
}

// SavePage saves/updates an existing page
func SavePage(w http.ResponseWriter, r *http.Request) {
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
	http.Redirect(w, r, PagePaths.PageViewPath+id, http.StatusFound)
}

// CreatePageHandler creates a new Page
func CreatePageHandler(w http.ResponseWriter, r *http.Request) {
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
	http.Redirect(w, r, fmt.Sprintf("%s%s", PagePaths.PageViewPath, page.ID), http.StatusFound)
}

// LoadPage given an id
func LoadPage(id string) (*models.Page, error) {
	uuid := uuid.FromStringOrNil(id)
	page := &models.Page{}
	err := page.Find(uuid)
	if err != nil {
		fmt.Println("Error finding page", err)
		return nil, err
	}
	return page, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *models.Page) {
	var templateData = struct {
		Page  *models.Page
		Links PagePatterns
	}{
		p,
		PagePaths,
	}
	err := Templates.ExecuteTemplate(w, tmpl+".html", templateData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// IndexHandler renders the index Page index page, the most recent Page and a list of other most recent pages
func IndexHandler(w http.ResponseWriter, r *http.Request) {
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
	defaultLimit := 5
	maxLimit := 10
	limit := defaultLimit
	if len(limitParam) > 0 {
		limitInt, err := strconv.Atoi(limitParam[0])
		if err != nil {
			fmt.Println("An error occurred parsing the resultsLimitParam", err)
		}
		if limitInt > 0 && limitInt < maxLimit {
			limit = limitInt
		}
	}

	offset := (resultsPage - 1) * limit
	pageCollection, err := models.GetPageCollection(offset, limit)
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

	indexData := struct {
		Page           *models.Page
		PageCollection *models.PageCollection
		Links          PagePatterns
	}{
		firstPage,
		pageCollection,
		PagePaths,
	}
	err = Templates.ExecuteTemplate(w, "index.html", indexData)
	if err != nil {
		fmt.Println("500", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
