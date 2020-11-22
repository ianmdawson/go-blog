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
	"time"

	"github.com/gofrs/uuid"
	"github.com/ianmdawson/go-blog/models"
)

const templateDir string = "tmpl"

var templates = template.Must(template.ParseGlob(templateDir + "/*.gohtml"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([-a-zA-Z0-9]+)$")

// TODO:
//
// - Spruce up the page templates by making them valid HTML and adding some CSS rules. use yield to crate an application layout instead of header/footer pattern (https://www.calhoun.io/intro-to-templates-p4-v-in-mvc/)
// - Paginated Page index
// - Page submission form template
// - Separate models into model folder
// - separate log/routing logic into "routes"?
// - separate DB access into its own folder
// - Add a handler to make the web root redirect for /
// - Implement inter-page linking by converting instances of [PageName] to
//     <a href="/view/PageName">PageName</a>. (hint: you could use regexp.ReplaceAllFunc to do this)
// - Users, permissions

func databaseURL() string {
	var databaseURL = os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		return databaseURL
	}
	return "postgres://goblog:password@localhost:5432/blog_dev"
}

// Page represents page data
type Page struct {
	ID        uuid.UUID
	Title     string
	Body      []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *Page) update() error {
	sql := ` -- name: PageUpdate :one
		UPDATE pages
		SET title = $2, body = $3, updated_at = now()
		WHERE id=$1
		RETURNING id, title, body, created_at, updated_at
		;`
	err := models.DB.QueryRow(context.Background(), sql, p.ID, p.Title, p.Body).Scan(&p.ID, &p.Title, &p.Body, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (p *Page) create() error {
	sql := ` -- name: PageCreate :one
		INSERT INTO pages
		(id, title, body)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
		RETURNING id, title, body, created_at, updated_at
		;`
	err := models.DB.QueryRow(context.Background(), sql, p.ID, p.Title, p.Body).Scan(&p.ID, &p.Title, &p.Body, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (p *Page) find(id uuid.UUID) error {
	sql := `SELECT id, title, body, created_at, updated_at FROM pages WHERE id=$1;`
	err := models.DB.QueryRow(context.Background(), sql, id).Scan(&p.ID, &p.Title, &p.Body, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

// GetAllPages retrieves all Page models in the database
// limit the number of results to return
// offset the number of results to skip, useful for pagination
func GetAllPages(offset int, limit int) ([]*Page, error) {
	sql := `SELECT id, title, body, created_at, updated_at FROM pages ORDER BY created_at DESC OFFSET $1 LIMIT $2;`
	rows, err := models.DB.Query(context.Background(), sql, offset, limit)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var pages []*Page

	for rows.Next() {
		p := &Page{}
		err = rows.Scan(&p.ID, &p.Title, &p.Body, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}

	return pages, nil
}

func loadPage(id string) (*Page, error) {
	uuid := uuid.FromStringOrNil(id)
	page := &Page{}
	err := page.find(uuid)
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
		// http.NotFound(w, r)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, id string) {
	p, err := loadPage(id)
	if err != nil {
		p = &Page{Title: id}
	}
	renderTemplate(w, "edit", p)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.URL.Path) // log request
	// TODO: get skip and limit from query params
	skip := 0
	limit := 50
	pages, err := GetAllPages(skip, limit)
	if err != nil {
		fmt.Println("Something went wrong loading pages:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = templates.ExecuteTemplate(w, "index.html.gohtml", pages)
	if err != nil {
		fmt.Println("500", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.URL.Path) // log request
	renderTemplate(w, "new", &Page{})
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
	page := &Page{ID: uuid, Body: []byte(body), Title: title}
	err = page.create()
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
	p := &Page{ID: uuid, Title: title, Body: []byte(body)}
	err = p.update()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+id, http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html.gohtml", p)
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
	defer models.DB.Close(context.Background())

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/new/", newHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/index", indexHandler)

	port := ":8080"
	fmt.Println("Setting up to listen on port ", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
