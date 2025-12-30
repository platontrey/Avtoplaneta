package main

import (
	"time"
)

// Order представляет заказ в системе
type Order struct {
	ID                 uint        `json:"id" gorm:"primaryKey"`
	CustomerID         int         `json:"customer_id"`
	SellerID           uint        `json:"seller_id"` // ID пользователя-продавца
	Seller             string      `json:"seller"`    // Имя продавца (для отображения)
	Part               string      `json:"part"`      // Название детали (для отображения)
	PartID             uint        `json:"part_id"`   // ID детали из инвентаря
	Location           string      `json:"location"`  // Склад, где находится запчасть
	BuyerNumber        string      `json:"buyer_number"`
	Status             string      `json:"status"` // red, brown, yellow, green
	StatusText         string      `json:"status_text"`
	AutoDeleted        bool        `json:"auto_deleted" gorm:"default:false"` // Автоматически удалённый заказ для статистики продаж
	CreatedAt          time.Time   `json:"created_at"`
	CreatedAtFormatted string      `json:"created_at_formatted"` // Форматированная дата для фронтенда
	TimeAgo            string      `json:"time_ago"`             // Относительное время (например, "3 дня назад")
	Items              []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
}

// OrderItem представляет позицию в заказе
type OrderItem struct {
	ID       uint    `json:"id" gorm:"primaryKey"`
	OrderID  uint    `json:"order_id"`
	PartID   uint    `json:"part_id"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"` // Цена на момент создания заказа
}

// Part представляет запчасть в системе (упрощенная версия для orders-service)
type Part struct {
	ID       uint    `gorm:"primaryKey"`
	Quantity int     `gorm:"not null"`
	Price    float64 `gorm:"not null"`
	Location string  `json:"location,omitempty"`
	Photo    string  `json:"photo,omitempty"`
}

// SalesHistory хранит исторические данные продаж по месяцам
type SalesHistory struct {
	ID       uint      `gorm:"primaryKey"`
	Month    string    `gorm:"not null"` // Формат: "2023-12"
	Sales    float64   `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// MonthlySales представляет продажи за месяц
type MonthlySales struct {
	Month string  `json:"month"`
	Sales float64 `json:"sales"`
}
