package item

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

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
	router.HandleFunc("/register-item", auth.WithJWTAuth(h.handleRegisterItem, h.userStore)).Methods("POST")
	router.HandleFunc("/delete/{rfid_tag}", h.handleDeleteItem).Methods("DELETE")
	router.HandleFunc("/register-warranty/{rfid_tag}", h.handleActivateNewWarranty).Methods("POST")
	router.HandleFunc("/item-sold/{rfid_tag}", h.handleItemSold).Methods("POST")
	router.HandleFunc("/get-items", h.handleGetItems).Methods("GET")
	router.HandleFunc("/get-types", h.handleGetItemTypes).Methods("GET")
	router.HandleFunc("/get-warranties", h.handleGetAllWarranties).Methods("GET")
	router.HandleFunc("/get-sold-items", h.handleGetAllSoldItem).Methods("GET")
	router.HandleFunc("/item-sold-bulk", h.handleItemSoldBulk).Methods("POST")
	router.HandleFunc("/ship-items", h.handleShipItems).Methods("PATCH")
	router.HandleFunc("/edit-item-sold", h.handleUpdateSoldItem).Methods("PATCH")
	router.HandleFunc("/get-item-rfid/{rfid_tag}", h.handleGetItemByRFID).Methods("GET")
	router.HandleFunc("/register-item-type", h.handleCreateItemType).Methods("POST")
	router.HandleFunc("/get-avail-item", h.handleGetAvailItemBySN).Methods("GET")
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
		Quantity: payload.Quantity,
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

	// Delete item
	err = h.store.DeleteItemByRFID(rfid_tag)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error deleting item: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, "item deleted")
}

func(h *Handler) handleGetItemByRFID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rfid_tag := vars["rfid_tag"]

	i, err := h.store.GetItemByRFIDTag(rfid_tag)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error getting item by rfid: %v", err))
	}

	utils.WriteJSON(w, http.StatusOK, i)
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
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("JSON parsing error: %v", err))
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
	}, i.Quantity, ctx)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error registering sold item: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "Sold item registered")
}

func (h *Handler) handleItemSoldBulk (w http.ResponseWriter, r *http.Request) {
	// Get JSON payload
	var payload types.SoldItemBulkPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("parsing error: %v", err))
		return
	}

	ctx := r.Context()

	// Validate JSON
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// Get and register items
	for _, itemTag := range payload.ItemTags {
		i, err := h.store.GetItemByRFIDTag(itemTag)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error getting items: %v", err))
			return
		}

		err = h.store.NewItemSold(types.SoldItem{
			ItemID: i.ID,
			Invoice: payload.Invoice,
			OnlineShop: payload.OnlineShop,
		}, i.Quantity, ctx)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error bulk registering sold items: %v", err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, "Sold items registered in bulk")
}

func (h *Handler) handleShipItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Get JSON payload
	var payload types.ShipItemsPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("parsing error: %v", err))
		return
	}

	// Validate JSON
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error validating payload: %v", errors))
		return
	}

	// Get and register items
	for _, itemTag := range payload.ItemTags {
		i, err := h.store.GetItemByRFIDTag(itemTag)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error getting items: %v", err))
			return
		}

		err = h.store.ShipItem(i.ID, ctx)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error shipping items: %v", err))
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, "Item Shipped")
}

func (h *Handler) handleUpdateSoldItem (w http.ResponseWriter, r *http.Request) {
	// Get JSON payload
	var payload types.SoldItem
	err := utils.ParseJSON(r, &payload) 
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error parsing JSON: %v", err))
		return
	}

	// Validate JSON
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error validating payload: %v", errors))
		return
	}

	// Update item sold record by id
	err = h.store.UpdateItemSold(types.SoldItem{
		PaymentStatus: payload.PaymentStatus,
		Journal: payload.Journal,
		ID: payload.ID,
	})
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("error updating item: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, "Sold item updated")
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
