// Package userdirectory отдаёт актуальные имена пользователей по идентификатору.
//
// Имя — это атрибут человека, а не строки заказа или запчасти. Поэтому хранимой
// правдой остаётся seller_id, а имя подставляется при чтении: тогда
// переименование в админке видно везде сразу, и не нужно догонять копии
// рассылкой событий.
//
// Снимок имени уместен только там, где запись обязана пережить удаление
// пользователя или где важно, как он назывался тогда, — например в журнале
// аудита. Такие места этим пакетом пользоваться не должны.
package userdirectory

import (
	"context"
	"sync"
	"time"
)

// Fetch возвращает карту «идентификатор — имя». Сервис подставляет сюда свой
// способ сходить в auth-service, поэтому пакет не зависит ни от gRPC, ни от HTTP.
type Fetch func(context.Context) (map[int64]string, error)

// Directory кэширует справочник на короткое время: сотрудников десятки,
// переименования редки, а спрашивают имена на каждый список заказов.
type Directory struct {
	fetch Fetch
	ttl   time.Duration
	now   func() time.Time

	mu        sync.RWMutex
	names     map[int64]string
	refreshed time.Time
}

func New(fetch Fetch, ttl time.Duration) *Directory {
	return &Directory{fetch: fetch, ttl: ttl, now: time.Now}
}

// Names возвращает карту имён, обновляя её не чаще раза в ttl.
//
// Ошибка похода в auth-service не считается фатальной: справочник имён —
// украшение ответа, а не его суть. В этом случае отдаётся последняя известная
// карта (возможно пустая), и вызывающий оставляет то имя, что лежит в строке.
func (d *Directory) Names(ctx context.Context) map[int64]string {
	if d == nil {
		return nil
	}

	d.mu.RLock()
	fresh := d.names != nil && d.now().Sub(d.refreshed) < d.ttl
	cached := d.names
	d.mu.RUnlock()
	if fresh {
		return cached
	}

	names, err := d.fetch(ctx)
	if err != nil {
		return cached
	}

	d.mu.Lock()
	d.names = names
	d.refreshed = d.now()
	d.mu.Unlock()
	return names
}

// Name возвращает актуальное имя пользователя. Второе значение — известен ли он:
// удалённого сотрудника в справочнике уже нет, и затирать его имя пустой
// строкой нельзя, иначе старые заказы останутся без продавца.
func (d *Directory) Name(ctx context.Context, id int64) (string, bool) {
	if id <= 0 {
		return "", false
	}
	name, ok := d.Names(ctx)[id]
	if !ok || name == "" {
		return "", false
	}
	return name, true
}
