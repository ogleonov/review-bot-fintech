package main

import (
	"encoding/json"
	"log"
	"net/url"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
)

func main() {
	host := "http://mattermost-feature-sandbox.mattermost-ingress-controller.mattermost.k8s.dev-el" // Замените на ваш хост
	token := "k5e44t9iipn9dfhdhatp4kui1a"

	// Создаем клиента API
	client := model.NewAPIv4Client(host)
	client.SetToken(token)

	// Проверяем аутентификацию
	botUser, _, err := client.GetUser("me", "")
	if err != nil {
		log.Fatalf("Ошибка аутентификации: %v", err)
	}
	log.Printf("Бот авторизован как %s (@%s)", botUser.GetFullName(), botUser.Username)

	// Правильно формируем URL для WebSocket
	parsedURL, err := url.Parse(host)
	if err != nil {
		log.Fatalf("Ошибка парсинга URL: %v", err)
	}

	// Определяем протокол для WebSocket
	wsScheme := "wss"
	if parsedURL.Scheme == "http" {
		wsScheme = "ws"
	}

	// Формируем правильный WebSocket URL
	wsURL := wsScheme + "://" + parsedURL.Host + "/api/v4/websocket"
	log.Printf("Подключаемся к WebSocket: %s", wsURL)

	// Создаем WebSocket клиента
	wsClient, err := model.NewWebSocketClient(wsURL, token)
	if err != nil {
		log.Fatalf("Ошибка создания WebSocket клиента: %v", err)
	}
	defer wsClient.Close()

	wsClient.Listen()
	log.Println("Слушаем события WebSocket...")

	// Основной цикл обработки событий
	for event := range wsClient.EventChannel {
		if event.EventType() != model.WebsocketEventPosted {
			continue
		}

		// Парсим данные события
		postData := event.GetData()["post"].(string)
		var post model.Post
		if err := json.Unmarshal([]byte(postData), &post); err != nil {
			log.Printf("Ошибка парсинга поста: %v", err)
			continue
		}

		// Пропускаем сообщения от самого бота
		if post.UserId == botUser.Id {
			continue
		}

		log.Printf("Получено сообщение: %s", post.Message)

		// Ответ на "ping"
		if strings.Contains(strings.ToLower(post.Message), "ping") {
			reply := &model.Post{
				ChannelId: post.ChannelId,
				Message:   "Pong!",
				RootId:    post.RootId,
			}

			if _, _, err := client.CreatePost(reply); err != nil {
				log.Printf("Ошибка отправки: %v", err)
			} else {
				log.Printf("Ответ отправлен")
			}
		}
	}
}
