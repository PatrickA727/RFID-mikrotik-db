package item

import (
	"context"
	"database/sql"
	"log"

	"github.com/PatrickA727/mikrotik-db-sys/types"
	_ "github.com/jackc/pgx/v5"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    return tx, nil
}

func (s *Store) CreateItem(item types.Item) error {
	_, err := s.db.Exec("INSERT INTO items (serial_number, rfid_tag, item_name, quantity, batch) VALUES ($1, $2, $3, $4, $5)", 
						item.SerialNumber, item.RFIDTag, item.ItemName, item.Quantity, item.Batch,
					);
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreateItemType(item_type types.ItemType) error {
	_, err := s.db.Exec("INSERT INTO item_type (item_type, price) VALUES ($1, $2)", item_type.TypeName, item_type.Price)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetItemTypes() ([]types.ItemType, error) {
	var item_types []types.ItemType

	rows, err := s.db.QueryContext(context.Background(), "SELECT * FROM item_type");
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var item_type types.ItemType

		if err := rows.Scan(&item_type.ID, &item_type.TypeName, &item_type.Price); err != nil {
			return nil, err
		}

		item_types = append(item_types, item_type)
	}
	 
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return item_types, nil
}

func (s *Store) GetItemByRFIDTag(rfid_tag string) (*types.Item, error) {
	var item types.Item

	err := s.db.QueryRow("SELECT id, serial_number, rfid_tag, item_name, sold, modal, keuntungan, quantity, batch FROM items WHERE rfid_tag = $1", rfid_tag).Scan(
		&item.ID, &item.SerialNumber, &item.RFIDTag, &item.ItemName, &item.Sold, &item.Modal, &item.Keuntungan, &item.Quantity, &item.Batch,
	)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *Store) GetItemById(item_id int) (*types.Item, error) {
	var item types.Item

	err := s.db.QueryRow("SELECT id, serial_number, rfid_tag, item_name FROM items WHERE id = $1", item_id).Scan(
		&item.ID, &item.SerialNumber, &item.RFIDTag, &item.ItemName,
	)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *Store) GetItems(limit int, offset int, search string) ([]types.Item ,int, error) {
	var (
		 rows *sql.Rows
		 err error
	)

	itemCount, err := s.GetItemCount(search)
	if err != nil {
		return nil, 0, err
	}

	if search != "" {

		searchPattern := "%" + search + "%"

		rows, err = s.db.QueryContext(context.Background(), 
			"SELECT id, serial_number, rfid_tag, item_name, warranty, sold, modal, keuntungan, quantity, batch, status, type_ref, createdat FROM items WHERE serial_number ILIKE $1 OR rfid_tag ILIKE $1 OR item_name ILIKE $1 ORDER BY batch DESC LIMIT $2 OFFSET $3", searchPattern, limit, offset,
		)
		if err != nil {
			return nil, 0, err
		}
	} else {
		rows, err = s.db.QueryContext(context.Background(), "SELECT id, serial_number, rfid_tag, item_name, warranty, sold, modal, keuntungan, quantity, batch, status, type_ref, createdat FROM items ORDER BY batch DESC LIMIT $1 OFFSET $2", limit, offset)
		if err != nil {
			return nil, 0, err
		}
	}

	defer rows.Close()	// Close rows database connection after finish processing the data/function

	var items []types.Item

	for rows.Next() {
		var item types.Item

		if err := rows.Scan(&item.ID, &item.SerialNumber, &item.RFIDTag, &item.ItemName, &item.Warranty, &item.Sold, &item.Modal, &item.Keuntungan, &item.Quantity, &item.Batch, &item.Status, &item.TypeRef, &item.CreatedAt); err != nil {
			return nil, 0, err
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
        return nil, 0, err
    }

    return items, itemCount, nil
}

func (s *Store) GetItemCount(search string) (int, error) {
	itemCount := 0
	if (search == "") {
		err := s.db.QueryRowContext(context.Background(), 
		"SELECT COUNT(*) FROM items", 
	).Scan(&itemCount)
		if err != nil {
			return 0, err
		}
	} else {
		searchPattern := "%" + search + "%"

		err := s.db.QueryRowContext(context.Background(), 
		"SELECT COUNT(*) FROM items WHERE serial_number ILIKE $1 OR rfid_tag ILIKE $1 OR item_name ILIKE $1", searchPattern, 
	).Scan(&itemCount)
		if err != nil {
			return 0, err
		}
	}

	return itemCount, nil
}

func (s *Store) DeleteItemByRFID(rfid_tag string) error {
	_, err := s.db.Exec("DELETE FROM items WHERE rfid_tag = $1", rfid_tag)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) CreateWarranty(warranty types.Warranty, ctx context.Context) error {
	tx, err := s.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	// Ensure transaction rollback on failure
	defer func() {	// Defer function runs when outer function finishes
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("failed to rollback transaction: %v", rbErr)
			}
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			log.Printf("failed to commit transaction: %v", commitErr)
		}
	}()

	_, err = tx.ExecContext(ctx,
		"INSERT INTO warranty (item_id, purchase_date, expiration, cust_name, cust_email, cust_phone) VALUES ($1, $2, $3, $4, $5, $6)",
			warranty.ItemID, warranty.PurchaseDate, warranty.Expiration, warranty.CustName, warranty.CustEmail, warranty.CustPhone,
		)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE items SET warranty = $1 WHERE id = $2", "active", warranty.ItemID)
	if err != nil {
		return err
	} 

	return nil
}

