package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ianmdawson/go-blog/models"
)

func TestNewPageHandler(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "pages/new", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NewPage)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expectedStrings := []string{"<h1>New</h1>", "Title", "Content", "<button type=\"submit\"", "id=\"submit-page-button\""}
	for _, expected := range expectedStrings {
		if strings.Contains(rr.Body.String(), expected) != true {
			t.Errorf("handler returned unexpected body: %v \n want: %v",
				rr.Body.String(), expected)
		}
	}
}

func TestIndexHandler(t *testing.T) {
	models.TSetUpDB()
	defer models.TTearDown()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(IndexHandler)

	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expectedStrings := []string{"Recent Posts", "No pages yet"}
	for _, expected := range expectedStrings {
		if strings.Contains(rr.Body.String(), expected) != true {
			t.Errorf("handler returned unexpected body: %v \n want: %v",
				rr.Body.String(), expected)
		}
	}
}
