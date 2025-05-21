package item

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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
	_, err := s.db.Exec("INSERT INTO items (serial_number, rfid_tag, batch, type_ref) VALUES ($1, $2, $3, $4)", 
						item.SerialNumber, item.RFIDTag, item.Batch, item.TypeRef,
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

	err := s.db.QueryRow("SELECT id, serial_number, rfid_tag, batch, type_ref FROM items WHERE rfid_tag = $1", rfid_tag).Scan(
		&item.ID, &item.SerialNumber, &item.RFIDTag, &item.Batch, &item.TypeRef,
	)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *Store) GetItemBySN(serial_num string, tx *sql.Tx, ctx context.Context) (*types.Item, error) {
	var item types.Item

	err := tx.QueryRowContext(ctx, "SELECT id, serial_number, rfid_tag, batch, type_ref FROM items WHERE serial_number = $1", serial_num).Scan(
		&item.ID, &item.SerialNumber, &item.RFIDTag, &item.Batch, &item.TypeRef,
	)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *Store) GetSoldItemByRFID(rfid_tag string) (*types.Item, error) {
	var item types.Item

	err := s.db.QueryRow("SELECT id, serial_number, rfid_tag, type_ref FROM items WHERE rfid_tag = $1 AND status = $2", rfid_tag, "sold-pending").Scan(
		&item.ID, &item.SerialNumber, &item.RFIDTag, &item.TypeRef,
	)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *Store) GetItemByIdSearch(search string) ([]types.ItemSellingResponse, error) {
	sanitizedInput := strings.ReplaceAll(search, "%", "\\%")
    sanitizedInput = strings.ReplaceAll(sanitizedInput, "_", "\\_")
	searchPattern := sanitizedInput + "%"

	rows, err := s.db.QueryContext(context.Background(), "SELECT id, serial_number, rfid_tag, type_ref FROM items where serial_number ILIKE $1 AND status ILIKE $2 ORDER BY batch DESC LIMIT 10", searchPattern, "not sold")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var items []types.ItemSellingResponse

	for rows.Next() {
		var item types.ItemSellingResponse

		if err := rows.Scan(&item.ID, &item.SerialNumber, &item.RFIDTag, &item.TypeRef); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}


func (s *Store) GetItems(limit int, offset int, search string, status string) ([]types.Item ,int, error) {
	var (
		 rows *sql.Rows
		 err error
	)

	var args []interface{}
   	var conditions []string

	query := `SELECT id, serial_number, rfid_tag, batch, status, type_ref, 
			  createdat FROM items`

	if search != "" {
		args = append(args, search+"%")

		conditions = append(conditions, fmt.Sprintf("(serial_number ILIKE $%d OR rfid_tag ILIKE $%d)", len(args), len(args)))
	}

	if status != "" {
		args = append(args, status)

		conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	args = append(args, limit, offset)
	query += fmt.Sprintf(" ORDER BY batch DESC, id DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err = s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}

	itemCount, err := s.GetItemCount(search, status)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()	// Close rows database connection after finish processing the data/function

	var items []types.Item

	for rows.Next() {
		var item types.Item

		if err := rows.Scan(&item.ID, &item.SerialNumber, &item.RFIDTag, &item.Batch, &item.Status, &item.TypeRef, &item.CreatedAt); err != nil {
			return nil, 0, err
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
        return nil, 0, err
    }

    return items, itemCount, nil
}

func (s *Store) GetItemCount(search string, status string) (int, error) {
	itemCount := 0

	var args []interface{}
   	var conditions []string

	query := `SELECT COUNT(*) FROM items`

	if search != "" {
		args = append(args, search+"%")

		conditions = append(conditions, fmt.Sprintf("(serial_number ILIKE $%d OR rfid_tag ILIKE $%d)", len(args), len(args)))
	}

	if status != "" {
		args = append(args, status+"%")

		conditions = append(conditions, fmt.Sprintf("status ILIKE $%d", len(args)))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	err := s.db.QueryRow(query, args...).Scan(&itemCount)
	if err != nil {
		return 0, err
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

func (s *Store) CreateInvoice(invoice string, ol_shop string, tx *sql.Tx, ctx context.Context) (int, error) {
	invoice_id := 0

	_, err := tx.ExecContext(ctx, 
	`INSERT INTO invoice (invoice_str, online_shop) VALUES ($1, $2)`, invoice, ol_shop)
	if err != nil {
		return 0, err
	}

	err = tx.QueryRowContext(ctx, `SELECT id FROM invoice WHERE invoice_str = $1`, invoice).Scan(&invoice_id)
	if err != nil {
		return 0, err
	}

	return invoice_id, nil
}

func (s *Store) NewItemSold(sold_item types.SoldItem, tx *sql.Tx, ctx context.Context) error {
	var (
		err error
	)

	_, err = tx.ExecContext(ctx,
		"INSERT INTO sold_items (item_id, ol_shop, invoice_id) VALUES ($1, $2, $3)",
			sold_item.ItemID, sold_item.OnlineShop, sold_item.InvoiceID,
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

// func (s *Store) UpdateItemSold(updated_solditem types.SoldItem) error {
// 	_, err := s.db.Exec("UPDATE sold_items SET payment_status = $1, journal = $2 WHERE id = $3", updated_solditem.ID)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

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
		   `SELECT s.id, s.item_id, s.datetime_sold, s.ol_shop, i.serial_number, i.status
			   FROM sold_items s 
			   JOIN items i ON s.item_id = i.id 
			   WHERE s.datetime_sold::text ILIKE $1 
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
	      `SELECT s.id, s.item_id, s.datetime_sold, s.ol_shop, i.serial_number, i.status 
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

		if err := rows.Scan(&soldItem.ID, &soldItem.ItemID, &soldItem.DatetimeSold, &soldItem.OnlineShop, &soldItem.ItemSN, &soldItem.Status); err != nil {
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
			   OR i.serial_number ILIKE $1`, searchPattern, 
	).Scan(&soldItemsCount)
		if err != nil {
			return 0, err
		}
	}

	return soldItemsCount, nil
}

func (s *Store) ShipItem(item_id int, tx *sql.Tx, ctx context.Context) error {
	_, err := tx.ExecContext(ctx, 
		"UPDATE items SET status = $1 WHERE id = $2 AND status = $3", "sold-shipped", item_id, "sold-pending",
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetItemsByInvoice (invoice_id int) ([]types.SoldItem, error) {
	var items []types.SoldItem

	rows, err := s.db.Query(`SELECT i.id, i.rfid_tag, i.serial_number, i.type_ref 
							FROM sold_items s JOIN items i ON s.item_id = i.id
							WHERE s.invoice_id = $1`, invoice_id);
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var item types.SoldItem

		if err = rows.Scan(&item.ID, &item.ItemTag, &item.ItemSN, &item.ItemType); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
        return nil, err
    }

	return items, nil
}

func (s *Store) GetInvoices (invoice string) ([]types.Invoice, error) {
	searchPattern := invoice + "%"

	rows, err := s.db.Query(`SELECT id, invoice_str FROM invoice
							 WHERE invoice_str ILIKE $1
							 ORDER BY id DESC LIMIT 10`, searchPattern)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var invoices []types.Invoice

	for rows.Next() {
		var invoice types.Invoice

		if err := rows.Scan(&invoice.ID, &invoice.InvoiceStr); err != nil {
			return nil, err
		}

		invoices = append(invoices, invoice)
	}

	if err = rows.Err(); err != nil {
        return nil, err
    }

	return invoices, nil
}

func (s *Store) GetInvoiceByID(id int) (*types.Invoice, error) {
	var invoice types.Invoice

	err := s.db.QueryRow("SELECT invoice_str, status, online_shop FROM invoice WHERE id = $1", id).Scan(
		&invoice.InvoiceStr, &invoice.Status, &invoice.OnlineShop,
	)

	if err != nil {
		return nil, err
	}

	return &invoice, nil
}

func (s *Store) ShipInvoice (invoice_id int, tx *sql.Tx, ctx context.Context) error {
	_, err := tx.ExecContext(ctx, `UPDATE invoice SET status = $1 WHERE id = $2`, "shipped", invoice_id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAllInvoice (limit int, offset int, invoice string, status string) ([]types.Invoice, int, error) {
	var (
		rows *sql.Rows
		err error
   )

   var invoices []types.Invoice
   var args []interface{}
   var conditions []string

   query := "SELECT id, invoice_str, status, online_shop FROM invoice"

	if invoice != "" {
		args = append(args, invoice+"%")

		conditions = append(conditions, fmt.Sprintf("invoice_str ILIKE $%d", len(args)))
	}

	if status != "" {
		args = append(args, status)

		conditions = append(conditions, fmt.Sprintf("status ILIKE $%d", len(args)))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	args = append(args, limit, offset)
	query += fmt.Sprintf(" ORDER BY id DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err = s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var invoice types.Invoice

		if err := rows.Scan(&invoice.ID, &invoice.InvoiceStr, &invoice.Status, &invoice.OnlineShop); err != nil {
			return nil, 0, err
		}

		invoices = append(invoices, invoice)
	}

	if err = rows.Err(); err != nil {
        return nil, 0, err
    }

	count, err := s.GetInvoiceCount(invoice, status)
	if err != nil {
		return nil, 0, err
	}

    return invoices, count, nil
}

func (s *Store) GetInvoiceCount (invoice string, status string) (int, error) {
	invoiceCount := 0

	var args []interface{}
   	var conditions []string

	query := "SELECT COUNT(*) FROM invoice"

	if invoice != "" {
		args = append(args, invoice+"%")

		conditions = append(conditions, fmt.Sprintf("invoice_str ILIKE $%d", len(args)))
	}

	if status != "" {
		args = append(args, status)

		conditions = append(conditions, fmt.Sprintf("status ILIKE $%d", len(args)))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	err := s.db.QueryRow(query, args...).Scan(&invoiceCount)
	if err != nil {
		return 0, err
	}

	return invoiceCount, nil
}

func (s *Store) EditInvoice(id int, payload types.EditInvoice) error {
	var setClauses []string
	var args []interface{}

	argIndex := 1

	if payload.Invoice != "" {
		setClauses = append(setClauses, fmt.Sprintf("invoice_str = $%d", argIndex))
		args = append(args, payload.Invoice)
		argIndex++
	}

	if payload.OnlineShop != "" {
		setClauses = append(setClauses, fmt.Sprintf("online_shop = $%d", argIndex))
		args = append(args, payload.OnlineShop)
		argIndex++
	}

	// No fields to update
	if len(setClauses) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Build final query
	query := fmt.Sprintf("UPDATE invoice SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "), argIndex)

	args = append(args, id)

	// Execute query
	_, err := s.db.Exec(query, args...)
	return err
}

func (s *Store) DeleteInvoice(id int, tx *sql.Tx, ctx context.Context) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM invoice WHERE id = $1`, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ResetItemsToNotSold(items []types.SoldItem, tx *sql.Tx, ctx context.Context) error {
	for _, item := range items {
		_, err := tx.ExecContext(ctx, `UPDATE items SET status = 'not sold' WHERE id = $1`, item.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) GetItemStatusCount() (int, int, int, error) {
	var notSoldCount, soldPendingCount, soldShippedCount int

	err := s.db.QueryRow(`
		SELECT
			COUNT(CASE WHEN status = 'not sold' THEN 1 END),
			COUNT(CASE WHEN status = 'sold-pending' THEN 1 END),
			COUNT(CASE WHEN status = 'sold-shipped' THEN 1 END)
		FROM items
	`).Scan(&notSoldCount, &soldPendingCount, &soldShippedCount)
	if err != nil {
		return 0, 0, 0, err
	}

	return notSoldCount, soldPendingCount, soldShippedCount, nil
}

func (s *Store) GetItemTypeCount() (map[string]int, error) {
	counts := make(map[string]int)

	rows, err := s.db.Query("SELECT type_ref, COUNT(*) FROM items GROUP BY type_ref")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var i_type string
		var t_count int

		if err := rows.Scan(&i_type, &t_count); err != nil {
			return nil, err
		}
		counts[i_type] = t_count
	}
	if err := rows.Err(); err != nil {
        return nil, err
    }

	return counts, nil
}