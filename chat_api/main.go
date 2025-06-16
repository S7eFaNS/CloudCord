package main

import (
	"cloudcord/chat_api/db"
	"cloudcord/chat_api/logic"
	"cloudcord/chat_api/mq"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func handleOK(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *statusResponseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func metricsMiddleware(path string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		srw := &statusResponseWriter{ResponseWriter: w, status: 200}

		next.ServeHTTP(srw, r)

		duration := time.Since(start).Seconds()
		method := r.Method
		status := fmt.Sprintf("%d", srw.status)

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)
	})
}

type createChatRequest struct {
	User1   string `json:"user1"`
	User2   string `json:"user2"`
	Content string `json:"content"`
}

type sendMessageRequest struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Content  string `json:"content"`
}

func sendMessageHandler(chatLogic *logic.ChatService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var req sendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err := chatLogic.SendMessageToUser(ctx, req.Sender, req.Receiver, req.Content)
		if err != nil {
			http.Error(w, "Failed to send message: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"message sent"}`))
	}
}

// get chat by two users
func getChatHandler(chatLogic *logic.ChatService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
			return
		}

		user1 := r.URL.Query().Get("user1")
		user2 := r.URL.Query().Get("user2")
		log.Printf("📥 Received query: user1=%q, user2=%q", user1, user2)

		if user1 == "" || user2 == "" {
			http.Error(w, "Missing user1 or user2 query parameters", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		chat, err := chatLogic.GetChatByUsers(ctx, user1, user2)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				chat, err = chatLogic.CreateChat(ctx, user1, user2)
				if err != nil {
					http.Error(w, "Failed to create chat: "+err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				http.Error(w, "Error retrieving chat: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chat)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin == "http://localhost:3000" || origin == "https://cloudcord.com" || origin == "https://cloudcord.info" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	user := os.Getenv("MONGODB_USER")
	pass := os.Getenv("MONGODB_PASS")

	if user == "" || pass == "" {
		log.Fatal("MongoDB credentials are not set in environment variables")
	}

	uri := fmt.Sprintf(
		"mongodb+srv://%s:%s@messages.vbkzymr.mongodb.net/?retryWrites=true&w=majority&appName=Messages",
		user,
		pass,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Could not connect to MongoDB:", err)
	}

	log.Println("✅ Successfully connected to MongoDB Atlas")

	mongoDB := client.Database("Messages")

	chatRepo := db.NewChatRepository(mongoDB)

	var publisher *mq.Publisher
	var err2 error
	rabbitURI := os.Getenv("RABBITMQ_URI")
	if rabbitURI == "" {
		log.Fatal("RabbitMQ path not set in environment")
	}

	for i := 0; i < 8; i++ {
		publisher, err2 = mq.NewPublisher(rabbitURI, "message_notifications")
		if err2 == nil {
			log.Println("✅ RabbitMQ publisher set up successfully")
			break
		}
		log.Printf("Attempt %d: Failed to set up RabbitMQ publisher: %v", i+1, err2)
		time.Sleep(3 * time.Second)
	}

	if err2 != nil {
		log.Fatalf("Failed to set up RabbitMQ publisher after retries: %v", err2)
	}

	go func() {
		maxRetries := 8
		for i := 0; i < maxRetries; i++ {
			err := mq.StartUserDeletionConsumer(rabbitURI, "user_deletion", chatRepo)
			if err == nil {
				log.Println("✅ User deletion consumer started successfully, listening on RabbitMQ...")
				return
			}

			log.Printf("Attempt %d: Failed to start user deletion consumer: %v", i+1, err)
			time.Sleep(3 * time.Second)
		}

		log.Fatal("❌ Failed to start user deletion consumer after retries")
	}()

	chatService := logic.NewChatService(chatRepo, publisher)

	http.HandleFunc("/message/", handleOK)

	http.Handle("/message/send", metricsMiddleware("/message/send", withCORS(sendMessageHandler(chatService))))
	http.Handle("/message/chat", metricsMiddleware("/message/chat", withCORS(getChatHandler(chatService))))

	go func() {
		fmt.Println("Starting metrics server on :2112...")
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":2112", nil); err != nil {
			fmt.Printf("Metrics server error: %v\n", err)
		}
	}()

	http.ListenAndServe(":8084", nil)
}
