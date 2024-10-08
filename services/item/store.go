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
	_, err := s.db.Exec("INSERT INTO items (serial_number, rfid_tag, item_name, price) VALUES ($1, $2, $3, $4)", item.SerialNumber, item.RFIDTag, item.ItemName, item.Price)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetItemByRFIDTag(rfid_tag string) (*types.Item, error) {
	var item types.Item

	err := s.db.QueryRow("SELECT id, serial_number, rfid_tag, item_name, price FROM items WHERE rfid_tag = $1", rfid_tag).Scan(
		&item.ID, &item.SerialNumber, &item.RFIDTag, &item.ItemName, &item.Price,
	)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *Store) GetItemById(item_id int) (*types.Item, error) {
	var item types.Item

	err := s.db.QueryRow("SELECT id, serial_number, rfid_tag, item_name, price FROM items WHERE id = $1", item_id).Scan(
		&item.ID, &item.SerialNumber, &item.RFIDTag, &item.ItemName, &item.Price,
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
			"SELECT id, serial_number, rfid_tag, item_name, price, warranty, sold FROM items WHERE serial_number ILIKE $1 OR rfid_tag ILIKE $1 OR item_name ILIKE $1 ORDER BY id ASC LIMIT $2 OFFSET $3", searchPattern, limit, offset,
		)
		if err != nil {
			return nil, 0, err
		}
	} else {
		rows, err = s.db.QueryContext(context.Background(), "SELECT id, serial_number, rfid_tag, item_name, price, warranty, sold FROM items ORDER BY id ASC LIMIT $1 OFFSET $2", limit, offset)
		if err != nil {
			return nil, 0, err
		}
	}

	defer rows.Close()	// Close rows after finish processing the data

	var items []types.Item

	for rows.Next() {
		var item types.Item

		if err := rows.Scan(&item.ID, &item.SerialNumber, &item.RFIDTag, &item.ItemName, &item.Price, &item.Warranty, &item.Sold); err != nil {
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

func (s *Store) DeleteItemBySN(serial_num int) error {
	_, err := s.db.Exec("DELETE FROM items WHERE serial_number = $1", serial_num)
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

func (s *Store) GetAllWarranty(limit int, offset int, search string) ([]types.Warranty, error) {

	var (
		 rows *sql.Rows
		 err error
	)

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
			return nil, err
		}
	} else {
		rows, err = s.db.QueryContext(context.Background(), 
		   "SELECT w.id, w.item_id, w.purchase_date, w.expiration, w.cust_name, w.cust_email, w.cust_phone, i.serial_number FROM warranty w JOIN items i ON w.item_id = i.id  ORDER BY id ASC LIMIT $1 OFFSET $2", limit, offset)
		if err != nil {
			return nil, err
		}
	}

	defer rows.Close()
	var warranties []types.Warranty

	for rows.Next() {
		var warranty types.Warranty

		if err := rows.Scan(&warranty.ID, &warranty.ItemID, &warranty.PurchaseDate, &warranty.Expiration, &warranty.CustName, &warranty.CustEmail, &warranty.CustPhone, &warranty.ItemSN); err != nil {
			return nil, err
		}

		warranties = append(warranties, warranty)
	}

	if err = rows.Err(); err != nil {
        return nil, err
    }

	return warranties, nil
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
			}
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			log.Printf("failed to commit transaction: %v", commitErr)
		}
	}()

	_, err = tx.ExecContext(ctx,
		"INSERT INTO sold_items (item_id, invoice, payment_method, payment_status) VALUES ($1, $2, $3, $4)",
			sold_item.ItemID, sold_item.Invoice, sold_item.PaymentMethod, sold_item.PaymentStatus,
		)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE items SET sold = $1 WHERE id = $2", true, sold_item.ItemID)
	if err != nil {
		return err
	} 

	return nil
}

func (s *Store) GetAllSoldItems(limit int, offset int, search string) ([]types.SoldItem, error) {
	var (
		rows *sql.Rows
		err error
   )

   if search != "" {

	   searchPattern := "%" + search + "%"

	   rows, err = s.db.QueryContext(context.Background(), 
		   `SELECT s.id, s.item_id, s.datetime_sold, s.invoice, s.payment_method, s.payment_status, i.serial_number 
			   FROM sold_items s 
			   JOIN items i ON s.item_id = i.id 
			   WHERE s.datetime_sold::text ILIKE $1 
			   OR s.invoice ILIKE $1 
			   OR s.payment_status ILIKE $1 
			   OR i.serial_number ILIKE $1 
			   ORDER BY s.id ASC 
			LIMIT $2 OFFSET $3`, searchPattern, limit, offset,
	   )
	   if err != nil {
		   return nil, err
	   }
   } else {
	   rows, err = s.db.QueryContext(context.Background(), 
	      `SELECT s.id, s.item_id, s.datetime_sold, s.invoice, s.payment_method, s.payment_status, i.serial_number 
		  FROM sold_items s JOIN items i ON s.item_id = i.id 
		  ORDER BY id ASC LIMIT $1 OFFSET $2`, limit, offset)
	   if err != nil {
		   return nil, err
	   }
   }

	defer rows.Close()
	var soldItems []types.SoldItem

	for rows.Next() {
		var soldItem types.SoldItem

		if err := rows.Scan(&soldItem.ID, &soldItem.ItemID, &soldItem.DatetimeSold, &soldItem.Invoice, &soldItem.PaymentMethod, &soldItem.PaymentStatus, &soldItem.ItemSN); err != nil {
			return nil, err
		}

		soldItems = append(soldItems, soldItem)
	}

	if err = rows.Err(); err != nil {
        return nil, err
    }

	return soldItems, nil
}
