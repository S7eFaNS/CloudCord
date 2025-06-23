package main

import (
	"cloudcord/user_api/db"
	"cloudcord/user_api/graphdb"
	"cloudcord/user_api/handlers"
	"cloudcord/user_api/logic"
	"cloudcord/user_api/middleware"
	"cloudcord/user_api/models"
	"cloudcord/user_api/mq"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func handleOK(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin == "http://localhost:3000" || origin == "https://cloudcord.com" || origin == "https://cloudcord.info" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE,OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	db.Connect()

	err := graphdb.Connect()
	if err != nil {
		log.Fatalf("failed to connect to Neo4j: %v", err)
	}
	defer graphdb.Close()

	repo := db.NewRepository(db.DB)

	middleware.InitMiddleware(repo)

	err = models.MigrateAll(db.DB)
	if err != nil {
		log.Fatal("could not migrate db")
	}
	log.Println("Database migrated successfully")

	var publisher *mq.Publisher
	var err2 error
	rabbitURI := os.Getenv("RABBITMQ_URI")
	if rabbitURI == "" {
		log.Fatal("RabbitMQ path not set in environment")
	}

	for i := 0; i < 10; i++ {
		publisher, err2 = mq.NewPublisher(rabbitURI, "user_deletion")
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

	userLogic := logic.NewUserLogicRabbitMQ(repo, publisher)

	http.HandleFunc("/", handleOK)

	http.Handle("/user/create", withCORS(middleware.ValidateJWT(http.HandlerFunc(handlers.HandleCreateUser))))
	http.Handle("/user/user", middleware.ValidateJWT(http.HandlerFunc(handlers.HandleGetUserByID)))
	http.Handle("/user/auth-user", middleware.ValidateJWT(http.HandlerFunc(handlers.HandleGetUserByAuth0ID)))
	http.Handle("/user/users", withCORS(middleware.ValidateJWT(http.HandlerFunc(handlers.HandleGetAllUsers))))
	http.Handle("/user/delete", withCORS(middleware.ValidateJWT(handlers.HandleDeleteUser(userLogic))))
	http.Handle("/user/add-friend", withCORS(middleware.ValidateJWT(handlers.HandleAddFriend(userLogic))))
	http.Handle("/user/is-friend", withCORS(middleware.ValidateJWT(handlers.HandleAreFriends(userLogic))))
	http.Handle("/user/recommendations", withCORS(middleware.ValidateJWT(handlers.HandleFriendRecommendations(userLogic))))

	go func() {
		fmt.Println("Starting metrics server on :2112...")
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":2112", nil); err != nil {
			fmt.Printf("Metrics server error: %v\n", err)
		}
	}()

	fmt.Println("Starting server on :8081...")

	http.ListenAndServe(":8081", nil)
}
