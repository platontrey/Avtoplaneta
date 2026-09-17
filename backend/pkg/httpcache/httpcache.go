// Package httpcache реализует условные GET-запросы: сервер помечает ответ
// версией, клиент присылает её обратно, и если версия совпала — тело не
// передаётся вовсе.
//
// Пакет лежит в общем модуле, чтобы все сервисы обращались с ETag одинаково.
// Зависимости от веб-фреймворка здесь нет: функции работают со стандартными
// http.ResponseWriter и *http.Request, а gin.Context отдаёт и то, и другое.
package httpcache

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ServeVersioned выставляет заголовки кэширования и, если клиент уже держит
// эту версию, сам отвечает 304. Возвращает true, когда ответ уже отправлен и
// обработчику больше делать нечего:
//
//	if httpcache.ServeVersioned(c.Writer, c.Request, version, 5*time.Minute) {
//		return
//	}
//
// Вызывать нужно до того, как собрано тело ответа: весь смысл в том, чтобы на
// повторном запросе не делать дорогую работу.
func ServeVersioned(w http.ResponseWriter, r *http.Request, version string, maxAge time.Duration) bool {
	if version == "" {
		// Валидатора нет — не выставляем пустой ETag, иначе клиент закэширует
		// ответ по заведомо бессмысленному ключу.
		return false
	}

	etag := Quote(version)
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(int(maxAge.Seconds()))+", must-revalidate")

	if Matches(r.Header.Get("If-None-Match"), etag) {
		// У 304 не должно быть тела: часть прокси на такой ответ реагирует плохо.
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	return false
}

// Quote приводит значение к виду, который требует спецификация: ETag всегда
// заключён в кавычки. Уже закавыченное значение (в том числе слабое, W/"...")
// возвращается как есть.
func Quote(value string) string {
	if strings.HasPrefix(value, `"`) || strings.HasPrefix(value, `W/"`) {
		return value
	}
	return `"` + value + `"`
}

// Matches разбирает If-None-Match по правилам RFC 9110.
//
// Наивное сравнение строк здесь не годится по трём причинам: клиент вправе
// прислать несколько версий через запятую, прокси вправе ослабить валидатор
// до W/"...", а звёздочка означает «любая версия, лишь бы ресурс существовал».
// Во всех трёх случаях строгое равенство молча вернуло бы 200, и кэш перестал
// бы работать, не сломавшись заметно.
func Matches(ifNoneMatch, etag string) bool {
	ifNoneMatch = strings.TrimSpace(ifNoneMatch)
	if ifNoneMatch == "" || etag == "" {
		return false
	}
	if ifNoneMatch == "*" {
		return true
	}

	want := weaken(etag)
	for _, candidate := range strings.Split(ifNoneMatch, ",") {
		if weaken(strings.TrimSpace(candidate)) == want {
			return true
		}
	}
	return false
}

// weaken убирает префикс слабого валидатора: для условных GET сравнение
// ведётся по слабым правилам, то есть W/"abc" и "abc" — одна и та же версия.
func weaken(etag string) string {
	return strings.TrimPrefix(strings.TrimSpace(etag), "W/")
}
