package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
    "path/filepath"
)

type Page struct {
    Title string
	Raw string
	Search string
    Body template.HTML
}

func (p *Page) save() error {
    return os.WriteFile("data/" + p.Title + ".txt", []byte(p.Raw), 0600)
}

var linksRe = regexp.MustCompile("\\[[a-zA-Z]+\\]")

func loadPage(title string) (*Page, error) {
	raw, err := os.ReadFile("data/" + title + ".txt")
	if err != nil {
		return nil, err
	}
	body := template.HTMLEscapeString(string(raw))
	body = linksRe.ReplaceAllStringFunc(body, func(s string) string {
		m := s[1 : len(s)-1]
		return fmt.Sprintf("<a href=\"/view/%s\">%s</a>", m, m)
	})
	return &Page{Title: title, Raw: string(raw), Body: template.HTML(body)}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
	if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
	renderTemplate(w, "view", p)
}

func viewHomeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/home", http.StatusFound)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    p := &Page{Title: title, Raw: r.FormValue("raw")}
    err := p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func searchHandler(w http.ResponseWriter, r *http.Request, title string) {
	// trouver comment recuperer l' input !!!

	filenames, err := filepath.Glob("data/" + title + ".txt")
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
	return
}
	for _, f := range filenames {
		f = title
		http.Redirect(w, r, "/view/"+f, http.StatusFound)
	}
}

var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view|search)/([a-zA-Z0-9]+)$")

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/", viewHomeHandler)
	http.HandleFunc("/search/", makeHandler(searchHandler))

	log.Println("starting server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
