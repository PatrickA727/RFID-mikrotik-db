package item

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/PatrickA727/mikrotik-db-sys/services/auth"
	"github.com/PatrickA727/mikrotik-db-sys/types"
	"github.com/PatrickA727/mikrotik-db-sys/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	store types.ItemStore
	userStore types.UserStore
}

func NewHandler (store types.ItemStore, userStore types.UserStore) *Handler {
	return &Handler{
		store: store,
		userStore: userStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/register-item", auth.MobileAuth(h.handleRegisterItem, h.userStore)).Methods("POST")	// Mobile App
	router.HandleFunc("/delete/{rfid_tag}", auth.WithJWTAuth(h.handleDeleteItem, h.userStore)).Methods("DELETE")
	// router.HandleFunc("/item-sold/{rfid_tag}", auth.WithJWTAuth(h.handleItemSold, h.userStore)).Methods("POST")	// Unused
	router.HandleFunc("/get-items", auth.WithJWTAuth(h.handleGetItems, h.userStore)).Methods("GET")
	router.HandleFunc("/get-types", auth.MobileAuth(h.handleGetItemTypes, h.userStore)).Methods("GET")	// Mobile App
	router.HandleFunc("/get-sold-items", auth.WithJWTAuth(h.handleGetAllSoldItem, h.userStore)).Methods("GET")
	router.HandleFunc("/item-sold-bulk", auth.WithJWTAuth(h.handleItemSoldBulk, h.userStore)).Methods("POST")
	router.HandleFunc("/ship-items/{invoice_id}", auth.MobileAuth(h.handleShipItems, h.userStore)).Methods("PATCH")	// Mobile App
	// router.HandleFunc("/edit-item-sold", auth.WithJWTAuth(h.handleUpdateSoldItem, h.userStore)).Methods("PATCH")
	router.HandleFunc("/get-item-rfid/{rfid_tag}", auth.WithJWTAuth(h.handleGetItemByRFID, h.userStore)).Methods("GET")	// Unused
	router.HandleFunc("/get-sold-by-rfid/{rfid_tag}", auth.MobileAuth(h.handleGetSoldItem, h.userStore)).Methods("GET")	// Mobile App
	router.HandleFunc("/register-item-type", auth.WithJWTAuth(h.handleCreateItemType, h.userStore)).Methods("POST")
	router.HandleFunc("/get-avail-item", auth.WithJWTAuth(h.handleGetAvailItemBySN, h.userStore)).Methods("GET")
	router.HandleFunc("/get-invoice-items/{id}", auth.MobileAuth(h.handleGetItemsByInvoice, h.userStore)).Methods("GET") // Mobile App
	router.HandleFunc("/get-invoices", auth.MobileAuth(h.handleGetInvoices, h.userStore)).Methods("GET") // Mobile App
	router.HandleFunc("/get-all-invoices", auth.WithJWTAuth(h.handleGetAllInvoice, h.userStore)).Methods("GET")
	router.HandleFunc("/edit-invoice/{id}", auth.WithJWTAuth(h.handleEditInvoice, h.userStore)).Methods("PATCH")
	router.HandleFunc("/delete-invoice/{id}", auth.WithJWTAuth(h.handleDeleteInvoice, h.userStore)).Methods("DELETE")
	router.HandleFunc("/get-status-count", auth.WithJWTAuth(h.handleGetItemStatusCount, h.userStore)).Methods("GET")
	router.HandleFunc("/get-type-count", auth.WithJWTAuth(h.handleGetItemTypeCount, h.userStore)).Methods("GET")
}

func (h *Handler) handleRegisterItem(w http.ResponseWriter, r *http.Request) {
	// Get JSON Payload
	var payload types.RegisterItemPayload
	err := utils.ParseJSON(r, &payload)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error parsing JSON: %v", err))
		return
	}

	// Validate JSON Payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// Create item
	err = h.store.CreateItem(types.Item{
		SerialNumber: payload.SerialNumber,
		RFIDTag: payload.RFIDTag,
		Batch: payload.Batch,
		TypeRef: payload.TypeRef,
	})
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error creating item %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "Item Created")
}

