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
	"github.com/jackc/pgx/v4"
)

var templates = template.Must(template.ParseFiles("tmpl/edit.html.gohtml", "tmpl/view.html.gohtml", "tmpl/new.html.gohtml"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([-a-zA-Z0-9]+)$")
var db *pgx.Conn

func getDB() *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), databaseURL())
	if err != nil {
		panic(err)
	}
	return conn
}

func databaseURL() string {
	var databaseURL = os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		return databaseURL
	}
	return "postgres://goblog:password@localhost:5432/blog_dev"
}

// TODO:
// - Add a handler to make the web root redirect for /
// - Spruce up the page templates by making them valid HTML and adding some CSS rules.
// - Implement inter-page linking by converting instances of [PageName] to
//     <a href="/view/PageName">PageName</a>. (hint: you could use regexp.ReplaceAllFunc to do this)

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
		UPDATE posts
		SET title = $2, body = $3, updated_at = now()
		WHERE id=$1
		RETURNING id, title, body, created_at, updated_at
		;`
	err := db.QueryRow(context.Background(), sql, p.ID, p.Title, p.Body).Scan(&p.ID, &p.Title, &p.Body, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (p *Page) create() error {
	sql := ` -- name: PageCreate :one
		INSERT INTO posts
		(id, title, body)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
		RETURNING id, title, body, created_at, updated_at
		;`
	err := db.QueryRow(context.Background(), sql, p.ID, p.Title, p.Body).Scan(&p.ID, &p.Title, &p.Body, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (p *Page) find(id uuid.UUID) error {
	sql := `SELECT id, title, body, created_at, updated_at FROM posts WHERE id=$1;`
	err := db.QueryRow(context.Background(), sql, id).Scan(&p.ID, &p.Title, &p.Body, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
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

// TODO: implement context
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
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/new/", newHandler)
	http.HandleFunc("/create", createHandler)

	db = getDB()
	defer db.Close(context.Background())

	log.Fatal(http.ListenAndServe(":8080", nil))
}
