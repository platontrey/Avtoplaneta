package main

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestParseBriefHTML(t *testing.T) {
	snippet := `<div class="dialog-brief "
	 data-dialog-id="1771417825"
	 data-interlocutor="AVKMotors777"
>
	<div class="dialog-brief__image " style="background-image: url(https://static.baza.drom.ru/v/1731923827174_thumbnail120)">
			</div>
	<div class="dialog-brief__accessory">
		<div class="dialog-brief__meta">
						<i class="bzr-dialog__message-check" data-role="message-check"
			   data-state="read" title="Прочитано адресатом">&nbsp;</i>
						<span class="bzr-dialog__message-dt">21.12.25</span>
		</div>
	</div>
	<div class="dialog-brief__body">
		<div class="dialog-brief__latest_msg auto-shy">Извините, что отвлекаю вас, я тестирую свой сайт. Делаю тест чата Drom внутри моего сайта.</div>
					<div class="dialog-brief__interlocutor">AVKMotors777</div>
					</div>
	<div class="clear"></div>
</div>`

	var dialog DromDialog
	parseBriefHTML(snippet, &dialog)

	expectedPhoto := "https://static.baza.drom.ru/v/1731923827174_thumbnail120"
	if dialog.PartPhotoURL != expectedPhoto {
		t.Errorf("expected photo %q, got %q", expectedPhoto, dialog.PartPhotoURL)
	}

	if dialog.LastMessage == nil || *dialog.LastMessage != "Извините, что отвлекаю вас, я тестирую свой сайт. Делаю тест чата Drom внутри моего сайта." {
		t.Errorf("unexpected last message: %v", dialog.LastMessage)
	}

	if dialog.LastMessageAt != "21.12.25" {
		t.Errorf("expected date 21.12.25, got %q", dialog.LastMessageAt)
	}
}

func TestParseDromHeaderHTML(t *testing.T) {
	headerHTML := `<div class="bzr-dialog-header__title">
	<h2 class="bzr-dialog-header__main-title">
<span class="userNick auto-shy ">
	<a href="https://baza.drom.ru/user/AVKMotors777/" rel="nofollow">AVKMotors777</a>
</span>
<nobr>
	<span class="userRating">
		<a rel="nofollow" href="https://baza.drom.ru/user/AVKMotors777/feedbacks?role=participant" class="ratingPositive">+118</a>
	</span>
</nobr>
	<span data-role="last-seen" class="bzr-dialog__user-meta">был(а) 01:43</span>
	</h2>
	<h3 class="bzr-dialog-header__sub-title">
		14 900 р.&nbsp;&bull;&nbsp;
		<a href="https://baza.drom.ru/kaliningrad/sell_spare_parts/shlejf-podrulevoj-toyota-lc-200-ls460-ls600-lx460-570-10-original-100-124415867.html">Шлейф подрулевой Toyota LC 200/LS460/LS600/LX460/570 10- Оригинал 100%</a>
	</h3>
</div>`

	var dialog DromDialog
	parseDromHeaderHTML(headerHTML, &dialog)

	if dialog.Interlocutor != "AVKMotors777" {
		t.Errorf("expected interlocutor AVKMotors777, got %q", dialog.Interlocutor)
	}
	if dialog.UserRating != "+118" {
		t.Errorf("expected user rating +118, got %q", dialog.UserRating)
	}
	if dialog.UserStatus != "был(а) 01:43" {
		t.Errorf("expected user status 'был(а) 01:43', got %q", dialog.UserStatus)
	}
	if dialog.PartTitle != "Шлейф подрулевой Toyota LC 200/LS460/LS600/LX460/570 10- Оригинал 100%" {
		t.Errorf("expected part title %q, got %q", "Шлейф подрулевой Toyota LC 200/LS460/LS600/LX460/570 10- Оригинал 100%", dialog.PartTitle)
	}
	if dialog.PartPrice != "14 900 р." {
		t.Errorf("expected part price '14 900 р.', got %q", dialog.PartPrice)
	}
	expectedURL := "https://baza.drom.ru/kaliningrad/sell_spare_parts/shlejf-podrulevoj-toyota-lc-200-ls460-ls600-lx460-570-10-original-100-124415867.html"
	if dialog.PartURL != expectedURL {
		t.Errorf("expected part URL %q, got %q", expectedURL, dialog.PartURL)
	}
	if dialog.BulletinID != "124415867" {
		t.Errorf("expected bulletin ID '124415867', got %q", dialog.BulletinID)
	}
}