func (h *Handler) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	var (
		err error
	)

	// Get serial number from path parameter
	vars := mux.Vars(r)
	rfid_tag := vars["rfid_tag"]


	_, err = h.store.GetItemByRFIDTag(rfid_tag)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("item doesnt exist"))
		return
	}

	// Delete item
	err = h.store.DeleteItemByRFID(rfid_tag)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error deleting item: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "item deleted")
}

func (h *Handler) handleGetSoldItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rfid_tag := vars["rfid_tag"]

	i, err := h.store.GetSoldItemByRFID(rfid_tag)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error getting sold item rfid: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, i)
}

func(h *Handler) handleGetItemByRFID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rfid_tag := vars["rfid_tag"]

	i, err := h.store.GetItemByRFIDTag(rfid_tag)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error getting item by rfid: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, i)
}

func (h *Handler) handleGetItems(w http.ResponseWriter, r *http.Request) {
	// Get limit, offset and search path parameters
	limitStr := r.URL.Query().Get("limit")
    offsetStr := r.URL.Query().Get("offset")
	searchQuery := r.URL.Query().Get("search")
	statusQuery := r.URL.Query().Get("status")

	// Convert limit and offset to int
    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 {
        limit = 10
    }

	offset, err := strconv.Atoi(offsetStr)
    if err != nil || offset < 0 {
        offset = 0
    }

	// Get items
	items, itemCount, err := h.store.GetItems(limit, offset, searchQuery, statusQuery)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error retrieving all items: %v", err))
		return
	}

	if len(items) == 0 {
		utils.WriteJSON(w, http.StatusOK, types.ItemsResponse{Items: []types.Item{}})
		return
	} else {
		response := types.ItemsResponse{
			Items:     items,
			ItemCount: itemCount,
		}
	
		utils.WriteJSON(w, http.StatusOK, response)
	}

}

func (h *Handler) handleGetAvailItemBySN(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("search")
	if searchQuery == "" {
		return
	}

	items, err := h.store.GetItemByIdSearch(searchQuery)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error getting items: %v", err))
		return
	}

	if len(items) == 0 {
		utils.WriteJSON(w, http.StatusOK, []types.Item{})
		return
	} else {
		utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"items": items,
		})
	}

}

// func (h *Handler) handleItemSold(w http.ResponseWriter, r *http.Request) {
// 	// Get uuid from path parameter
// 	vars := mux.Vars(r)
// 	rfid_tag := vars["rfid_tag"]

// 	ctx := r.Context()

// 	// Get JSON payload
// 	var payload types.SoldItemPayload 
// 	if err := utils.ParseJSON(r, &payload); err != nil {
// 		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("JSON parsing error: %v", err))
// 		return
// 	}

// 	// Validate JSON
// 	if err := utils.Validate.Struct(payload); err != nil {
// 		errors := err.(validator.ValidationErrors)
// 		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
// 		return
// 	}

// 	// Get item
// 	i, err := h.store.GetItemByRFIDTag(rfid_tag)
// 	if err != nil {
// 		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error getting item: %v", err))
// 		return
// 	}

// 	// Register new sold item
// 	err = h.store.NewItemSold(types.SoldItem{
// 		ItemID: i.ID,
// 		Invoice: payload.Invoice,
// 		OnlineShop: payload.OnlineShop,
// 	}, i.Quantity, ctx)
// 	if err != nil {
// 		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error registering sold item: %v", err))
// 		return
// 	}

// 	utils.WriteJSON(w, http.StatusCreated, "Sold item registered")
// }

