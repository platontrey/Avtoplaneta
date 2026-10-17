// Package constants содержит общесистемные доменные константы и инварианты Avtoplaneta
package constants

// Инварианты количества деталей на складе (Inventory Invariants)
const (
	// PartQuantitySold маркер полностью проданной/списанной детали (-1).
	// Такие детали сокрыты из поисковой выдачи, каталогов и XML-выгрузок.
	PartQuantitySold = -1

	// PartQuantityDefective маркер дефектных ведомостей / спецпозиций (0).
	// ВАЖНО: согласно AGENTS.md, детали с quantity == 0 ОБЯЗАНЫ оставаться
	// видимыми в поиске, UI и выгрузках (WHERE quantity >= 0).
	PartQuantityDefective = 0
)

// Роли пользователей в системе (User Roles)
const (
	RoleAdmin    = "admin"
	RoleManager  = "manager"
	RoleOperator = "operator"
)

// Статусы заказов (Order Statuses)
const (
	OrderStatusPending    = "pending"
	OrderStatusProcessing = "processing"
	OrderStatusCompleted  = "completed"
	OrderStatusCancelled  = "cancelled"
)

// Стримы брокера Redis Streams
const (
	StreamOrdersEvents = "events:orders"
	StreamPartsEvents  = "events:parts"
)

// Каналы PostgreSQL LISTEN / NOTIFY для Transactional Outbox
const (
	OutboxChannelParts  = "outbox_events_channel"
	OutboxChannelOrders = "orders_outbox_channel"
	OutboxChannelAuth   = "auth_outbox_channel"
)

// IsVisiblePart возвращает true, если деталь должна отображаться в каталоге/поиске
func IsVisiblePart(quantity int32) bool {
	return quantity >= PartQuantityDefective
}

// IsValidRole проверяет корректность роли
func IsValidRole(role string) bool {
	switch role {
	case RoleAdmin, RoleManager, RoleOperator:
		return true
	default:
		return false
	}
}
