package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Имя продавца живёт в parts.salesman копией, снятой при создании запчасти.
// В заказах такую копию достаточно подменять при чтении, а здесь этого мало:
// по продавцу ещё и фильтруют, и фильтр уходит в Elasticsearch, где лежит та же
// копия. Показывать новое имя и не находить по нему — хуже, чем не чинить вовсе.
//
// Поэтому расхождение не прикрывается, а устраняется: при чтении имя
// подставляется актуальное, а на несовпадение отправляется событие, по которому
// consumer обновляет строки продавца в базе и переиндексирует их. Переименования
// редки, так что починка срабатывает один раз после смены имени.
const (
	sellerRenamedEventType = "seller_renamed"
	// Окно, в течение которого повторная просьба починить того же продавца
	// считается лишней: consumer успевает разобрать очередь, а список инвентаря
	// за это время могут открыть десятки раз.
	sellerRepairCooldown = 5 * time.Minute
)

type sellerRepairThrottle struct {
	mu       sync.Mutex
	lastSent map[int64]time.Time
	now      func() time.Time
}

func newSellerRepairThrottle() *sellerRepairThrottle {
	return &sellerRepairThrottle{lastSent: map[int64]time.Time{}, now: time.Now}
}

func (t *sellerRepairThrottle) allow(sellerID int64) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	if sent, ok := t.lastSent[sellerID]; ok && now.Sub(sent) < sellerRepairCooldown {
		return false
	}
	t.lastSent[sellerID] = now
	return true
}

// withCurrentSalesman подставляет актуальные имена продавцов и просит починить
// те строки, где копия разошлась с справочником.
func (s *inventoryService) withCurrentSalesman(ctx context.Context, parts []Part) {
	if s.users == nil || len(parts) == 0 {
		return
	}

	names := s.users.Names(ctx)
	if len(names) == 0 {
		return
	}

	stale := map[int64]string{}
	for i := range parts {
		sellerID := parts[i].SellerID
		if sellerID <= 0 {
			continue
		}
		name, ok := names[sellerID]
		if !ok || name == "" {
			// Продавца уже нет в справочнике: оставляем имя из строки,
			// иначе старые запчасти останутся без продавца вовсе.
			continue
		}
		if parts[i].Salesman != name {
			stale[sellerID] = name
			parts[i].Salesman = name
		}
	}

	for sellerID, name := range stale {
		s.requestSellerRepair(ctx, sellerID, name)
	}
}

func (s *inventoryService) requestSellerRepair(ctx context.Context, sellerID int64, name string) {
	if s.redis == nil || s.repairThrottle == nil || !s.repairThrottle.allow(sellerID) {
		return
	}

	err := s.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: "events:orders",
		Values: map[string]interface{}{
			"type":      sellerRenamedEventType,
			"seller_id": fmt.Sprintf("%d", sellerID),
			"name":      name,
		},
	}).Err()
	if err != nil {
		logrus.WithError(err).WithField("seller_id", sellerID).Warn("Failed to publish seller_renamed")
		return
	}

	logrus.WithFields(logrus.Fields{
		"seller_id": sellerID,
		"name":      name,
	}).Info("Копия имени продавца устарела, отправлена починка")
}
