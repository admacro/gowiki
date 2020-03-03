// https://golang.org/doc/articles/wiki/
package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
}

const (
	VIEW      = "/view/"
	EDIT      = "/edit/"
	SAVE      = "/save/"
	MIME_PATH = "/static/"

	STATIC_FILE_PATTERN = "^/static/([a-zA-Z0-9]+)\\.(css|js|jpg|png)$"

	FRONTPAGE = "/view/FrontPage"
)

// regexp groups: /handlerName/title
var validPath = regexp.MustCompile("^/(view|edit|save)/([a-zA-Z0-9]+)$")
var mimePath = regexp.MustCompile(STATIC_FILE_PATTERN)

// var templates = template.Must(template.ParseFiles("tmpl/view.html", "tmpl/edit.html"))
var templates = template.Must(template.ParseGlob("tmpl/*.html"))

func (p *Page) save() error {
	return ioutil.WriteFile(getDataFilename(p.Title), p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	body, err := ioutil.ReadFile(getDataFilename(title))
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func getDataFilename(title string) string {
	return fmt.Sprintf("data/%s.txt", title)
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	// example of m: [/view/TestPage view TestPage]
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("invalid page title")
	}
	fmt.Printf("URL matched: %v\n", m)
	return m[2], nil
}

func renderTemplate(w http.ResponseWriter, page *Page, tmplName string) {
	err := templates.ExecuteTemplate(w, tmplName+".html", page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(w http.ResponseWriter, r *http.Request, title string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handling URL: %s\n", r.URL.Path)
		title, err := getTitle(w, r)
		if err != nil {
			return
		}
		fn(w, r, title)
	}
}

func frontpageHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, FRONTPAGE, http.StatusFound)
}

func mimeHandler(w http.ResponseWriter, r *http.Request) {
	m := mimePath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
	}
	fmt.Printf("URL matched: %v\n", m)
	mimeFilePath := m[0]
	mimeExtention := m[2]
	mimeFile, err := ioutil.ReadFile(mimeFilePath[1:])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", mime.TypeByExtension(mimeExtention))
	w.Write(mimeFile)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, EDIT+title, http.StatusFound)
		return
	}

	renderTemplate(w, page, "view")
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		page = &Page{Title: title}
	}

	renderTemplate(w, page, "edit")
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	page := &Page{Title: title, Body: []byte(r.FormValue("body"))}
	err := page.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, VIEW+title, http.StatusFound)
}

func main() {
	fmt.Println("Starting server at localhost:8080")

	http.HandleFunc("/", frontpageHandler)
	http.HandleFunc(MIME_PATH, mimeHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
	fmt.Println("Server shutdown")
}

// func main() {
// 	page := Page{"TestPage", []byte("This is a test page.")}
// 	writeErr := page.save()
// 	if writeErr != nil {
// 		panic(fmt.Sprintf("Error saving file: \n%v", writeErr.Error()))
// 	}

// 	p, loadErr := loadPage("TestPage")
// 	if loadErr != nil {
// 		panic(fmt.Sprintf("Error reading file: \n%v", loadErr.Error()))
// 	}
// 	fmt.Printf("Page Title: %v\n", p.Title)
// 	fmt.Printf("Page Body: %v\n", string(p.Body))
// }
