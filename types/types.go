package types

import (
	"context"
	"database/sql"
	"time"
)

type ItemStore interface {
	BeginTransaction(ctx context.Context) (*sql.Tx, error)
	CreateItem(item Item) error
	CreateItemType(item_type ItemType) error
	DeleteItemByRFID(rfid_tag string) error
	GetItemByRFIDTag(rfid_tag string) (*Item, error)
	GetItemBySN(serial_num string, tx *sql.Tx, ctx context.Context) (*Item, error)
	GetSoldItemByRFID(rfid_tag string) (*Item, error)
	GetItemByIdSearch(search string) ([]ItemSellingResponse, error)
	GetItems(limit int, offset int, search string, status string) ([]Item ,int, error)
	NewItemSold(sold_item SoldItem, tx *sql.Tx, ctx context.Context) error
	GetItemCount(search string, status string) (int, error)
	GetSoldItemsCount (search string) (int, error)
	GetAllSoldItems(limit int, offset int, search string) ([]SoldItem, int, error)
	// UpdateItemSold(updated_solditem SoldItem) error
	GetItemTypes() ([]ItemType, error)
	ShipItem(item_id int, tx *sql.Tx, ctx context.Context) error
	GetItemsByInvoice (invoice_id int) ([]SoldItem, error)
	GetInvoices (invoice string) ([]Invoice, error)
	CreateInvoice(invoice string, ol_shop string, tx *sql.Tx, ctx context.Context) (int, error)
	ShipInvoice (invoice_id int, tx *sql.Tx, ctx context.Context) error
	GetAllInvoice (limit int, offset int, invoice string, status string) ([]Invoice, int, error)
	EditInvoice(id int, payload EditInvoice) error
	DeleteInvoice(id int, tx *sql.Tx, ctx context.Context) error
	GetInvoiceByID(id int) (*Invoice, error)
	GetItemStatusCount() (int, int, int, error)
	GetItemTypeCount() (map[string]int, error)
	ResetItemsToNotSold(items []SoldItem, tx *sql.Tx, ctx context.Context) error
}

type Item struct {
	ID           int    `json:"id"`
	SerialNumber string    `json:"serial_number"`
	RFIDTag      string `json:"rfid_tag"`
	Batch		 int	`json:"batch"`
	Status		 string	`json:"status"`
	TypeRef		 string	`json:"type_ref"`
	CreatedAt	 time.Time	`json:"createdat"`	
}

type ItemType struct {
	ID			int 	`json:"id"`
	TypeName 	string	`json:"item_type"`
	Price		int		`json:"price"`
}

type TypesResponse struct {
	ItemTypes	[]ItemType	`json:"types"`
}

type ItemsResponse struct {
	Items     []Item `json:"items"`
	ItemCount int    `json:"item_count"`
}

type ItemSellingResponse struct {
	ID           int    `json:"id"`
	SerialNumber string    `json:"serial_number"`
	RFIDTag      string `json:"rfid_tag"`
	TypeRef		 string	`json:"type_ref"`
}

type SoldItemsResponse struct {
	SoldItems 		[]SoldItem	`json:"sold_items"`
	SoldItemsCount	int			`json:"sold_items_count"`	
}

type SoldItem struct {
	ID				int 		`json:"id"`
	ItemID			int 		`json:"item_id"`
	ItemSN			string		`json:"item_sn"`
	ItemTag			string		`json:"item_tag"`
	Status			string		`json:"status"`
	DatetimeSold	time.Time 	`json:"datetime_sold"`
	InvoiceID		int			`json:"invoice_id"`
	OnlineShop		string		`json:"ol_shop"`
	ItemType		string		`json:"item_type"`
}

type RegisterItemPayload struct {
	SerialNumber string    `json:"serial_number" validate:"required"`
	RFIDTag      string `json:"rfid_tag" validate:"required"`
	TypeRef		 string	`json:"type_ref" validate:"required"`
	Batch	 int	`json:"batch" validate:"required"`
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

type ItemTypePayload struct {
	ItemType	string	`json:"item_type" validate:"required"`
	Price		int		`json:"price" validate:"required"`
}

type SoldItemPayload struct {
	Invoice			string		`json:"invoice" validate:"required"`
	OnlineShop	string			`json:"ol_shop" validate:"required"`
}

type EditInvoice struct {
	Invoice		string	`json:"invoice"`
	OnlineShop	string	`json:"ol_shop"`
}

type SoldItemBulkPayload struct {
	SerialNums	[]string		`json:"serial_numbers" validate:"required"`
	Invoice			string		`json:"invoice" validate:"required"`
	OnlineShop	string			`json:"ol_shop" validate:"required"`
}

type ShipItemsPayload struct {
	InvoiceID		int		`json:"invoice_id" validate:"required"`
}

type Invoice struct {
	ID				int			`json:"id"`
	InvoiceStr		string		`json:"invoice_str"`
	Status			string		`json:"status"`
	OnlineShop		string		`json:"online_shop"`
}
type InvoicePayload struct {
	ID			int		`json:"id" validate:"required"`
	InvoiceStr	string	`json:"invoice_str" validate:"required"`
}

type InvoiceItemsResponse struct {
	SoldItems 	[]SoldItem	`json:"sold_items"`
	InvoiceStr	string		`json:"invoice"`
	OnlineShop	string		`json:"online_shop"`
}

type InvoicesResponse struct {
	Invoices 	[]Invoice	`json:"invoices"`
	Count 		int			`json:"count"`
}

type ItemStatusCount struct {
	NotSold		int		`json:"not_sold"`
	SoldPending	int		`json:"sold_pending"`
	SoldShipped	int		`json:"sold_shipped"`
}