func (h *Handler) handleItemSoldBulk (w http.ResponseWriter, r *http.Request) {
	var (
		i *types.Item
		err  error
	)

	// Transaction
	ctx := r.Context()
	tx, err := h.store.BeginTransaction(ctx)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error1: %v", err))
		return
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

	// Get JSON payload
	var payload types.SoldItemBulkPayload
	if err = utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("parsing error: %v", err))
		return
	}

	// Validate JSON
	if err = utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	if payload.SerialNums == nil || len(payload.SerialNums) == 0 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("no tags found"))
		return
	}

	// Create Invoice
	invoice_id, err := h.store.CreateInvoice(payload.Invoice, payload.OnlineShop, tx, ctx)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error2: %v", err))
		return
	}

	// Get and register items
	for _, SerialNum := range payload.SerialNums {
		i, err = h.store.GetItemBySN(SerialNum, tx, ctx)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error getting items: %v", err))
			return
		}

		err = h.store.NewItemSold(types.SoldItem{
			ItemID: i.ID,
			InvoiceID: invoice_id,
			OnlineShop: payload.OnlineShop,
		},  tx, ctx)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error bulk registering sold items: %v", err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, "Sold items registered in bulk")
}

func (h *Handler) handleShipItems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoice_id_str := vars["invoice_id"]
	invoice_id, err := strconv.Atoi(invoice_id_str)
    if err != nil{
        utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error: %v", err))
		return
    }

	ctx := r.Context()
	tx, err := h.store.BeginTransaction(ctx)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error1: %v", err))
		return
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

	err = h.store.ShipInvoice(invoice_id, tx, ctx)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error1: %v", err))
		return
	}

	soldItems, err := h.store.GetItemsByInvoice(invoice_id)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error2: %v", err))
		return
	}

	// Get and register items
	for _, itemRFIDTag := range soldItems {
		i, err := h.store.GetItemByRFIDTag(itemRFIDTag.ItemTag)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error getting items: %v", err))
			return
		}

		err = h.store.ShipItem(i.ID, tx, ctx)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error shipping items: %v", err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, "Item Shipped")
}

// func (h *Handler) handleUpdateSoldItem (w http.ResponseWriter, r *http.Request) {
// 	// Get JSON payload
// 	var payload types.SoldItem
// 	err := utils.ParseJSON(r, &payload) 
// 	if err != nil {
// 		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error parsing JSON: %v", err))
// 		return
// 	}

// 	// Validate JSON
// 	if err := utils.Validate.Struct(payload); err != nil {
// 		errors := err.(validator.ValidationErrors)
// 		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error validating payload: %v", errors))
// 		return
// 	}

// 	// Update item sold record by id
// 	err = h.store.UpdateItemSold(types.SoldItem{
// 		ID: payload.ID,
// 	})
// 	if err != nil {
// 		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error updating item: %v", err))
// 		return
// 	}

// 	utils.WriteJSON(w, http.StatusCreated, "Sold item updated")
// }

func (h *Handler) handleGetAllSoldItem (w http.ResponseWriter, r *http.Request) {
	// Get limit, offset and search path parameters
	limitStr := r.URL.Query().Get("limit")
    offsetStr := r.URL.Query().Get("offset")
	searchQuery := r.URL.Query().Get("search")

	// Convert limit and offset to int
    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 {
        limit = 10
    }

	offset, err := strconv.Atoi(offsetStr)
    if err != nil || offset < 0 {
        offset = 0
    }
	
	// Get sold items
	soldItems, soldItemsCount, err := h.store.GetAllSoldItems(limit, offset, searchQuery)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error getting sold items: %v", err))
		return
	}

	if len(soldItems) == 0 {
		utils.WriteJSON(w, http.StatusOK, []types.SoldItem{})
	} else {
		response := types.SoldItemsResponse {
			SoldItems: soldItems,
			SoldItemsCount: soldItemsCount,
		}
	
		utils.WriteJSON(w, http.StatusOK, response)
	}

}

func (h *Handler) handleCreateItemType(w http.ResponseWriter, r *http.Request) {
	// Get JSON
	var payload types.ItemTypePayload
	err := utils.ParseJSON(r, &payload)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error parsing: %v", err))
		return
	}

	//Validate JSON
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// Create item
	err = h.store.CreateItemType(types.ItemType{
		TypeName: payload.ItemType,
		Price: payload.Price,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error creating type: %v", err))
	}

	utils.WriteJSON(w, http.StatusCreated, payload);
}

