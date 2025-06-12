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

type Recommendation struct {
	UserID            uint
	MutualFriendCount int
}

func GetFriendRecommendations(userID uint) ([]Recommendation, error) {
	ctx := context.Background()
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	userIDStr := strconv.Itoa(int(userID))
	recommendations := []Recommendation{}

	_, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (me:User {id: $userID})-[:FRIEND]->(mutual:User)-[:FRIEND]->(recommended:User)
			WHERE NOT (me)-[:FRIEND]->(recommended)
			  AND me.id <> recommended.id
			RETURN recommended.id AS userID, count(mutual) AS mutualFriendCount
			ORDER BY mutualFriendCount DESC
			LIMIT 3
		`

		params := map[string]interface{}{
			"userID": userIDStr,
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		for result.Next(ctx) {
			record := result.Record()
			userIDVal, _ := record.Get("userID")
			mutualCountVal, _ := record.Get("mutualFriendCount")

			uidStr, ok := userIDVal.(string)
			if !ok {
				continue
			}
			uidInt, err := strconv.Atoi(uidStr)
			if err != nil {
				continue
			}
			mutualCount, _ := mutualCountVal.(int64)

			recommendations = append(recommendations, Recommendation{
				UserID:            uint(uidInt),
				MutualFriendCount: int(mutualCount),
			})
		}

		return nil, result.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("Neo4j GetFriendRecommendations error: %w", err)
	}
	return recommendations, nil
}
