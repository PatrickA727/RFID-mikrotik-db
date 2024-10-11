package item

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/PatrickA727/mikrotik-db-sys/types"
	"github.com/PatrickA727/mikrotik-db-sys/utils"
	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
)

type Handler struct {
	store types.ItemStore
}

func NewHandler (store types.ItemStore) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/register-item", h.handleRegisterItem).Methods("POST")
	router.HandleFunc("/delete/{serial_num}", h.handleDeleteItem).Methods("DELETE")
	router.HandleFunc("/register-warranty/{rfid_tag}", h.handleActivateNewWarranty).Methods("POST")
	router.HandleFunc("/item-sold/{rfid_tag}", h.handleItemSold).Methods("POST")
	router.HandleFunc("/get-items", h.handleGetItems).Methods("GET")
	router.HandleFunc("/get-warranties", h.handleGetAllWarranties).Methods("GET")
	router.HandleFunc("/get-sold-items", h.handleGetAllSoldItem).Methods("GET")
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
		ItemName: payload.ItemName,
	})
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error creating item %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "Item Created")
}

func (h *Handler) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	// Get serial number from path parameter
	vars := mux.Vars(r)
	serial_num := vars["serial_num"]

	serialNum, err := strconv.Atoi(serial_num)
    if err != nil {
        // Handle the error, maybe return a 400 Bad Request
        utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid serial number"))
        return
    }

	// Delete item
	err = h.store.DeleteItemBySN(serialNum)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error deleting item: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "item deleted")
}

func (h *Handler) handleGetItems(w http.ResponseWriter, r *http.Request) {
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

	// Get items
	items, itemCount, err := h.store.GetItems(limit, offset, searchQuery)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error retrieving all items: %v", err))
		return
	}

	if len(items) == 0 {
		utils.WriteJSON(w, http.StatusOK, []types.Item{})
		return
	}

	response := types.ItemsResponse{
		Items:     items,
		ItemCount: itemCount,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) handleActivateNewWarranty(w http.ResponseWriter, r *http.Request) {
	// Get RFID tags from path params
	vars := mux.Vars(r)
	rfid_tag := vars["rfid_tag"]

	// Context for transaction in 'CreateWarranty'
	ctx := r.Context()

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Store payload to struct
	var payload = types.NewWarrantyPayload{
		PurchaseDate: r.FormValue("purchase_date"),
		CustName:     r.FormValue("cust_name"),
		CustEmail:    r.FormValue("cust_email"),
		CustPhone:    r.FormValue("cust_phone"),
	}

	// Validating payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// Parse form date from string to datetime 
	purchaseDate, err := time.Parse("2006-01-02", r.FormValue("purchase_date"))
	if err != nil {
		http.Error(w, "Invalid purchase date format", http.StatusBadRequest)
		return
	}

	// Generate expiration date
	expirationDate := purchaseDate.AddDate(1, 0, 0)

	// Get item ID
	i, err := h.store.GetItemByRFIDTag(rfid_tag)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error getting item: %v", err))
		return
	}

	err = h.store.CreateWarranty(
		types.Warranty{
			ItemID: i.ID,
			PurchaseDate: purchaseDate,
			Expiration: expirationDate,
			CustName: payload.CustName,
			CustEmail: payload.CustEmail,
			CustPhone: payload.CustPhone,
		}, ctx,
	)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error creating warranty: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "warranty created")
}

func (h *Handler) handleGetAllWarranties(w http.ResponseWriter, r *http.Request) {
	// Get limit, offset and search query parameters
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

	// Get warranties
	warranties, warrantyCount, err := h.store.GetAllWarranty(limit, offset, searchQuery)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error getting warranties: %v", err))
		return
	}

	if len(warranties) == 0 {
		utils.WriteJSON(w, http.StatusOK, []types.Warranty{})
		return
	}

	response := types.WarrantyResponse {
		Warranties: warranties,
		WarrantyCount: warrantyCount,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) handleItemSold(w http.ResponseWriter, r *http.Request) {
	// Get uuid from path parameter
	vars := mux.Vars(r)
	rfid_tag := vars["rfid_tag"]

	ctx := r.Context()

	// Get JSON payload
	var payload types.SoldItemPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("JSON parsing error"))
		return
	}

	// Validate JSON
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// Get item
	i, err := h.store.GetItemByRFIDTag(rfid_tag)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error getting item: %v", err))
		return
	}

	// Register new sold item
	err = h.store.NewItemSold(types.SoldItem{
		ItemID: i.ID,
		Invoice: payload.Invoice,
		OnlineShop: payload.OnlineShop,
		PaymentStatus: payload.PaymentStatus,
	}, ctx)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error registering sold item: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "Sold item registered")
}

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
	}

	response := types.SoldItemsResponse {
		SoldItems: soldItems,
		SoldItemsCount: soldItemsCount,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}
