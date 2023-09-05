package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type templateData struct {
	StringMap            map[string]string
	IntMap               map[string]int
	FloatMap             map[string]float32
	Data                 map[string]interface{}
	CSRFToken            string
	Flash                string
	Warnign              string
	Error                string
	IsAuthenticated      int
	API                  string
	CSSVersion           string
	StripeSecretKey      string
	StripePublishableKey string
}

var functions = template.FuncMap{
	"formatCurrency": formatCurrency,
}

func formatCurrency(n int) string {
	//f := float32(n * 1.0 / 100.0)
	f := float32(n) / float32(100)
	return fmt.Sprintf("$%.2f", f)
}

//go:embed templates

var templateFS embed.FS

func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	td.API = app.config.api
	td.StripeSecretKey = app.config.stripe.secret
	td.StripePublishableKey = app.config.stripe.key

	if app.Session.Exists(r.Context(), "userID") {
		td.IsAuthenticated = 1
	} else {
		td.IsAuthenticated = 0
	}
	return td
}

func (app *application) renderTemplate(w http.ResponseWriter, r *http.Request, page string, td *templateData, partials ...string) error {
	var t *template.Template
	var err error
	templateToRender := fmt.Sprintf("templates/%s.page.tmpl", page)

	_, templateInMap := app.templateCache[templateToRender]

	if app.config.env == "production" && templateInMap {

		t = app.templateCache[templateToRender]
	} else {

		t, err = app.parseTemplate(partials, page, templateToRender)

		if err != nil {
			app.errorLog.Println(err)
			return err
		}
	}

	if td == nil {
		td = &templateData{}
	}

	td = app.addDefaultData(td, r)

	/*
		err = t.Execute(w, td)
		if err != nil {
			app.errorLog.Println(err)
			return err
		}
	*/

	buf := new(bytes.Buffer)
	err = t.Execute(buf, td)
	if err != nil {
		log.Println(err)
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		fmt.Println("Error writing template to browser", err)
		return err
	}

	return nil
}

/*
	func (app *application) parseTemplate(partials []string, page, templateToRender string) (*template.Template, error) {
		var t *template.Template
		var err error

		//build partials
		if len(partials) > 0 {
			for i, x := range partials {
				partials[i] = fmt.Sprintf("templates/%s.partial.tmpl", x)
			}
		}
		if len(partials) > 0 {
			t, err = template.New(fmt.Sprintf("%s.page.tmpl", page)).Funcs(functions).ParseFS(templateFS, "templates/base.layout.tmpl", strings.Join(partials, ","), templateToRender)
		} else {
			t, err = template.New(fmt.Sprintf("%s.page.tmpl", page)).Funcs(functions).ParseFS(templateFS, "templates/base.layout.tmpl", templateToRender)

		}

		if err != nil {
			app.errorLog.Println(err)
			return nil, err
		}

		app.templaceCache[templateToRender] = t

		return t, nil
	}
*/
func (app *application) parseTemplate(partials []string, page, templateToRender string) (*template.Template, error) {
	var t *template.Template
	var err error

	// build partials
	if len(partials) > 0 {
		for i, x := range partials {
			partials[i] = fmt.Sprintf("templates/%s.partial.tmpl", x)
		}
	}

	if len(partials) > 0 {
		t, err = template.New(fmt.Sprintf("%s.page.tmpl", page)).Funcs(functions).ParseFS(templateFS, "templates/base.layout.tmpl", strings.Join(partials, ","), templateToRender)
	} else {
		t, err = template.New(fmt.Sprintf("%s.page.tmpl", page)).Funcs(functions).ParseFS(templateFS, "templates/base.layout.tmpl", templateToRender)

	}
	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}

	/*
		page = fmt.Sprintf("%s.page.tmpl", page)

		templateToRender = "./cmd/web/" + templateToRender
		layout1 := "./cmd/web/templates/base.layout.tmpl"
		t, err = template.New(page).Funcs(functions).ParseFiles(layout1, templateToRender)
		if err != nil {
			app.errorLog.Println(err)
			return nil, err
		}
	*/

	app.templateCache[templateToRender] = t
	return t, nil
}
