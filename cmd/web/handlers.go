package main

import (
	"fmt"
	"myapp/internal/cards"
	"myapp/internal/encryption"
	"myapp/internal/models"
	urlsinger "myapp/internal/urlsigner"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// Home Display the home page
func (app *application) Home(w http.ResponseWriter, r *http.Request) {

	// stringMap := make(map[string]string)
	// stringMap["publishable_key"] = app.config.stripe.key

	if err := app.renderTemplate(w, r, "home", &templateData{
		// StringMap: stringMap,
	}); err != nil {
		app.errorLog.Println(err)

	}
}

func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {

	// stringMap := make(map[string]string)
	// stringMap["publishable_key"] = app.config.stripe.key

	if err := app.renderTemplate(w, r, "terminal", &templateData{
		// StringMap: stringMap,
	}); err != nil {
		app.errorLog.Println(err)

	}
}

type TransactionData struct {
	FirstName       string
	LastName        string
	Email           string
	PaymentIntentID string
	PaymentMethodID string
	PaymentAmount   int
	PaymentCurrency string
	LastFour        string
	ExpiryMonth     int
	ExpiryYear      int
	BankReturnCode  string
}

// GetTransactionData gets txn Data from post and stripe
func (app *application) GetTransactionData(r *http.Request) (TransactionData, error) {
	var txnData TransactionData

	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}
	//read posted data
	firstName := r.Form.Get("first_name")
	lastName := r.Form.Get("last_name")
	email := r.Form.Get("email")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")

	amount, _ := strconv.Atoi(paymentAmount)

	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.key,
	}

	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	txnData = TransactionData{
		FirstName:       firstName,
		LastName:        lastName,
		Email:           email,
		PaymentIntentID: paymentIntent,
		PaymentMethodID: paymentMethod,
		PaymentAmount:   amount,
		PaymentCurrency: paymentCurrency,
		LastFour:        lastFour,
		ExpiryMonth:     int(expiryMonth),
		ExpiryYear:      int(expiryYear),
		BankReturnCode:  pi.LatestCharge.ID,
	}

	return txnData, nil

}

// PaymentSucceeded display the recipient page
func (app *application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	widgetID, _ := strconv.Atoi(r.Form.Get("product_id"))

	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//create a new customer
	customerID, err := app.SaveCustomer(txnData.FirstName, txnData.LastName, txnData.Email)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//create a new transaction

	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		BankReturnCode:      txnData.BankReturnCode,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
		TransactionStatusID: 2,
	}

	txnID, err := app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//create a new order
	order := models.Order{
		WidgetID:      widgetID,
		TransactionID: txnID,
		CustomerID:    customerID,
		StatusID:      1,
		Quantity:      1,
		Amount:        txnData.PaymentAmount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = app.SaveOrder(order)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// data := make(map[string]interface{})

	// data["email"] = email
	// data["pi"] = paymentIntent
	// data["pm"] = paymentMethod
	// data["pa"] = paymentAmount
	// data["pc"] = paymentCurrency

	// data["last_four"] = lastFour
	// data["expiry_month"] = expiryMonth
	// data["expiry_year"] = expiryYear
	// data["bank_return_code"] = pi.LatestCharge.ID
	// data["first_name"] = firstName
	// data["last_name"] = lastName

	//write this data to session, and the redirect user to new page

	app.Session.Put(r.Context(), "receipt", txnData)

	/*
		if err := app.renderTemplate(w, r, "succeeded", &templateData{
			Data: data,
		}); err != nil {
			app.errorLog.Println(err)
		}
	*/
	http.Redirect(w, r, "/receipt", http.StatusSeeOther)

}

// VirtualTerminalPaymentSucceeded display the recipient page for vitual terminal transaction
func (app *application) VirtualTerminalPaymentSucceeded(w http.ResponseWriter, r *http.Request) {

	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//create a new transaction

	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		BankReturnCode:      txnData.BankReturnCode,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
		TransactionStatusID: 2,
	}

	_, err = app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// data := make(map[string]interface{})

	// data["email"] = email
	// data["pi"] = paymentIntent
	// data["pm"] = paymentMethod
	// data["pa"] = paymentAmount
	// data["pc"] = paymentCurrency

	// data["last_four"] = lastFour
	// data["expiry_month"] = expiryMonth
	// data["expiry_year"] = expiryYear
	// data["bank_return_code"] = pi.LatestCharge.ID
	// data["first_name"] = firstName
	// data["last_name"] = lastName

	//write this data to session, and the redirect user to new page

	app.Session.Put(r.Context(), "receipt", txnData)

	/*
		if err := app.renderTemplate(w, r, "succeeded", &templateData{
			Data: data,
		}); err != nil {
			app.errorLog.Println(err)
		}
	*/
	http.Redirect(w, r, "/virtual-terminal-receipt", http.StatusSeeOther)

}

