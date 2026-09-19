package main

import (
	"encoding/xml"
	"testing"
	"time"
)

func parse(t *testing.T, data []byte) Offers {
	t.Helper()
	var offers Offers
	if err := xml.Unmarshal(data, &offers); err != nil {
		t.Fatalf("прайс-лист не разбирается как XML: %v", err)
	}
	return offers
}

// Запчасть, помеченную на удаление, площадка видеть не должна: её вот-вот не
// станет, а заказ по ней уже не выполнить.
func TestGenerateXMLPriceListSkipsPartsMarkedForDeletion(t *testing.T) {
	deleteAt := time.Now().Add(24 * time.Hour)
	parts := []Part{
		{Name: "Живая", Status: true},
		{Name: "Удаляется", Status: true, ToDeleteAt: &deleteAt},
	}

	data, err := GenerateXMLPriceList(parts, nil)
	if err != nil {
		t.Fatalf("GenerateXMLPriceList: %v", err)
	}

	offers := parse(t, data)
	if len(offers.Offers) != 1 {
		t.Fatalf("ожидалось 1 предложение, получено %d", len(offers.Offers))
	}
	if offers.Offers[0].Name != "Живая" {
		t.Fatalf("в выгрузку попала не та запчасть: %q", offers.Offers[0].Name)
	}
}

// ИНН берётся из справочника auth-service по идентификатору продавца.
// Своей копии реквизитов у сервиса выгрузки нет — и заводить её не нужно.
func TestGenerateXMLPriceListFillsSupplierINN(t *testing.T) {
	parts := []Part{
		{Name: "С реквизитами", SellerID: 7, Status: true},
		{Name: "Без реквизитов", SellerID: 9, Status: true},
	}

	data, err := GenerateXMLPriceList(parts, map[int64]string{7: "700123456789"})
	if err != nil {
		t.Fatalf("GenerateXMLPriceList: %v", err)
	}

	offers := parse(t, data)
	if got := offers.Offers[0].SupplierInn; got != "700123456789" {
		t.Fatalf("ИНН продавца 7: ожидался 700123456789, получен %q", got)
	}
	// Неизвестный продавец не должен ронять выгрузку или подставлять чужой ИНН.
	if got := offers.Offers[1].SupplierInn; got != "" {
		t.Fatalf("ИНН продавца 9 неизвестен, но подставлен %q", got)
	}
}

// Пустое состояние площадка понимает как «новая», а это не так:
// на складе почти всё бывшее в употреблении.
func TestGenerateXMLPriceListDefaultsCondition(t *testing.T) {
	parts := []Part{
		{Name: "Без состояния", Status: true},
		{Name: "С состоянием", Status: true, Condition: "Новая"},
	}

	offers := parse(t, mustGenerate(t, parts))
	if got := offers.Offers[0].Condition; got != "Б/у" {
		t.Fatalf("ожидалось состояние по умолчанию Б/у, получено %q", got)
	}
	if got := offers.Offers[1].Condition; got != "Новая" {
		t.Fatalf("указанное состояние затёрто: %q", got)
	}
}

// Фотографии лежат относительными путями, а Drom ходит за ними снаружи.
func TestGenerateXMLPriceListMakesAbsolutePictureURL(t *testing.T) {
	parts := []Part{
		{Name: "С фото", Status: true, Photos: []string{"/uploads/12_1700000000.jpg", "/uploads/12_1700000001.jpg"}},
		{Name: "Без фото", Status: true},
	}

	offers := parse(t, mustGenerate(t, parts))
	want := "https://avtoplaneta70.ru/uploads/12_1700000000.jpg"
	if got := offers.Offers[0].Picture; got != want {
		t.Fatalf("ожидался адрес %q, получен %q", want, got)
	}
	if got := offers.Offers[1].Picture; got != "" {
		t.Fatalf("у запчасти без фото появился адрес %q", got)
	}
}

func mustGenerate(t *testing.T, parts []Part) []byte {
	t.Helper()
	data, err := GenerateXMLPriceList(parts, nil)
	if err != nil {
		t.Fatalf("GenerateXMLPriceList: %v", err)
	}
	return data
}
