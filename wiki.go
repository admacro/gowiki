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

const (
	validPathPattern      = "^/(view|edit|save)/([a-zA-Z0-9]+)$"
	staticFilePathPattern = "^/static/([a-zA-Z0-9]+)\\.(css|js|jpg|png)$"
	pageInstancePattern   = "\\[([a-zA-Z0-9]+)\\]"
	frontpage             = "/view/FrontPage"
)

var validPath = regexp.MustCompile(validPathPattern)
var mimePath = regexp.MustCompile(staticFilePathPattern)
var pageInstance = regexp.MustCompile(pageInstancePattern)

// var templates = template.Must(template.ParseFiles("tmpl/view.html", "tmpl/edit.html"))
var templates = template.Must(template.ParseGlob("tmpl/*.html"))

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) BodyHtml() template.HTML {
	return template.HTML(p.Body)
}

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

// http.Handlerfunc implements http.Handler interface
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
	http.Redirect(w, r, frontpage, http.StatusFound)
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

func linkInterPages(body []byte) []byte {
	repl := func(m []byte) []byte {
		s := string(m[1 : len(m)-1])
		pageLink := fmt.Sprintf("<a href='/view/%s'>%s</a>", s, s)
		return []byte(pageLink)
	}
	return pageInstance.ReplaceAllFunc(body, repl)
}

func formatPage(body []byte) []byte {
	return regexp.MustCompile("\n").ReplaceAll(body, []byte("<br>"))
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	page.Body = formatPage(linkInterPages(page.Body))
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
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func main() {
	fmt.Println("Starting server at localhost:8080")

	http.HandleFunc("/", frontpageHandler)
	http.HandleFunc("/static/", mimeHandler)
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
