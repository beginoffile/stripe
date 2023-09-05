package models

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// DBModel is the type for database connection values
type DBModel struct {
	DB *sql.DB
}

// Models is the wrapper for all models
type Models struct {
	DB DBModel
}

// NewModels returns a model type with database connection pool
func NewModel(db *sql.DB) Models {
	return Models{
		DB: DBModel{DB: db},
	}
}

// Widget is the type for all widgets
type Widget struct {
	ID             int       `json: "id"`
	Name           string    `json: "name"`
	Description    string    `json: "description"`
	InventoryLevel int       `json:"inventory_level"`
	Price          int       `json:"price"`
	Image          string    `json:"image"`
	IsRecurring    bool      `json:"is_recurring"`
	PlanID         string    `json:"plan_id"`
	CreateAt       time.Time `json:"_"`
	UpdatedAt      time.Time `json:"_"`
}

// Order is the type for all orders
type Order struct {
	ID            int         `json:"id"`
	WidgetID      int         `json:"widget_id"`
	TransactionID int         `json:"transaction_id"`
	CustomerID    int         `json:"customers_id"`
	StatusID      int         `json:"status_id"`
	Quantity      int         `json:"quantity"`
	Amount        int         `json:"amount"`
	CreatedAt     time.Time   `json:"-"`
	UpdatedAt     time.Time   `json:"-"`
	Widget        Widget      `json:"widget"`
	Transaction   Transaction `json:"transaction"`
	Customer      Customer    `json:"customer"`
}

// status is the type for statuses
type Status struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Transactionstatus is the type for transaction statuses
type TransactionStatus struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Transaction is the type for transactions
type Transaction struct {
	ID                  int       `json:"id"`
	Amount              int       `json:"amount"`
	Currency            string    `json:"currency"`
	LastFour            string    `json:"last_four"`
	ExpiryMonth         int       `json:"expiry_month"`
	ExpiryYear          int       `json:"expiry_year"`
	PaymentIntent       string    `json:"payment_intent"`
	PaymentMethod       string    `json:"payment_method"`
	BankReturnCode      string    `json:"bank_return_code"`
	TransactionStatusID int       `json:"transaction_status_id"`
	CreatedAt           time.Time `json:"-"`
	UpdatedAt           time.Time `json:"-"`
}

// User is the type for users
type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Customer is the type for customers
type Customer struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (m *DBModel) GetWidget(id int) (Widget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var widget Widget

	row := m.DB.QueryRowContext(ctx,
		`select id, 			name, 
			description, 	inventory_level, 
			price, 			coalesce(image,''), 
			is_recurring,   plan_id,
			created_at, 	updated_at
	from widgets 
	where id = ?`, id)
	err := row.Scan(
		&widget.ID, &widget.Name,
		&widget.Description, &widget.InventoryLevel,
		&widget.Price, &widget.Image,
		&widget.IsRecurring, &widget.PlanID,
		&widget.CreateAt, &widget.UpdatedAt)

	if err != nil {
		return widget, err
	}

	return widget, nil

}

// InsertTransaction inserts  a new txn, and return its id
func (m *DBModel) InsertTransaction(txn Transaction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	stmt := `
	Insert into transactions
	(amount, 			currency, 				last_four, 
	bank_return_code,	expiry_month,			expiry_year,
	payment_intent,		payment_method,			transaction_status_id, 	
	created_at, 		updated_at
	)
	values
	(?, 				?,						?,
	?,					?,						?,
	?,					?,						?,
	?,					?)
	`
	result, err := m.DB.ExecContext(ctx, stmt,
		txn.Amount, txn.Currency, txn.LastFour,
		txn.BankReturnCode, txn.ExpiryMonth, txn.ExpiryYear,
		txn.PaymentIntent, txn.PaymentMethod, txn.TransactionStatusID,
		time.Now(), time.Now())

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil

}

// InsertOrder inserts  a new order, and return its id
func (m *DBModel) InsertOrder(order Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	stmt := `
	Insert into orders
	(widget_id, 		transaction_id, 				status_id, 
	quantity,			customer_id,					amount, 						
	created_at, 		updated_at
	)
	values
	(?, 				?,						?,
	?,					?,						?,
	?,					?)
	`
	result, err := m.DB.ExecContext(ctx, stmt,
		order.WidgetID, order.TransactionID, order.StatusID,
		order.Quantity, order.CustomerID, order.Amount,
		time.Now(), time.Now())

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil

}

// InsertOrder inserts  a new order, and return its id
func (m *DBModel) InsertCustomer(c Customer) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	stmt := `
	Insert into customers
	(first_name, 		last_name, 				email, 
	created_at, 		updated_at
	)
	values
	(?, 				?,						?,
	?,					?)
	`
	result, err := m.DB.ExecContext(ctx, stmt,
		c.FirstName, c.LastName, c.Email,
		time.Now(), time.Now())

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil

}

// GetUserByEmail getsa user by email address
func (m *DBModel) GetUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	email = strings.ToLower(email)
	var u User
	row := m.DB.QueryRowContext(ctx, `
	select 	id, 			first_name, 
			last_name, 		email, 
			password, 		created_at, 	
			updated_at
	From Users 
	Where email = ?`, email)

	err := row.Scan(&u.ID, &u.FirstName,
		&u.LastName, &u.Email,
		&u.Password, &u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return u, err
	}

	return u, nil

}

func (m *DBModel) Authenticate(email, password string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, `
	Select t1.id, t1.password
	From users t1
	Where t1.email = ?`, email)

	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, errors.New("Incorrect Password")
	}
	if err != nil {
		return 0, err
	}

	return id, nil

}

func (m *DBModel) UpdatePasswordForUser(u User, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	update users 
	set 
	password=? 
	Where id = ?`
	_, err := m.DB.ExecContext(ctx, stmt, hash, u.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) GetAllOrders() ([]*Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var orders []*Order

	query := `
	SELECT 	t1.id, 				t1.widget_id, 	t1.transaction_id, 
			t1.customer_id, 	t1.status_id, 	t1.quantity, 
			t1.amount, 			t1.created_at, 	t1.updated_at,
			t2.name,			t3.currency,	t3.last_four,		
			t3.expiry_month, 	t3.expiry_year,	t3.payment_intent, 	
			t3.bank_return_code,t4.first_name, 	t4.last_name, 		
			t4.email
	From orders t1
		Left JOIN widgets t2
		  On t2.id = t1.widget_id
		LEFT JOIN transactions t3
		  On t3.id = t1.transaction_id
		LEFT JOIN customers t4
		  On t4.id = t1.customer_id
	ORDER BY t1.created_at DESC`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var o Order
		err = rows.Scan(
			&o.ID, &o.WidgetID, &o.TransactionID,
			&o.CustomerID, &o.StatusID, &o.Quantity,
			&o.Amount, &o.CreatedAt, &o.UpdatedAt,
			&o.Widget.Name, &o.Transaction.Currency, &o.Transaction.LastFour,
			&o.Transaction.ExpiryMonth, &o.Transaction.ExpiryYear, &o.Transaction.PaymentIntent,
			&o.Transaction.BankReturnCode, &o.Customer.FirstName, &o.Customer.LastName,
			&o.Customer.Email,
		)

		if err != nil {
			return nil, err
		}

		orders = append(orders, &o)

	}

	return orders, nil

}
