// https://golang.org/doc/articles/wiki/
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"text/template"
)

type Page struct {
	Title string
	Body  []byte
}

const (
	VIEW = "/view/"
	EDIT = "/edit/"
	SAVE = "/save/"
)

// regexp groups: /handlerName/title
var validPaths = regexp.MustCompile("^/(view|edit|save)/([a-zA-Z0-9]+)$")

var templates = template.Must(template.ParseFiles("view.html", "edit.html"))

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	// example of m: [/view/TestPage view TestPage]
	m := validPaths.FindStringSubmatch(r.URL.Path)
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

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
	fmt.Println("Server shutdown")
}

// func main() {
// 	page := Page{"TestPage", []byte("This is a test page.")}
// 	writeErr := page.Save()
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