func TestExtractDromMessagesFromHTML(t *testing.T) {
	dialogHTML := `<div class="bzr-dialog">
	<div class="bzr-dialog__msg-container" data-message-id="msg_1">
		<div class="bzr-dialog__message bzr-dialog__message_out">
			<div class="bzr-dialog__message-body">
				<div class="bzr-dialog__text">Здравствуйте, есть в наличии?</div>
				<div class="bzr-dialog__message-meta">
					<span class="bzr-dialog__message-dt">10:00</span>
					<i class="bzr-dialog__message-check" data-role="message-check" data-state="read">&nbsp;</i>
				</div>
			</div>
		</div>
	</div>
	<div class="bzr-dialog__msg-container" data-message-id="msg_2">
		<div class="bzr-dialog__message bzr-dialog__message_in">
			<div class="bzr-dialog__message-body">
				<div class="bzr-dialog__text">Да, в наличии на складе</div>
				<div class="bzr-dialog__message-meta">
					<span class="bzr-dialog__message-dt">10:05</span>
				</div>
			</div>
		</div>
	</div>
</div>`

	messages := extractDromMessagesFromHTML(dialogHTML, "Продавец")
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}

	if messages[0].MessageID != "msg_1" || messages[0].Author != "Я" || messages[0].Direction != "outgoing" || !messages[0].IsRead {
		t.Errorf("unexpected message 0: %+v", messages[0])
	}
	if messages[1].MessageID != "msg_2" || messages[1].Author != "Продавец" || messages[1].Direction != "incoming" {
		t.Errorf("unexpected message 1: %+v", messages[1])
	}
}

func TestExtractCSRFToken(t *testing.T) {
	htmlSample := `<div><input  name="csrfToken" type="hidden" value="6b15974b5f08024685c798784a70e4a6" id="csrfToken"/></div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlSample))
	if err != nil {
		t.Fatalf("failed to parse html: %v", err)
	}
	val, exists := doc.Find("input[name='csrfToken']").Attr("value")
	if !exists || val != "6b15974b5f08024685c798784a70e4a6" {
		t.Errorf("expected csrf token 6b15974b5f08024685c798784a70e4a6, got %q", val)
	}
}

func TestParseDromLoginErrors(t *testing.T) {
	// 1. Field error
	fieldErrorHTML := `<div class="control" data-name="password">
		<label for="password" class="formError">Пароль</label>
		<div class='comment formError controlErrors' id='password_errors'><div>Данные для входа неверны</div></div>
	</div>`
	doc1, _ := goquery.NewDocumentFromReader(strings.NewReader(fieldErrorHTML))
	errMsg := strings.TrimSpace(doc1.Find("#sign_errors, #password_errors, .formError.controlErrors").First().Text())
	if errMsg != "Данные для входа неверны" {
		t.Errorf("expected 'Данные для входа неверны', got %q", errMsg)
	}

	// 2. Banner error
	bannerHTML := `<div class="notification-banner notification-banner_alert infoExists sign-form-banner">
		<div class="notification-banner__title-l">Пользователь не найден</div>
		<div class="notification-banner__content">Мы не нашли пользователя с такими логином</div>
	</div>`
	doc2, _ := goquery.NewDocumentFromReader(strings.NewReader(bannerHTML))
	title := strings.TrimSpace(doc2.Find(".notification-banner__title-l").First().Text())
	desc := strings.TrimSpace(doc2.Find(".notification-banner__content").First().Text())
	if title != "Пользователь не найден" || desc != "Мы не нашли пользователя с такими логином" {
		t.Errorf("unexpected banner extraction: title=%q, desc=%q", title, desc)
	}
}
