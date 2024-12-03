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
	GetItems(limit int, offset int, search string) ([]Item ,int, error)
	CreateWarranty(warranty Warranty,  ctx context.Context) error
	GetWarrantyByItemId(item_id int) (*Warranty, error)
	GetAllWarranty(limit int, offset int, search string) ([]Warranty, int, error)
	NewItemSold(sold_item SoldItem, ctx context.Context) error
	GetItemCount(search string) (int, error)
	GetWarrantyCount(search string) (int, error)
	GetSoldItemsCount (search string) (int, error)
	GetAllSoldItems(limit int, offset int, search string) ([]SoldItem, int, error)
	UpdateItemSold(updated_solditem SoldItem) error
	GetItemTypes() ([]ItemType, error)
}

type Item struct {
	ID           int    `json:"id"`
	SerialNumber string    `json:"serial_number"`
	RFIDTag      string `json:"rfid_tag"`
	ItemName     string `json:"item_name"`
	Warranty	 string `json:"warranty"`
	Sold 		 bool	`json:"sold"`
	Modal		 int 	`json:"modal"`
	Keuntungan	 int 	`json:"keuntungan"`
	Quantity	 int 	`json:"quantity"`
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
	TypeCount	int			`json:"count"`
}

type ItemsResponse struct {
	Items     []Item `json:"items"`
	ItemCount int    `json:"item_count"`
}

type WarrantyResponse struct {
	Warranties     []Warranty `json:"warranties"`
	WarrantyCount int    `json:"warranty_count"`
} 

type SoldItemsResponse struct {
	SoldItems 		[]SoldItem	`json:"sold_items"`
	SoldItemsCount	int			`json:"sold_items_count"`	
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
	OnlineShop		string		`json:"ol_shop"`
	PaymentStatus	bool		`json:"payment_status"`
	Journal			bool 		`json:"journal"`
}

type RegisterItemPayload struct {
	SerialNumber string    `json:"serial_number" validate:"required"`
	RFIDTag      string `json:"rfid_tag" validate:"required"`
	ItemName     string `json:"item_name" validate:"required"`
	Quantity	 int	`json:"quantity" validate:"required"`
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

type SoldItemBulkPayload struct {
	ItemTags	[]string	`json:"item_tags" validate:"required"`
	Invoice			string		`json:"invoice" validate:"required"`
	OnlineShop	string			`json:"ol_shop" validate:"required"`
}