func (s *Store) GetWarrantyByItemId(item_id int) (*types.Warranty, error) {
	var warranty types.Warranty

	err := s.db.QueryRow("SELECT id, item_id, purchase_date, expiration, cust_name, cust_email, cust_phone FROM warranty WHERE item_id = $1", item_id,
			).Scan(&warranty.ID, &warranty.ItemID, &warranty.PurchaseDate, &warranty.Expiration, &warranty.CustName, &warranty.CustEmail, &warranty.CustPhone)
	if err != nil {
		return nil, err
	}

	return &warranty, nil
}

func (s *Store) GetAllWarranty(limit int, offset int, search string) ([]types.Warranty, int, error) {
	var (
		 rows *sql.Rows
		 err error
	)

	warrantyCount, err := s.GetWarrantyCount(search)
	if err != nil {
		return nil, 0, err
	}

	if search != "" {

		searchPattern := "%" + search + "%"

		rows, err = s.db.QueryContext(context.Background(), 
			`SELECT w.id, w.item_id, w.purchase_date, w.expiration, w.cust_name, w.cust_email, w.cust_phone, i.serial_number 
				FROM warranty w 
				JOIN items i ON w.item_id = i.id 
				WHERE w.cust_name ILIKE $1 
				OR w.cust_email ILIKE $1 
				OR w.purchase_date::text ILIKE $1 
				OR w.expiration::text ILIKE $1 
				OR i.serial_number ILIKE $1 
				ORDER BY w.id ASC 
			 LIMIT $2 OFFSET $3`, searchPattern, limit, offset,
		)
		if err != nil {
			return nil, 0, err
		}
	} else {
		rows, err = s.db.QueryContext(context.Background(), 
		   "SELECT w.id, w.item_id, w.purchase_date, w.expiration, w.cust_name, w.cust_email, w.cust_phone, i.serial_number FROM warranty w JOIN items i ON w.item_id = i.id  ORDER BY id ASC LIMIT $1 OFFSET $2", limit, offset)
		if err != nil {
			return nil, 0, err
		}
	}

	defer rows.Close()
	var warranties []types.Warranty

	for rows.Next() {
		var warranty types.Warranty

		if err := rows.Scan(&warranty.ID, &warranty.ItemID, &warranty.PurchaseDate, &warranty.Expiration, &warranty.CustName, &warranty.CustEmail, &warranty.CustPhone, &warranty.ItemSN); err != nil {
			return nil, 0, err
		}

		warranties = append(warranties, warranty)
	}

	if err = rows.Err(); err != nil {
        return nil, 0, err
    }

	return warranties, warrantyCount, nil
}

func (s *Store) GetWarrantyCount(search string) (int, error) {
	warrantyCount := 0
	if (search == "") {
		err := s.db.QueryRowContext(context.Background(), 
		"SELECT COUNT(*) FROM warranty", 
	).Scan(&warrantyCount)
		if err != nil {
			return 0, err
		}
	} else {
		searchPattern := "%" + search + "%"

		err := s.db.QueryRowContext(context.Background(), 
		`SELECT COUNT(*) FROM warranty w 
				JOIN items i ON w.item_id = i.id 
				WHERE w.cust_name ILIKE $1 
				OR w.cust_email ILIKE $1 
				OR w.purchase_date::text ILIKE $1 
				OR w.expiration::text ILIKE $1 
				OR i.serial_number ILIKE $1`, searchPattern, 
	).Scan(&warrantyCount)
		if err != nil {
			return 0, err
		}
	}

	return warrantyCount, nil
}

