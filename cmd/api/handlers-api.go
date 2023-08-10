package main

import (
	"encoding/json"
	"fmt"
	"myapp/internal/cards"
	"net/http"
	"strconv"
)

type stripePayLoad struct {
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
}

type jsonResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message, omitempty"`
	Content string `json:"content, omitempty"`
	ID      int    `json:"id, omitempty"`
}

func (app *application) GetPaymentIntent(w http.ResponseWriter, r *http.Request) {
	var payload stripePayLoad

	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	amount, err := strconv.Atoi(payload.Amount)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	card := cards.Card{
		Secret:    app.config.stripe.secret,
		Key:       app.config.stripe.key,
		Currrency: payload.Currency,
	}

	fmt.Println("Esta es la Secret", card.Secret)
	fmt.Println("Esta es la Key", card.Key)
	fmt.Println("Esta es la Currency", card.Currrency)

	okay := true

	pi, msg, err := card.Charge(payload.Currency, amount)
	if err != nil {
		okay = false
	}

	fmt.Println("Ok", okay)
	fmt.Println("pi", pi)

	if okay {
		out, err := json.MarshalIndent(pi, "", "   ")
		if err != nil {
			app.errorLog.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(out)

	} else {

		j := jsonResponse{
			Ok:      false,
			Message: msg,
			Content: "",
		}

		out, err := json.MarshalIndent(j, "", "   ")
		if err != nil {
			app.errorLog.Println(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(out)

	}

}
