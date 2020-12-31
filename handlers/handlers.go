package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/ianmdawson/go-blog/models"
	"golang.org/x/crypto/bcrypt"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
var (
	_, b, _, _ = runtime.Caller(0)
	// basepath is the package root directory
	basepath = filepath.Dir(b)
)

var templateDir string = basepath + "/../tmpl"

// Templates are the html templates for the blog
var Templates = template.Must(template.ParseGlob(templateDir + "/*.html"))

// PageURIPatterns dictates which paths are available for interacting with the Page model
type PageURIPatterns struct {
	PageEditPath   string
	PageIndexPath  string
	PageNewPath    string
	PageViewPath   string
	PageCreatePath string
	PageSavePath   string
}

// UserURIPatterns disctates which paths are available for the User model
type UserURIPatterns struct {
	UserCreatePath       string
	UserLogInPath        string
	UserAuthenticatePath string
}

// Links makes handling navigation-related logic a little easier
type Links struct {
	PagePatterns PageURIPatterns
	UserPatterns UserURIPatterns
	CurrentRoute string
}

// UserPaths contains all paths for routing and linking
var UserPaths = UserURIPatterns{
	UserCreatePath:       "/users/create",
	UserLogInPath:        "/users/log_in/",
	UserAuthenticatePath: "/users/authenticate/",
}

// PagePaths Returns all page URI pattern prefixes
// PagePaths page paths for routing and linking
var PagePaths = PageURIPatterns{
	PageEditPath:   "/pages/edit/",
	PageIndexPath:  "/",
	PageNewPath:    "/pages/new/",
	PageViewPath:   "/pages/",
	PageCreatePath: "/pages/create/",
	PageSavePath:   "/pages/save/",
}

// LogInHandler renders the view for user to log in
func LogInHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := "log_in"
	links := Links{
		PagePaths,
		UserPaths,
		tmpl,
	}
	var templateData = struct {
		Links Links
	}{
		links,
	}
	err := Templates.ExecuteTemplate(w, tmpl+".html", templateData)
	if err != nil {
		log.Println("Error exectuing template: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// SignUpHandler renders the view for the user sign up form
func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := "sign_up"
	links := Links{
		PagePaths,
		UserPaths,
		tmpl,
	}
	var templateData = struct {
		Links Links
	}{
		links,
	}
	err := Templates.ExecuteTemplate(w, tmpl+".html", templateData)
	if err != nil {
		log.Println("Error exectuing template: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// AuthenticateUserHandler handles sign in authentication
func AuthenticateUserHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "Username or password were empty, try again", http.StatusForbidden)
		return
	}

	u := models.User{}
	err := u.FindByUsername(username)
	if err != nil {
		log.Println("Something went wrong while trying to find the user: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = u.Authenticate(password)
	if err != nil {
		log.Println("Something went wrong while trying to validate user's credentials: ", err)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// TODO: persist auth to session somehow, JWT?
	// https://blog.usejournal.com/authentication-in-golang-c0677bcce1a8

	http.Redirect(w, r, PagePaths.PageIndexPath, http.StatusFound)
}

// CreateUserHandler processes sign up form and creates a new user
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "Username or password were empty, try again", http.StatusUnprocessableEntity)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Could not generate hash from password: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	u := models.User{
		Username: username,
		Password: hash,
		Role:     "user",
	}
	err = u.Create()
	if err != nil {
		log.Println("Something went wrong while trying to create the user: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, PagePaths.PageIndexPath, http.StatusFound)
}

// NewPage renders the new page template for users to create a new Page
func NewPage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "new", &models.Page{})
}

// ViewPage renders the Page if the given ID exists, otherwise it redirects to NewPage
func ViewPage(w http.ResponseWriter, r *http.Request) {
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

// some shared template rendering logic for simple New/Edit/View templates
func renderTemplate(w http.ResponseWriter, tmpl string, p *models.Page) {
	links := Links{
		PagePaths,
		UserPaths,
		tmpl,
	}
	var templateData = struct {
		Page  *models.Page
		Links Links
	}{
		p,
		links,
	}
	err := Templates.ExecuteTemplate(w, tmpl+".html", templateData)
	if err != nil {
		fmt.Println("Error exectuing template: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// IndexHandler renders the index Page index page, the most recent Page and a list of other most recent pages
func IndexHandler(w http.ResponseWriter, r *http.Request) {
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

	const tmpl string = "index"

	indexData := struct {
		Page           *models.Page
		PageCollection *models.PageCollection
		Links          Links
	}{
		firstPage,
		pageCollection,
		Links{
			PagePaths,
			UserPaths,
			tmpl,
		},
	}
	err = Templates.ExecuteTemplate(w, tmpl+".html", indexData)
	if err != nil {
		fmt.Println("500", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
