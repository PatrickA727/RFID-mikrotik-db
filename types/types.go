package types

import (
	"context"
	"database/sql"
	"time"
)

type ItemStore interface {
	BeginTransaction(ctx context.Context) (*sql.Tx, error)
	CreateItem(item Item) error
	DeleteItemBySN(serial_num int) error
	GetItemByRFIDTag(rfid_tag string) (*Item, error)
	GetItems(limit int, offset int, search string) ([]Item ,int, error)
	CreateWarranty(warranty Warranty,  ctx context.Context) error
	GetWarrantyByItemId(item_id int) (*Warranty, error)
	GetAllWarranty(limit int, offset int, search string) ([]Warranty, error)
	NewItemSold(sold_item SoldItem, ctx context.Context) error
	GetItemCount(search string) (int, error)
	GetAllSoldItems(limit int, offset int, search string) ([]SoldItem, error)
}

type Item struct {
	ID           int    `json:"id"`
	SerialNumber string    `json:"serial_number"`
	RFIDTag      string `json:"rfid_tag"`
	ItemName     string `json:"item_name"`
	Price		 int	`json:"price"`
	Warranty	 string `json:"warranty"`
	Sold 		 bool	`json:"sold"`
}

type ItemsResponse struct {
	Items     []Item `json:"items"`
	ItemCount int    `json:"item_count"`
}

type Warranty struct {
	ID			int	`json:"id"`
	ItemID		int	`json:"item_id"`
	ItemSN		string	`json:"item_sn"`
	PurchaseDate	time.Time	`json:"purchase_date"`
	Expiration	time.Time	`json:"expiration"`
	CustName	string	`json:"cust_name"`
	CustEmail	string	`json:"cust_email"`
	CustPhone	string	`json:"cust_phone"`
}

type SoldItem struct {
	ID				int 		`json:"id"`
	ItemID			int 		`json:"item_id"`
	ItemSN			string		`json:"item_sn"`
	DatetimeSold	time.Time 	`json:"datetime_sold"`
	Invoice			string		`json:"invoice"`
	PaymentMethod	string		`json:"payment_method"`
	PaymentStatus	string		`json:"payment_status"`
}

type RegisterItemPayload struct {
	SerialNumber string    `json:"serial_number" validate:"required"`
	RFIDTag      string `json:"rfid_tag" validate:"required"`
	ItemName     string `json:"item_name" validate:"required"`
	Price		 int	`json:"price" validate:"required"`
}

type NewWarrantyPayload struct {
	PurchaseDate	string	`form:"purchase_date" validate:"required"`
	CustName		string	`form:"cust_name" validate:"required"`
	CustEmail		string	`form:"cust_email" validate:"required,email"`
	CustPhone		string	`form:"cust_phone"`
}

type GetItemAndWarrantyPayload struct {
	RFIDTag	string	`json:"rfid_tag" validate:"required"`
}

type SoldItemPayload struct {
	Invoice			string		`json:"invoice" validate:"required"`
	PaymentMethod	string		`json:"payment_method" validate:"required"`
	PaymentStatus	string		`json:"payment_status" validate:"required"`
}