func (app *application) Receipt(w http.ResponseWriter, r *http.Request) {
	// data := app.Session.Get(r.Context(), "receipt").(map[string]interface{})
	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn

	app.Session.Remove(r.Context(), "receipt")
	if err := app.renderTemplate(w, r, "receipt", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}

}

func (app *application) VirtualTerminalReceipt(w http.ResponseWriter, r *http.Request) {
	// data := app.Session.Get(r.Context(), "receipt").(map[string]interface{})
	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn

	app.Session.Remove(r.Context(), "receipt")
	if err := app.renderTemplate(w, r, "virtual-terminal-receipt", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}

}

// SaveCustomer saves the customer and returns id
func (app *application) SaveCustomer(firstName, lastName, email string) (int, error) {
	customer := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	id, err := app.DB.InsertCustomer(customer)
	if err != nil {
		return 0, err
	}

	return id, nil

}

// SaveTransaction saves a txn and returns id
func (app *application) SaveTransaction(txn models.Transaction) (int, error) {

	id, err := app.DB.InsertTransaction(txn)
	if err != nil {
		return 0, err
	}

	return id, nil

}

// SaveOrder saves a order and returns id
func (app *application) SaveOrder(order models.Order) (int, error) {

	id, err := app.DB.InsertOrder(order)
	if err != nil {
		return 0, err
	}

	return id, nil

}

func (app *application) ChargeOne(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	widgetID, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidget(widgetID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	err = app.renderTemplate(w, r, "buy-once", &templateData{
		Data: data,
	}, "stripe-js")

	if err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) BronzePlan(w http.ResponseWriter, r *http.Request) {

	widget, err := app.DB.GetWidget(2)

	if err != nil {
		app.errorLog.Println(err)
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	// intMap := make(map[string]int)

	// intMap["plan_id"] = 1

	err = app.renderTemplate(w, r, "bronze-plan", &templateData{
		Data: data,
	})

	if err != nil {
		app.errorLog.Println(err)
	}

}

// BronzePlanReceipt display the receipt for bronze Plan
func (app *application) BronzePlanReceipt(w http.ResponseWriter, r *http.Request) {

	err := app.renderTemplate(w, r, "receipt-plan", &templateData{})

	if err != nil {
		app.errorLog.Println(err)
	}

}

// LoginPage display the login page
func (app *application) LoginPage(w http.ResponseWriter, r *http.Request) {
	err := app.renderTemplate(w, r, "login", &templateData{})

	if err != nil {
		app.errorLog.Println(err)
	}

}

func (app *application) PostLoginPage(w http.ResponseWriter, r *http.Request) {
	app.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	id, err := app.DB.Authenticate(email, password)

	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	app.Session.Put(r.Context(), "userID", id)

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (app *application) Logout(w http.ResponseWriter, r *http.Request) {
	app.Session.Destroy(r.Context())
	app.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *application) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	err := app.renderTemplate(w, r, "forgot-password", &templateData{})

	if err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) ShowResetPassword(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	theURL := r.RequestURI
	testURL := fmt.Sprintf("%s%s", app.config.frontend, theURL)

	signer := urlsinger.Signer{
		Secret: []byte(app.config.secretkey),
	}
	valid := signer.VerifyToken(testURL)

	if !valid {
		app.errorLog.Println("Invalid url - tampering detected")
		return
	}

	//make sure not expired
	expired := signer.Expired(testURL, 60)
	if expired {
		app.errorLog.Println("Link Expired")
		return
	}

	encryptor := encryption.Encryption{
		Key: []byte(app.config.secretkey),
	}

	encryptedEmail, err := encryptor.Encrypt(email)
	if err != nil {
		app.errorLog.Println("Encryption Failed", err)
	}

	data := make(map[string]interface{})
	data["email"] = encryptedEmail

	err = app.renderTemplate(w, r, "reset-password", &templateData{Data: data})

	if err != nil {
		app.errorLog.Println(err)
	}

}

func (app *application) AllSales(w http.ResponseWriter, r *http.Request) {

	err := app.renderTemplate(w, r, "all-sales", &templateData{})

	if err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) AllSubscriptions(w http.ResponseWriter, r *http.Request) {

	err := app.renderTemplate(w, r, "all-subscriptions", &templateData{})
	if err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) ShowSale(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["title"] = "Sale"
	stringMap["cancel"] = "/admin/all-sales"
	err := app.renderTemplate(w, r, "sale", &templateData{StringMap: stringMap})
	if err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) ShowSubscription(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["title"] = "Subscription"
	stringMap["cancel"] = "/admin/all-subscriptions"
	err := app.renderTemplate(w, r, "sale", &templateData{StringMap: stringMap})
	if err != nil {
		app.errorLog.Println(err)
	}
}