func (h *Handler) handleGetItemTypes (w http.ResponseWriter, r *http.Request) {
	item_types, err := h.store.GetItemTypes()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting types: %v", err));
		return
	}

	response := types.TypesResponse{
		ItemTypes: item_types,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) handleGetItemsByInvoice (w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoice_id_str := vars["id"]
	invoice_id, err := strconv.Atoi(invoice_id_str)
    if err != nil{
        utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error: %v", err))
		return
    }

	// Get invoice by id
	invoice, err := h.store.GetInvoiceByID(invoice_id)
	if err != nil{
        utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error: %v", err))
		return
    }

	// Get items by invoice
	items, err := h.store.GetItemsByInvoice(invoice_id)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error get items: %v", err))
		return
	}

	if len(items) == 0 {
		utils.WriteJSON(w, http.StatusOK, []types.SoldItem{})
	} else {
		response := types.InvoiceItemsResponse {
			SoldItems: items,
			InvoiceStr: invoice.InvoiceStr,
			OnlineShop: invoice.OnlineShop,
		}
	
		utils.WriteJSON(w, http.StatusOK, response)
	}
}

func (h *Handler) handleGetInvoices (w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("search")

	invoices, err := h.store.GetInvoices(searchQuery)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error: %v", err))
		return
	}

	if len(invoices) == 0 {
		utils.WriteJSON(w, http.StatusOK, []types.Invoice{})
	} else {
		utils.WriteJSON(w, http.StatusOK, invoices)
	}
}

func (h *Handler) handleGetAllInvoice (w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	invoice := r.URL.Query().Get("invoice")
	status := r.URL.Query().Get("status")
	limit := 10

	page, err := strconv.Atoi(pageStr)
    if err != nil {
        limit = 10
    }

	offset := (page - 1) * limit

	invoices, count, err := h.store.GetAllInvoice(limit, offset, invoice, status)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error: %v", err))
		return
	}

	if len(invoices) == 0 {
		utils.WriteJSON(w, http.StatusOK, types.InvoicesResponse{Invoices: []types.Invoice{}})
		return
	} else {
		response := types.InvoicesResponse{
			Invoices: invoices,
			Count: count,
		}
	
		utils.WriteJSON(w, http.StatusOK, response)
	}
}

func (h *Handler) handleGetItemStatusCount (w http.ResponseWriter, r *http.Request) {
	not_sold, sold_pending, sold_shipped, err := h.store.GetItemStatusCount()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error: %v", err));
		return
	}

	utils.WriteJSON(w, http.StatusOK, types.ItemStatusCount{
		NotSold: not_sold,
		SoldPending: sold_pending,
		SoldShipped: sold_shipped,
	})
}

func (h *Handler) handleGetItemTypeCount (w http.ResponseWriter, r *http.Request) {
	counts, err := h.store.GetItemTypeCount()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error: %v", err));
		return
	}

	utils.WriteJSON(w, http.StatusOK, counts)
}

func (h *Handler) handleEditInvoice (w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rfid_tag_str := vars["id"]

	rfid_tag, err := strconv.Atoi(rfid_tag_str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error: %v", err))
		return
	}

	var payload types.EditInvoice
	err = utils.ParseJSON(r, &payload) 
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error parsing JSON: %v", err))
		return
	}

	err = h.store.EditInvoice(rfid_tag, payload)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Invoice updated")
}

func (h *Handler) handleDeleteInvoice (w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoice_id_str := vars["id"]

	invoice_id, err := strconv.Atoi(invoice_id_str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error: %v", err))
		return
	}

	items, err := h.store.GetItemsByInvoice(invoice_id)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error1: %v", err))
		return
	}

	if len(items) == 0 {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("no items found for invoice %d", invoice_id))
		return
	}

	ctx := r.Context()
	tx, err := h.store.BeginTransaction(ctx)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error: %v", err))
		return
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

	err = h.store.DeleteInvoice(invoice_id, tx, ctx)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error2: %v", err))
		return
	}

	err = h.store.ResetItemsToNotSold(items, tx, ctx)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error3: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "Invoice deleted")
}
