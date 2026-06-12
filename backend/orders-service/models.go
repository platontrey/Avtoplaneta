package main

import (
	"time"
)

// Order представляет заказ в системе
type Order struct {
	ID                 int64       `json:"id"`
	CustomerID         int64       `json:"customer_id"`
	SellerID           int64       `json:"seller_id"` // ID пользователя-продавца
	Seller             string      `json:"seller"`    // Имя продавца (для отображения)
	Part               string      `json:"part"`      // Название детали (для отображения)
	PartID             int64       `json:"part_id"`   // ID детали из инвентаря
	Location           string      `json:"location"`  // Склад, где находится запчасть
	BuyerNumber        string      `json:"buyer_number"`
	Status             string      `json:"status"` // red, brown, yellow, green
	StatusText         string      `json:"status_text"`
	AutoDeleted        bool        `json:"auto_deleted"` // Автоматически удалённый заказ для статистики продаж
	CreatedAt          time.Time   `json:"created_at"`
	CreatedAtFormatted string      `json:"created_at_formatted"` // Форматированная дата для фронтенда
	TimeAgo            string      `json:"time_ago"`             // Относительное время (например, "3 дня назад")
	Items              []OrderItem `json:"items"`
}

// OrderItem представляет позицию в заказе
type OrderItem struct {
	ID       int64   `json:"id"`
	OrderID  int64   `json:"order_id"`
	PartID   int64   `json:"part_id"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"` // Цена на момент создания заказа
}

// Part представляет запчасть в системе (упрощенная версия для orders-service)
type Part struct {
	ID       int64   `json:"id"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Location string  `json:"location,omitempty"`
	Photo    string  `json:"photo,omitempty"`
}

// SalesHistory хранит исторические данные продаж по месяцам
type SalesHistory struct {
	ID        int64     `json:"id"`
	Month     string    `json:"month"` // Формат: "2023-12"
	Sales     float64   `json:"sales"`
	CreatedAt time.Time `json:"created_at"`
}

// MonthlySales представляет продажи за месяц
type MonthlySales struct {
	Month string  `json:"month"`
	Sales float64 `json:"sales"`
}
