package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/nrbernard/gator/internal/database"
)

func main() {
	dbPath := os.Getenv("DATABASE_PATH")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("Failed to connect to database: %s\n", err)
		os.Exit(1)
	}
	defer db.Close()

	dbQueries := database.New(db)

	// Get all users
	users, err := dbQueries.GetUsers(context.Background())
	if err != nil {
		fmt.Printf("Failed to get users: %s\n", err)
		os.Exit(1)
	}

	for _, user := range users {
		// Get all posts for this user's followed feeds
		posts, err := dbQueries.GetPostsByUser(context.Background(), database.GetPostsByUserParams{
			UserID: user.ID,
			Limit:  1000, // Get a large number of posts to ensure we get all of them
		})
		if err != nil {
			fmt.Printf("Failed to get posts for user %s: %s\n", user.Name, err)
			continue
		}

		if len(posts) <= 1 {
			fmt.Printf("User %s has 1 or fewer posts, skipping\n", user.Name)
			continue
		}

		// Mark all posts as read except the most recent one
		for i := 1; i < len(posts); i++ {
			post := posts[i]
			err := dbQueries.SaveReadPost(context.Background(), database.SaveReadPostParams{
				ID:     uuid.New().String(),
				PostID: post.ID,
				UserID: user.ID,
			})
			if err != nil {
				fmt.Printf("Failed to mark post %s as read for user %s: %s\n", post.ID, user.Name, err)
				continue
			}
		}

		fmt.Printf("Marked %d posts as read for user %s\n", len(posts)-1, user.Name)
	}
}
