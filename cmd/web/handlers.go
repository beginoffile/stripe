package main

import "net/http"

func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	if err := app.rederTemplate(w, r, "terminal", nil); err != nil {
		app.errorLog.Println(err)

	}
}
