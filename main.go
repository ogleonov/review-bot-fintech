package main

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
)

func main() {
	// Конфигурация
	host := "http://mattermost-feature-sandbox.mattermost-ingress-controller.mattermost.k8s.dev-el" // Замените на ваш хост
	token := "token"

	// Создаем клиента API
	client := model.NewAPIv4Client(host)
	client.SetToken(token)

	// Проверяем аутентификацию
	botUser, _, err := client.GetUser("me", "")
	if err != nil {
		log.Fatalf("Ошибка аутентификации: %v", err)
	}
	log.Printf("Бот авторизован как %s (@%s)", botUser.GetFullName(), botUser.Username)

	// Явно создаем WebSocket URL
	wsURL := "wss://" + strings.TrimPrefix(host, "https://") + "/api/v4/websocket"
	log.Printf("Подключаемся к WebSocket: %s", wsURL)

	// Создаем WebSocket клиента с явным URL
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
