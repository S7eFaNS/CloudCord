package graphdb

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var driver neo4j.DriverWithContext

func Connect() error {
	var err error
	driver, err = neo4j.NewDriverWithContext(
		os.Getenv("NEO4J_URI"),
		neo4j.BasicAuth(os.Getenv("NEO4J_USERNAME"), os.Getenv("NEO4J_PASSWORD"), ""),
	)
	if err != nil {
		return err
	}
	return driver.VerifyConnectivity(context.Background())
}

func Close() {
	driver.Close(context.Background())
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
