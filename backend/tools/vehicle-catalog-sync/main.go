// Command vehicle-catalog-sync собирает справочник марок и моделей с drom.ru
// и перезаписывает parts-service/vehicles/vehicles.json.
//
// Команда запускается руками (или по расписанию CI), а не сервисом: результат
// коммитится в репозиторий и вшивается в бинарь через go:embed. Благодаря этому
// у сервиса нет внешней зависимости в рантайме, а у клиентов — одного источника
// на двоих, который не надо дублировать в TypeScript и Dart.
//
// robots.txt drom.ru не запрещает /catalog/; единственный объявленный там
// Crawl-Delay равен одной секунде, его и берём по умолчанию.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	catalogURL       = "https://www.drom.ru/catalog/"
	defaultUserAgent = "AvtoplanetaCatalogSync/1.0 (+https://avtoplaneta.ru; contact: platontrey5@gmail.com)"
)

type vehicleModel struct {
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

type vehicleBrand struct {
	Name   string         `json:"name"`
	Slug   string         `json:"slug,omitempty"`
	Models []vehicleModel `json:"models"`
}

type vehicleCatalog struct {
	Version     string         `json:"version"`
	Source      string         `json:"source,omitempty"`
	GeneratedAt string         `json:"generated_at,omitempty"`
	Brands      []vehicleBrand `json:"brands"`
}

// Служебные разделы каталога, которые лежат на тех же путях, что и марки.
var notABrand = map[string]bool{
	"all": true, "new": true, "used": true, "compare": true,
	"reviews": true, "search": true, "rating": true, "sale": true,
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:[-_+][a-z0-9]+)*$`)

func main() {
	out := flag.String("out", "../../parts-service/vehicles/vehicles.json", "путь к vehicles.json")
	delay := flag.Duration("delay", time.Second, "пауза между запросами")
	timeout := flag.Duration("timeout", 30*time.Second, "таймаут одного запроса")
	retries := flag.Int("retries", 3, "число повторов на запрос")
	limit := flag.Int("limit", 0, "обработать только первые N марок (0 — все); для отладки")
	userAgent := flag.String("user-agent", defaultUserAgent, "User-Agent")
	dump := flag.String("dump", "", "сохранить HTML индекса каталога в файл; для диагностики разбора")
	force := flag.Bool("force", false, "записать результат, даже если марок стало заметно меньше")
	flag.Parse()

	client := &fetcher{
		http:      &http.Client{Timeout: *timeout},
		userAgent: *userAgent,
		retries:   *retries,
		delay:     *delay,
	}

	brands, err := client.brands(*dump)
	if err != nil {
		log.Fatalf("не удалось получить список марок: %v", err)
	}
	log.Printf("марок найдено: %d", len(brands))

	// Источник живёт своей жизнью: вёрстка меняется, страница может отдаться
	// урезанной. Молча записать вместо двухсот марок двадцать — худшее, что
	// может сделать синхронизация, поэтому резкое сокращение считаем отказом,
	// а не результатом.
	if previous, err := readExisting(*out); err == nil && !*force {
		if len(brands)*5 < len(previous.Brands)*4 {
			log.Fatalf(
				"разбор дал %d марок против %d в %s — похоже, изменилась страница каталога.\n"+
					"Файл не тронут. Посмотрите разметку (-dump dump.html), а если сокращение настоящее — повторите с -force.",
				len(brands), len(previous.Brands), *out)
		}
	}
	if *limit > 0 && *limit < len(brands) {
		brands = brands[:*limit]
		log.Printf("ограничение -limit: обрабатываем %d", len(brands))
	}

	var failed []string
	for i := range brands {
		models, err := client.models(brands[i].Slug)
		if err != nil {
			// Одна упавшая марка не должна ронять весь прогон: помечаем и идём
			// дальше, чтобы не потерять уже собранное.
			log.Printf("марка %s: %v", brands[i].Slug, err)
			failed = append(failed, brands[i].Slug)
			continue
		}
		brands[i].Models = models
		log.Printf("[%d/%d] %s: %d моделей", i+1, len(brands), brands[i].Name, len(models))
	}

	catalog := vehicleCatalog{
		Source:      catalogURL,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Brands:      brands,
	}
	sortCatalog(&catalog)
	catalog.Version = version(catalog.Brands)

	if err := write(*out, catalog); err != nil {
		log.Fatalf("не удалось записать %s: %v", *out, err)
	}

	total := 0
	for _, brand := range catalog.Brands {
		total += len(brand.Models)
	}
	log.Printf("записано %s: версия %s, марок %d, моделей %d", *out, catalog.Version, len(catalog.Brands), total)
	if len(failed) > 0 {
		log.Printf("не удалось обойти марки: %s", strings.Join(failed, ", "))
		os.Exit(1)
	}
}

type fetcher struct {
	http      *http.Client
	userAgent string
	retries   int
	delay     time.Duration
}

func (f *fetcher) document(pageURL string) (*goquery.Document, error) {
	var lastErr error
	for attempt := 0; attempt <= f.retries; attempt++ {
		if attempt > 0 {
			time.Sleep(f.delay * time.Duration(1<<uint(attempt-1)))
		}
		request, err := http.NewRequest(http.MethodGet, pageURL, nil)
		if err != nil {
			return nil, err
		}
		request.Header.Set("User-Agent", f.userAgent)
		request.Header.Set("Accept-Language", "ru,en;q=0.8")

		response, err := f.http.Do(request)
		if err != nil {
			lastErr = err
			continue
		}
		if response.StatusCode != http.StatusOK {
			response.Body.Close()
			lastErr = fmt.Errorf("%s: %s", pageURL, response.Status)
			continue
		}
		document, err := goquery.NewDocumentFromReader(response.Body)
		response.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		return document, nil
	}
	return nil, lastErr
}

// brands читает индекс каталога. Разметку не трогаем: берём любые ссылки вида
// /catalog/<slug>/ — так парсер переживает редизайн, пока живы сами адреса.
func (f *fetcher) brands(dumpPath string) ([]vehicleBrand, error) {
	document, err := f.document(catalogURL)
	if err != nil {
		return nil, err
	}

	if dumpPath != "" {
		if html, err := document.Html(); err == nil {
			if err := os.WriteFile(dumpPath, []byte(html), 0o644); err != nil {
				log.Printf("не удалось сохранить дамп: %v", err)
			} else {
				log.Printf("HTML индекса сохранён в %s", dumpPath)
			}
		}
	}

	candidates := 0
	seen := make(map[string]vehicleBrand)
	seenNames := make(map[string]struct{})
	document.Find(`a[href*="/catalog/"]`).Each(func(_ int, selection *goquery.Selection) {
		candidates++
		href, ok := selection.Attr("href")
		if !ok {
			return
		}
		slug, ok := brandSlug(href)
		if !ok {
			return
		}
		name := strings.TrimSpace(selection.Text())
		if name == "" {
			return
		}
		// Одна и та же марка встречается и в популярных, и в полном списке.
		// Дубли убираем по имени, а не по slug: клиенты показывают именно имя,
		// и два одинаковых пункта в списке бессмысленны.
		if _, exists := seenNames[sortKey(name)]; exists {
			return
		}
		if _, exists := seen[slug]; !exists {
			seen[slug] = vehicleBrand{Name: name, Slug: slug, Models: []vehicleModel{}}
			seenNames[sortKey(name)] = struct{}{}
		}
	})

	// Разница между числом ссылок на странице и числом принятых марок сразу
	// показывает, что случилось: страница отдалась короткой или фильтр слишком строгий.
	log.Printf("ссылок на /catalog/ на странице: %d, из них принято марок: %d", candidates, len(seen))

	if len(seen) == 0 {
		return nil, fmt.Errorf("на %s не нашлось ни одной ссылки на марку", catalogURL)
	}

	brands := make([]vehicleBrand, 0, len(seen))
	for _, brand := range seen {
		brands = append(brands, brand)
	}
	return brands, nil
}

func (f *fetcher) models(brandSlugValue string) ([]vehicleModel, error) {
	time.Sleep(f.delay)

	pageURL := catalogURL + brandSlugValue + "/"
	document, err := f.document(pageURL)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]vehicleModel)
	seenNames := make(map[string]struct{})
	prefix := "/catalog/" + brandSlugValue + "/"
	document.Find(`a[href*="` + prefix + `"]`).Each(func(_ int, selection *goquery.Selection) {
		href, ok := selection.Attr("href")
		if !ok {
			return
		}
		slug, ok := modelSlug(href, brandSlugValue)
		if !ok {
			return
		}
		name := strings.TrimSpace(selection.Text())
		if name == "" {
			return
		}
		// См. дедупликацию марок: у Drom hilux и hilux_pick_up — оба «Hilux».
		if _, exists := seenNames[sortKey(name)]; exists {
			return
		}
		if _, exists := seen[slug]; !exists {
			seen[slug] = vehicleModel{Name: name, Slug: slug}
			seenNames[sortKey(name)] = struct{}{}
		}
	})

	models := make([]vehicleModel, 0, len(seen))
	for _, model := range seen {
		models = append(models, model)
	}
	sort.Slice(models, func(i, j int) bool { return sortKey(models[i].Name) < sortKey(models[j].Name) })
	return models, nil
}

func pathOf(href string) (string, bool) {
	parsed, err := url.Parse(href)
	if err != nil {
		return "", false
	}
	if parsed.Host != "" && !strings.HasSuffix(parsed.Host, "drom.ru") {
		return "", false
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", false
	}
	return parsed.Path, true
}

func brandSlug(href string) (string, bool) {
	path, ok := pathOf(href)
	if !ok {
		return "", false
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[0] != "catalog" {
		return "", false
	}
	slug := parts[1]
	if notABrand[slug] || !slugPattern.MatchString(slug) {
		return "", false
	}
	return slug, true
}

func modelSlug(href, brand string) (string, bool) {
	path, ok := pathOf(href)
	if !ok {
		return "", false
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 || parts[0] != "catalog" || parts[1] != brand {
		return "", false
	}
	slug := parts[2]
	if !slugPattern.MatchString(slug) {
		return "", false
	}
	return slug, true
}

func sortKey(name string) string { return strings.ToLower(strings.TrimSpace(name)) }

func sortCatalog(catalog *vehicleCatalog) {
	sort.Slice(catalog.Brands, func(i, j int) bool {
		return sortKey(catalog.Brands[i].Name) < sortKey(catalog.Brands[j].Name)
	})
	for i := range catalog.Brands {
		models := catalog.Brands[i].Models
		sort.Slice(models, func(a, b int) bool { return sortKey(models[a].Name) < sortKey(models[b].Name) })
	}
}

// version — дата плюс отпечаток содержимого: пока данные не поменялись, версия
// и ETag стабильны, и клиенты не перекачивают справочник после каждого прогона.
func version(brands []vehicleBrand) string {
	payload, err := json.Marshal(brands)
	if err != nil {
		payload = []byte(fmt.Sprint(len(brands)))
	}
	sum := sha256.Sum256(payload)
	return time.Now().UTC().Format("2006-01-02") + "." + hex.EncodeToString(sum[:4])
}

func write(path string, catalog vehicleCatalog) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(payload, '\n'), 0o644)
}

// readExisting читает уже записанный справочник, чтобы было с чем сравнить
// результат разбора.
func readExisting(path string) (vehicleCatalog, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return vehicleCatalog{}, err
	}
	var catalog vehicleCatalog
	if err := json.Unmarshal(payload, &catalog); err != nil {
		return vehicleCatalog{}, err
	}
	return catalog, nil
}
