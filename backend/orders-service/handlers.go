package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func getOrdersHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("getOrdersHandler called: Method=%s, URL=%s", r.Method, r.URL.Path)

	// Добавляем CORS заголовки напрямую
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

	// Mark expired orders (older than 14 days) as auto_deleted for sales statistics
	now := time.Now()
	fourteenDaysAgo := now.AddDate(0, 0, -14)
	if err := db.Model(&Order{}).Where("created_at <= ? AND auto_deleted = ?", fourteenDaysAgo, false).Update("auto_deleted", true).Error; err != nil {
		log.Printf("Error marking expired orders as auto_deleted: %v", err)
		// Continue anyway, don't fail the request
	} else {
		log.Printf("Marked expired orders as auto_deleted for sales statistics")
	}

	var orders []Order
	if err := db.Preload("Items").Where("auto_deleted = ?", false).Find(&orders).Error; err != nil {
		log.Printf("Error fetching orders: %v", err)
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}

	log.Printf("getOrdersHandler: Found %d orders", len(orders))

	// Заполнить форматированные поля дат и location для каждой запчасти
	for i := range orders {
		orders[i].CreatedAtFormatted = orders[i].CreatedAt.Format("2006-01-02 15:04:05")
		orders[i].TimeAgo = formatTimeAgo(now.Sub(orders[i].CreatedAt))

		// Получить location из Parts для каждого item в заказе
		if len(orders[i].Items) > 0 {
			// Для простоты берем location первой запчасти, так как в заказе может быть несколько
			var part Part
			if err := db.Where("id = ?", orders[i].Items[0].PartID).First(&part).Error; err != nil {
				log.Printf("Error fetching part %d for location: %v", orders[i].Items[0].PartID, err)
				orders[i].Location = "Неизвестно"
			} else {
				orders[i].Location = part.Location
			}
		} else {
			orders[i].Location = "Нет деталей"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	// Добавляем CORS заголовки напрямую
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

	var requestData struct {
		CustomerID  int    `json:"customer_id"`
		Part        string `json:"part"`
		PartID      uint   `json:"part_id"`
		BuyerNumber string `json:"buyer_number"`
		Items       []struct {
			PartID   uint `json:"part_id"`
			Quantity int  `json:"quantity"`
		} `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if requestData.BuyerNumber == "" {
		http.Error(w, "Buyer number is required", http.StatusBadRequest)
		return
	}

	// For multi-part orders, we need at least one item
	if len(requestData.Items) == 0 {
		http.Error(w, "At least one part must be selected", http.StatusBadRequest)
		return
	}

	// Получить ID текущего пользователя из заголовка (установленного authMiddleware)
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Получить информацию о пользователе из auth-service
	sellerName, err := getUserName(uint(userID))
	if err != nil {
		log.Printf("Failed to get user name for ID %d: %v", userID, err)
		http.Error(w, "Failed to get seller information", http.StatusInternalServerError)
		return
	}

	// Create the order with color-coded status (starts as red)
	order := Order{
		CustomerID:  requestData.CustomerID,
		SellerID:    uint(userID),
		Seller:      sellerName,
		Part:        requestData.Part,
		PartID:      requestData.PartID,
		BuyerNumber: requestData.BuyerNumber,
		Status:      "red",
		StatusText:  "Need to order transport company",
		CreatedAt:   time.Now(),
	}

	if err := db.Create(&order).Error; err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	// Create order items and decrease part quantities
	for _, item := range requestData.Items {
		// Get part price before creating order item
		var part Part
		if err := db.Where("id = ?", item.PartID).First(&part).Error; err != nil {
			log.Printf("Failed to get part %d for price: %v", item.PartID, err)
			http.Error(w, "Failed to get part information", http.StatusInternalServerError)
			return
		}

		orderItem := OrderItem{
			OrderID:  order.ID,
			PartID:   item.PartID,
			Quantity: item.Quantity,
			Price:    part.Price, // Save price at order creation time
		}

		if err := db.Create(&orderItem).Error; err != nil {
			http.Error(w, "Failed to create order item", http.StatusInternalServerError)
			return
		}

		// Decrease the quantity of the part in the parts service database
		// Since we're using the same database, we can update directly
		// If the part quantity becomes 0 or negative after decrease, set it to -1 to hide it
		if err := db.Model(&Part{}).Where("id = ?", item.PartID).Update("quantity", db.Raw("CASE WHEN quantity - ? <= 0 THEN -1 ELSE quantity - ? END", item.Quantity, item.Quantity)).Error; err != nil {
			log.Printf("Failed to decrease quantity for part %d: %v", item.PartID, err)
			http.Error(w, "Failed to update part quantity", http.StatusInternalServerError)
			return
		}

		log.Printf("Decreased quantity of part %d by %d", item.PartID, item.Quantity)
	}

	// Load the complete order with items
	if err := db.Preload("Items").First(&order, order.ID).Error; err != nil {
		http.Error(w, "Failed to load created order", http.StatusInternalServerError)
		return
	}

	// Заполнить форматированные поля дат для нового заказа
	now := time.Now()
	order.CreatedAtFormatted = order.CreatedAt.Format("2006-01-02 15:04:05")
	order.TimeAgo = formatTimeAgo(now.Sub(order.CreatedAt))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func updateOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIDStr := vars["id"]
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var requestData struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	// Validate status and set status text
	var statusText string
	validStatuses := map[string]string{
		"red":    "Need to order transport company",
		"brown":  "Waiting for response",
		"yellow": "Need to deliver",
		"green":  "Transported",
	}

	statusText, isValid := validStatuses[requestData.Status]
	if !isValid {
		http.Error(w, "Invalid status. Valid statuses: red, brown, yellow, green", http.StatusBadRequest)
		return
	}

	// Update order status and status text
	updates := map[string]interface{}{
		"status":      requestData.Status,
		"status_text": statusText,
	}

	if err := db.Model(&Order{}).Where("id = ?", orderID).Updates(updates).Error; err != nil {
		log.Printf("Failed to update order status: %v", err)
		http.Error(w, "Failed to update order status", http.StatusInternalServerError)
		return
	}

	// Return updated order
	var order Order
	if err := db.Preload("Items").First(&order, orderID).Error; err != nil {
		log.Printf("Failed to fetch updated order: %v", err)
		http.Error(w, "Failed to fetch updated order", http.StatusInternalServerError)
		return
	}

	// Заполнить форматированные поля дат и location для обновленного заказа
	now := time.Now()
	order.CreatedAtFormatted = order.CreatedAt.Format("2006-01-02 15:04:05")
	order.TimeAgo = formatTimeAgo(now.Sub(order.CreatedAt))

	// Получить location из Parts
	if len(order.Items) > 0 {
		var part Part
		if err := db.Where("id = ?", order.Items[0].PartID).First(&part).Error; err != nil {
			log.Printf("Error fetching part %d for location: %v", order.Items[0].PartID, err)
			order.Location = "Неизвестно"
		} else {
			order.Location = part.Location
		}
	} else {
		order.Location = "Нет деталей"
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func completeOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIDStr := vars["id"]
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	// Check if order exists
	var order Order
	if err := db.Preload("Items").First(&order, orderID).Error; err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Get all parts for this order to delete them
	for _, item := range order.Items {
		// Get the part details to delete photo
		var part Part
		if err := db.Where("id = ?", item.PartID).First(&part).Error; err != nil {
			log.Printf("Failed to find part %d for deletion: %v", item.PartID, err)
			continue
		}

		// Delete photo file if exists
		if part.Photo != "" {
			// Extract filename from path (remove "/uploads/" prefix) and build path relative to parts-service
			filename := strings.TrimPrefix(part.Photo, "/uploads/")
			photoPath := "../parts-service/uploads/" + filename
			log.Printf("DEBUG: Attempting to delete photo file: %s", photoPath)
			if err := os.Remove(photoPath); err != nil && !os.IsNotExist(err) {
				log.Printf("Warning: Failed to delete photo file %s: %v", photoPath, err)
			} else if err == nil {
				log.Printf("Successfully deleted photo file: %s", photoPath)
			} else {
				log.Printf("Photo file %s does not exist (already deleted)", photoPath)
			}
		}

		// Delete the part from database
		if err := db.Delete(&Part{}, item.PartID).Error; err != nil {
			log.Printf("Failed to delete part %d: %v", item.PartID, err)
			http.Error(w, "Failed to delete part", http.StatusInternalServerError)
			return
		}
		log.Printf("Successfully deleted part %d from database", item.PartID)
	}

	// Mark order as completed for sales statistics
	log.Printf("DEBUG: Setting auto_deleted=true for order %d", orderID)
	if err := db.Model(&Order{}).Where("id = ?", orderID).Update("auto_deleted", true).Error; err != nil {
		log.Printf("Failed to complete order: %v", err)
		http.Error(w, "Failed to complete order", http.StatusInternalServerError)
		return
	}
	log.Printf("DEBUG: Successfully set auto_deleted=true for order %d", orderID)

	// Debug: Verify the update
	var verifyOrder Order
	if err := db.First(&verifyOrder, orderID).Error; err != nil {
		log.Printf("DEBUG: Failed to verify order update: %v", err)
	} else {
		log.Printf("DEBUG: Verified order %d auto_deleted=%t", orderID, verifyOrder.AutoDeleted)
	}

	// Return updated order
	if err := db.Preload("Items").First(&order, orderID).Error; err != nil {
		log.Printf("Failed to fetch completed order: %v", err)
		http.Error(w, "Failed to fetch completed order", http.StatusInternalServerError)
		return
	}

	// Заполнить форматированные поля дат
	now := time.Now()
	order.CreatedAtFormatted = order.CreatedAt.Format("2006-01-02 15:04:05")
	order.TimeAgo = formatTimeAgo(now.Sub(order.CreatedAt))

	// Location is not available since parts are deleted
	order.Location = "Запчасти удалены"

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func deleteOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIDStr := vars["id"]
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	// Check if order exists
	var order Order
	if err := db.First(&order, orderID).Error; err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Get order items to restore part quantities before deleting
	var orderItems []OrderItem
	if err := db.Where("order_id = ?", orderID).Find(&orderItems).Error; err != nil {
		log.Printf("Failed to fetch order items for order %d: %v", orderID, err)
		http.Error(w, "Failed to fetch order items", http.StatusInternalServerError)
		return
	}

	// Restore part quantities - set back to original quantity, not just add
	for _, item := range orderItems {
		// Get the original part to restore its quantity
		var part Part
		if err := db.Where("id = ?", item.PartID).First(&part).Error; err != nil {
			log.Printf("Failed to find part %d for restoration: %v", item.PartID, err)
			continue
		}

		// If part was hidden (quantity = -1), restore to 0, otherwise add back the ordered quantity
		newQuantity := part.Quantity + item.Quantity
		if part.Quantity == -1 {
			newQuantity = item.Quantity // Restore to the ordered quantity
		}

		if err := db.Model(&Part{}).Where("id = ?", item.PartID).Update("quantity", newQuantity).Error; err != nil {
			log.Printf("Failed to restore quantity for part %d: %v", item.PartID, err)
			http.Error(w, "Failed to restore part quantity", http.StatusInternalServerError)
			return
		}
	}

	// Delete order items first (due to foreign key constraint)
	if err := db.Where("order_id = ?", orderID).Delete(&OrderItem{}).Error; err != nil {
		log.Printf("Failed to delete order items for order %d: %v", orderID, err)
		http.Error(w, "Failed to delete order items", http.StatusInternalServerError)
		return
	}

	// Удалить заказ
	if err := db.Delete(&Order{}, orderID).Error; err != nil {
		log.Printf("Failed to delete order %d: %v", orderID, err)
		http.Error(w, "Failed to delete order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": "Order deleted successfully"}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Пропускаем OPTIONS запросы для CORS preflight
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		session, err := store.Get(r, "auth-session")
		if err != nil {
			log.Printf("SECURITY: Invalid session from %s: %v", r.RemoteAddr, err)
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		userID, ok := session.Values["user_id"]
		if !ok || userID == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Проверка времени сессии
		if loginTime, ok := session.Values["login_time"].(int64); ok {
			if time.Now().Unix()-loginTime > 86400*30 { // 30 days
				log.Printf("SECURITY: Session expired for user ID %v from %s", userID, r.RemoteAddr)
				http.Error(w, "Session expired", http.StatusUnauthorized)
				return
			}
		}

		// For orders service, we just need to ensure user is authenticated
		// The actual user data would be fetched from auth service in production
		r.Header.Set("X-User-ID", strconv.FormatUint(uint64(userID.(uint)), 10))

		next.ServeHTTP(w, r)
	})
}

// formatTimeAgo форматирует время в относительный формат (например, "3 дня назад")
func formatTimeAgo(duration time.Duration) string {
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		if days == 1 {
			return "1 день назад"
		} else if days < 5 {
			return fmt.Sprintf("%d дня назад", days)
		} else {
			return fmt.Sprintf("%d дней назад", days)
		}
	} else if hours > 0 {
		if hours == 1 {
			return "1 час назад"
		} else if hours < 5 {
			return fmt.Sprintf("%d часа назад", hours)
		} else {
			return fmt.Sprintf("%d часов назад", hours)
		}
	} else if minutes > 0 {
		if minutes == 1 {
			return "1 минуту назад"
		} else if minutes < 5 {
			return fmt.Sprintf("%d минуты назад", minutes)
		} else {
			return fmt.Sprintf("%d минут назад", minutes)
		}
	} else {
		return "только что"
	}
}

func adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.Header.Get("X-User-ID")
		if userIDStr == "" {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// In a real implementation, you'd fetch user data from auth service
		// For now, we'll assume admin check is done at auth service level
		// This is a simplified version for the orders service

		// For demo purposes, we'll allow all authenticated users to access admin routes
		// In production, this should validate admin status
		log.Printf("SECURITY: Admin access granted for user ID %d from %s", userID, r.RemoteAddr)

		next.ServeHTTP(w, r)
	})
}

// getUserName получает имя пользователя напрямую из auth-service БД
func getUserName(userID uint) (string, error) {
	// Создаем подключение к auth-service БД (используем ту же БД)
	type AuthUser struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}

	var user AuthUser

	// Запрашиваем пользователя по ID из той же БД
	if err := db.Table("users").Where("id = ?", userID).First(&user).Error; err != nil {
		log.Printf("Failed to fetch user name for ID %d: %v", userID, err)
		return fmt.Sprintf("Пользователь %d", userID), nil // Fallback
	}

	return user.Name, nil
}

func addOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIDStr := vars["id"]
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	// Добавляем CORS заголовки напрямую
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

	var requestData struct {
		PartID   uint `json:"part_id"`
		Quantity int  `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if requestData.PartID == 0 || requestData.Quantity <= 0 {
		http.Error(w, "Part ID and positive quantity are required", http.StatusBadRequest)
		return
	}

	// Check if order exists
	var order Order
	if err := db.First(&order, orderID).Error; err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Check if part already exists in order
	var existingItem OrderItem
	if err := db.Where("order_id = ? AND part_id = ?", orderID, requestData.PartID).First(&existingItem).Error; err == nil {
		// Part already exists, update quantity
		existingItem.Quantity += requestData.Quantity
		if err := db.Save(&existingItem).Error; err != nil {
			http.Error(w, "Failed to update order item", http.StatusInternalServerError)
			return
		}
	} else {
		// Get part price before creating order item
		var part Part
		if err := db.Where("id = ?", requestData.PartID).First(&part).Error; err != nil {
			log.Printf("Failed to get part %d for price: %v", requestData.PartID, err)
			http.Error(w, "Failed to get part information", http.StatusInternalServerError)
			return
		}

		// Part doesn't exist, create new item
		orderItem := OrderItem{
			OrderID:  uint(orderID),
			PartID:   requestData.PartID,
			Quantity: requestData.Quantity,
			Price:    part.Price, // Save price at order creation time
		}

		if err := db.Create(&orderItem).Error; err != nil {
			http.Error(w, "Failed to create order item", http.StatusInternalServerError)
			return
		}
	}

	// Decrease the quantity of the part in the parts service database
	if err := db.Model(&Part{}).Where("id = ?", requestData.PartID).Update("quantity", db.Raw("CASE WHEN quantity - ? <= 0 THEN -1 ELSE quantity - ? END", requestData.Quantity, requestData.Quantity)).Error; err != nil {
		log.Printf("Failed to decrease quantity for part %d: %v", requestData.PartID, err)
		http.Error(w, "Failed to update part quantity", http.StatusInternalServerError)
		return
	}

	log.Printf("Added %d quantity of part %d to order %d", requestData.Quantity, requestData.PartID, orderID)

	// Load the updated order with items
	if err := db.Preload("Items").First(&order, orderID).Error; err != nil {
		http.Error(w, "Failed to load updated order", http.StatusInternalServerError)
		return
	}

	// Заполнить форматированные поля дат и location для обновленного заказа
	now := time.Now()
	order.CreatedAtFormatted = order.CreatedAt.Format("2006-01-02 15:04:05")
	order.TimeAgo = formatTimeAgo(now.Sub(order.CreatedAt))

	// Получить location из Parts
	if len(order.Items) > 0 {
		var part Part
		if err := db.Where("id = ?", order.Items[0].PartID).First(&part).Error; err != nil {
			log.Printf("Error fetching part %d for location: %v", order.Items[0].PartID, err)
			order.Location = "Неизвестно"
		} else {
			order.Location = part.Location
		}
	} else {
		order.Location = "Нет деталей"
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
