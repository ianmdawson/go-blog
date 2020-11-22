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

	"github.com/gofrs/uuid"
	"github.com/ianmdawson/go-blog/models"
)

const templateDir string = "tmpl"

var templates = template.Must(template.ParseGlob(templateDir + "/*.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([-a-zA-Z0-9]+)$")

// TODO:
// - New Page button
// - Spruce up the page templates by making them valid HTML and adding some CSS rules. use yield to crate an application layout instead of header/footer pattern (https://www.calhoun.io/intro-to-templates-p4-v-in-mvc/)
// - Paginated Page index
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
	PageIndexPath: "/index",
	PageNewPath:   "/new",
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

func viewHandler(w http.ResponseWriter, r *http.Request, id string) {
	p, err := loadPage(id)
	if err != nil {
		http.Redirect(w, r, "/new/", http.StatusFound)
		// TODO: make this 404 NotFound instead
		// http.NotFound(w, r)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, id string) {
	p, err := loadPage(id)
	if err != nil {
		p = &models.Page{Title: id}
	}
	renderTemplate(w, "edit", p)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.URL.Path) // log request
	// TODO: get skip and limit from query params
	offset := 0
	limit := 50
	pages, err := models.GetAllPages(offset, limit)
	if err != nil {
		fmt.Println("Something went wrong loading pages:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	count, err := models.CountAllPages()
	if err != nil {
		fmt.Println("Something went wrong loading pages count:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	indexData := struct {
		Pages             []*models.Page
		Page              *models.Page
		Count             *int
		ResultsPageNumber int
		Limit             int
		Links             pagePaths
	}{
		pages,
		pages[0],
		count,
		1 + (offset / limit),
		limit,
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

func saveHandler(w http.ResponseWriter, r *http.Request, id string) {
	body := r.FormValue("body")
	title := r.FormValue("title")
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

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s\n", r.Method, r.URL.Path) // log request
		id, err := getIDFromRequest(w, r)
		if err != nil {
			fmt.Println(err)
			http.NotFound(w, r)
			return
		}
		fn(w, r, id)
	}
}

func main() {
	err := models.InitDB(databaseURL())
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/new/", newHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/index", indexHandler)

	port := ":8080"
	fmt.Println("Setting up to listen on port ", port)
	log.Fatal(http.ListenAndServe(port, nil))
	defer models.DB.Close(context.Background())
}
