package graphdb

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var driver neo4j.DriverWithContext

func InitNeo4j() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading .env")
	}

	uri := os.Getenv("NEO4J_URI")
	username := os.Getenv("NEO4J_USERNAME")
	password := os.Getenv("NEO4J_PASSWORD")

	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		log.Fatalf("Failed to create Neo4j driver: %v", err)
	}

	err = driver.VerifyConnectivity(context.Background())
	if err != nil {
		log.Fatalf("Neo4j not reachable: %v", err)
	}

	log.Println("✅ Connected to Neo4j!")
}

func CloseNeo4j() {
	if driver != nil {
		driver.Close(context.Background())
	}
}

func CreateFriendship(userID, friendID uint) error {
	ctx := context.Background()

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MERGE (u1:User {id: $userID})
			MERGE (u2:User {id: $friendID})
			MERGE (u1)-[:FRIEND]->(u2)
			MERGE (u2)-[:FRIEND]->(u1)
		`
		params := map[string]interface{}{
			"userID":   strconv.Itoa(int(userID)),
			"friendID": strconv.Itoa(int(friendID)),
		}
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})
	if err != nil {
		return fmt.Errorf("Neo4j CreateFriendship error: %w", err)
	}
	return nil
}