func (s *Store) NewItemSold(sold_item types.SoldItem, ctx context.Context) error {
	tx, err := s.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	defer func() {	
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("failed to rollback transaction: %v", rbErr)
				return
			}
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			log.Printf("failed to commit transaction: %v", commitErr)
			return
		}
	}()

	_, err = tx.ExecContext(ctx,
		"INSERT INTO sold_items (item_id, invoice, ol_shop) VALUES ($1, $2, $3)",
			sold_item.ItemID, sold_item.Invoice, sold_item.OnlineShop,
		)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE items SET status = $1 WHERE id = $2", "sold-pending", sold_item.ItemID)
	if err != nil {
		return err
	} 

	return nil
}

func (s *Store) UpdateItemSold(updated_solditem types.SoldItem) error {
	_, err := s.db.Exec("UPDATE sold_items SET payment_status = $1, journal = $2 WHERE id = $3", updated_solditem.PaymentStatus, updated_solditem.Journal, updated_solditem.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAllSoldItems(limit int, offset int, search string) ([]types.SoldItem, int, error) {
	var (
		rows *sql.Rows
		err error
   )

   soldItemsCount, err := s.GetSoldItemsCount(search)
   if err != nil {
	return nil, 0, err
   }

   if search != "" {
	   searchPattern := "%" + search + "%"

	   rows, err = s.db.QueryContext(context.Background(), 
		   `SELECT s.id, s.item_id, s.datetime_sold, s.invoice, s.ol_shop, s.payment_status, s.journal, i.serial_number 
			   FROM sold_items s 
			   JOIN items i ON s.item_id = i.id 
			   WHERE s.datetime_sold::text ILIKE $1 
			   OR s.invoice ILIKE $1 
			   OR s.ol_shop ILIKE $1
			   OR i.serial_number ILIKE $1 
			   ORDER BY s.id DESC 
			LIMIT $2 OFFSET $3`, searchPattern, limit, offset,
		)
	   if err != nil {
		   return nil, 0, err
	   }
   } else {
	   rows, err = s.db.QueryContext(context.Background(), 
	      `SELECT s.id, s.item_id, s.datetime_sold, s.invoice, s.ol_shop, s.payment_status, s.journal, i.serial_number 
		  FROM sold_items s JOIN items i ON s.item_id = i.id 
		  ORDER BY id DESC LIMIT $1 OFFSET $2`, limit, offset)
	   if err != nil {
		   return nil, 0, err
	   }
   }

	defer rows.Close()
	var soldItems []types.SoldItem

	for rows.Next() {
		var soldItem types.SoldItem

		if err := rows.Scan(&soldItem.ID, &soldItem.ItemID, &soldItem.DatetimeSold, &soldItem.Invoice, &soldItem.OnlineShop, &soldItem.PaymentStatus, &soldItem.Journal, &soldItem.ItemSN); err != nil {
			return nil, 0, err
		}

		soldItems = append(soldItems, soldItem)
	}

	if err = rows.Err(); err != nil {
        return nil, 0, err
    }

	return soldItems, soldItemsCount, nil
}

func (s *Store) GetSoldItemsCount (search string) (int, error) {
	soldItemsCount := 0

	if (search == "") {
		err := s.db.QueryRowContext(context.Background(), 
		"SELECT COUNT(*) FROM sold_items", 
	).Scan(&soldItemsCount)
		if err != nil {
			return 0, err
		}
	} else {
		searchPattern := "%" + search + "%"

		err := s.db.QueryRowContext(context.Background(), 
		`SELECT COUNT(*) FROM sold_items s 
			   JOIN items i ON s.item_id = i.id 
			   WHERE s.datetime_sold::text ILIKE $1 
			   OR s.invoice ILIKE $1 
			   OR i.serial_number ILIKE $1`, searchPattern, 
	).Scan(&soldItemsCount)
		if err != nil {
			return 0, err
		}
	}

	return soldItemsCount, nil
}
