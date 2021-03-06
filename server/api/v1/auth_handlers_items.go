package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

//getters
func ItemsHandler(w http.ResponseWriter, r *http.Request) {

	if items, err := GetItems(); err != nil {
		http.Error(w, err.Error(), 500)
	} else {
		json.NewEncoder(w).Encode(items)
	}
}

func WishlistHandler(w http.ResponseWriter, r *http.Request) {

	if items, err := GetItemsWithStatus(3); err != nil {
		http.Error(w, err.Error(), 500)
	} else {
		json.NewEncoder(w).Encode(items)
	}
}

func StockItemsHandler(w http.ResponseWriter, r *http.Request) {

	if items, err := GetItemsWithStatus(1); err != nil {
		http.Error(w, err.Error(), 500)
	} else {
		json.NewEncoder(w).Encode(items)
	}
}

func PendingItemsHandler(w http.ResponseWriter, r *http.Request) {

	if items, err := GetItemsWithStatus(2); err != nil {
		http.Error(w, err.Error(), 500)
	} else {
		json.NewEncoder(w).Encode(items)
	}
}

//specific getters
func ItemHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	var itemID int
	var categoryID int

	var err1 error
	var err2 error

	itemID, err1 = strconv.Atoi(vars["item_id"])
	categoryID, err2 = strconv.Atoi(vars["category_id"])

	if err1 != nil && err2 != nil {
		http.Error(w, err1.Error(), 500)
		return
	}

	var item *Item
	var items *Items
	var err error

	if itemID > 0 {
		item, err = GetItem(itemID)
	} else if categoryID > 0 {
		items, err = GetItemByCategory(categoryID)
	}

	if err != nil {
		http.Error(w, err.Error(), 404)
	} else {
		if item != nil {
			json.NewEncoder(w).Encode(item)
		}
		if items != nil {
			json.NewEncoder(w).Encode(items)
		}
	}
}

func ItemsWithCategoriesAndSubcategoriesHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	var catID int
	var err error

	if catID, err = strconv.Atoi(vars["category_id"]); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if items, err := GetItemsWithCategoriesAndSubcategories(catID); err != nil {
		http.Error(w, err.Error(), 404)
	} else {
		json.NewEncoder(w).Encode(items)
	}
}

func ItemsHandlerStatusStockCat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var startCatID int
	var stockID int
	var status int
	var err error

	if startCatID, err = strconv.Atoi(vars["start_cat_id"]); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if stockID, err = strconv.Atoi(vars["stock_id"]); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if status, err = strconv.Atoi(vars["status"]); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if items, err := GetItemsWithStatusStockCategory(status, stockID, startCatID); err != nil {
		http.Error(w, err.Error(), 404)
	} else {
		json.NewEncoder(w).Encode(items)
	}

}

//post items
func PostItemHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var item *Item
	err := decoder.Decode(&item)
	if err != nil {
		fmt.Println("Error Decoding Form")
		http.Error(w, err.Error(), 500)
		return
	}

	//adding item to the wishlist, status code will be 3 == wishlist
	if err = PostItem(item, 3); err != nil {
		http.Error(w, err.Error(), 500)
	} else {
		items, _ := GetItems()
		json.NewEncoder(w).Encode(items) //should return 201
	}
}

func PurchaseWishlistItemHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	var itemID int
	var item *Item
	var err error

	if itemID, err = strconv.Atoi(vars["item_id"]); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if item, err = GetItem(itemID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//first check that the user is a stock_taker
	if _, list, err := IsUserStockTaker(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	} else {

		//it's a stock taker, check that he owns the stock
		flag := intInSlice(item.StockId, list)
		if flag {
			if err := UpdateItemStatusToPending(item); err != nil {
				http.Error(w, err.Error(), 500)
			} else {
				json.NewEncoder(w).Encode("Done") //TODO: what should I put here?
			}

		} else {
			http.Error(w, errors.New("This stock is not yours bro").Error(), http.StatusUnauthorized)
		}
	}
}

func PutPurchasedItemIntoStockHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	var itemID int
	var item *Item
	var err error

	if itemID, err = strconv.Atoi(vars["item_id"]); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if item, err = GetItem(itemID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//first check that the user is a stock_taker
	if _, list, err := IsUserStockTaker(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	} else {

		//it's a stock taker, check that he owns the stock
		flag := intInSlice(item.StockId, list)
		if flag {
			if err := PutItemIntoStock(item); err != nil {
				http.Error(w, err.Error(), 500)
			} else {
				json.NewEncoder(w).Encode("Done") //TODO: what should I put here?
			}

		} else {
			http.Error(w, errors.New("This stock is not yours bro").Error(), http.StatusUnauthorized)
		}
	}
}

func PutNewItemIntoStockHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	var username string
	var err error
	var ok bool

	if username, ok = vars["username"]; ok == false {
		http.Error(w, err.Error(), 500)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var item *Item
	err = decoder.Decode(&item)
	if err != nil {
		fmt.Println("Error Decoding Form")
		http.Error(w, err.Error(), 500)
		return
	}

	//first check that the user is a stock_taker
	if _, list, err := IsUserStockTaker(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	} else {
		//it's a stock taker, check that he owns the stock
		flag := intInSlice(item.StockId, list)
		if flag {

			if err = PostItem(item, 1); err != nil {
				http.Error(w, err.Error(), 500)
			} else {
				if user, err := GetUserByUsername(username); err != nil {
					http.Error(w, err.Error(), 500)
				} else {
					if err = AddAmountToUserBalance(user, item.Coins*item.Quantity); err != nil {
						http.Error(w, err.Error(), 500)
					} else {
						json.NewEncoder(w).Encode(user)
					}
				}
			}

		} else {
			http.Error(w, errors.New("This stock is not yours bro").Error(), http.StatusUnauthorized)
		}
	}

}

//test
//curl -H "Content-Type: application/json" -H "Authorization: Bearer uIh381xmpmRb9sNW62IAyV1GGvU=" -X POST http://localhost:8080/purchase/3/5
func PurchaseItemHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	var itemID int
	var quantity int
	var err error

	if itemID, err = strconv.Atoi(vars["item_id"]); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if quantity, err = strconv.Atoi(vars["quantity"]); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if user, err := GetUserFromToken(r); err != nil {
		http.Error(w, err.Error(), 500)
	} else {
		if item, err := PurchaseItem(itemID, quantity, user); err != nil {
			http.Error(w, err.Error(), 500)
		} else {

			//put right quantity (not the global one) and the right coins
			item.Coins = item.Coins * quantity
			item.Quantity = quantity
			//create receipt
			if receipt, err := ReceiptForItem(item); err != nil {

				errStr := fmt.Sprintf("%s. Please contact you stock taker. stock_id: %d item_id: %d, quantity_purchased: %d",
					err.Error(), item.StockId, item.Id, quantity)

				json.NewEncoder(w).Encode(errStr)
			} else {
				if image, err := QRImageFromReceipt(receipt); err != nil {
					json.NewEncoder(w).Encode(receipt)
				} else {
					writeImage(w, &image)
				}
			}
		}
	}
}

//tests
func TestHandler(w http.ResponseWriter, r *http.Request) {

	item, _ := GetItem(1)
	receipt, _ := ReceiptForItem(item)
	if image, err := QRImageFromReceipt(receipt); err != nil {
		http.Error(w, err.Error(), 500)
	} else {
		writeImage(w, &image)
	}

}

//private helpers
func intInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func writeImage(w http.ResponseWriter, img *image.Image) {

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, *img, nil); err != nil {
		log.Println("unable to encode image.")
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		log.Println("unable to write image.")
	}
}
