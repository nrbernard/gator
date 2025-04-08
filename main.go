package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/nrbernard/gator/internal/config"
	"github.com/nrbernard/gator/internal/database"
	"github.com/nrbernard/gator/internal/rss"
)

type state struct {
	db     *database.Queries
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commandMap map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandMap[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	if f, ok := c.commandMap[cmd.name]; ok {
		return f(s, cmd)
	}
	return fmt.Errorf("unknown command: %s", cmd.name)
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.config.GetUser())
		if err != nil {
			return fmt.Errorf("failed to get user: %s", err)
		}

		return handler(s, cmd, user)
	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("username required")
	}

	username := cmd.args[0]

	if _, err := s.db.GetUser(context.Background(), username); err != nil {
		return fmt.Errorf("user not found: %s", username)
	}

	if err := s.config.SetUser(username); err != nil {
		return fmt.Errorf("failed to set user: %s", err)
	}

	fmt.Println("User set to:", username)
	return nil
}

func handlerReset(s *state, cmd command) error {
	if err := s.db.DeleteUsers(context.Background()); err != nil {
		return fmt.Errorf("failed to delete users: %s", err)
	}

	fmt.Println("Users deleted")

	return nil
}

func handlerListUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get users: %s", err)
	}

	currentUser := s.config.GetUser()
	for _, user := range users {
		suffix := ""
		if user.Name == currentUser {
			suffix = " (current)"
		}
		fmt.Println(user.Name + suffix)
	}

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("username required")
	}

	username := cmd.args[0]

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:   uuid.New(),
		Name: username,
	})
	if err != nil {
		fmt.Printf("failed to create user: %s\n", err)
		os.Exit(1)
	}

	s.config.SetUser(username)

	fmt.Printf("User created: %+v\n", user)
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("fetch interval required")
	}

	fetchInterval, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("failed to parse time between scrapes: %s", err)
	}

	fmt.Printf("Collecting feeds every %s\n", fetchInterval)

	ticker := time.NewTicker(fetchInterval)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("feed URL and name required")
	}

	feedName := cmd.args[0]
	feedURL := cmd.args[1]

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:     uuid.New(),
		Name:   feedName,
		Url:    feedURL,
		UserID: user.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to create feed: %s", err)
	}

	if _, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:     uuid.New(),
		UserID: user.ID,
		FeedID: feed.ID,
	}); err != nil {
		return fmt.Errorf("failed to create feed follow: %s", err)
	}

	fmt.Printf("Feed created: %+v\n", feed)
	return nil
}

func handlerListFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get feeds: %s", err)
	}

	for _, feed := range feeds {
		fmt.Printf("%s (%s) - %s\n", feed.Name, feed.Url, feed.UserName)
	}

	return nil
}

func handlerFollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("feed URL required")
	}

	feedURL := cmd.args[0]

	feed, err := s.db.GetFeedByUrl(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("failed to get feed: %s", err)
	}

	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to get feed follows: %s", err)
	}

	for _, follow := range follows {
		if follow.FeedID == feed.ID {
			return fmt.Errorf("you are already following this feed")
		}
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:     uuid.New(),
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to create feed follow: %s", err)
	}

	fmt.Printf("%s followed %s\n", feedFollow.UserName, feedFollow.FeedName)
	return nil
}

func handlerListFollowedFeeds(s *state, cmd command, user database.User) error {
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to get feeds: %s", err)
	}

	for _, feed := range feeds {
		fmt.Printf("%s\n", feed.FeedName)
	}

	return nil
}

func handlerUnfollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("feed URL required")
	}

	feedURL := cmd.args[0]

	if err := s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		Url:    feedURL,
	}); err != nil {
		return fmt.Errorf("failed to delete feed follow: %s", err)
	}

	fmt.Printf("%s unfollowed %s\n", user.Name, feedURL)
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	postCount := 2
	if len(cmd.args) > 0 {
		count, err := strconv.Atoi(cmd.args[0])
		if err == nil {
			postCount = count
		}
	}

	fmt.Printf("id: %s\n", user.ID)

	posts, err := s.db.GetPostsByUser(context.Background(), database.GetPostsByUserParams{
		UserID: user.ID,
		Limit:  int32(postCount),
	})
	if err != nil {
		return fmt.Errorf("failed to get posts: %s", err)
	}

	for _, post := range posts {
		fmt.Printf("%s\n", post.Title)
		fmt.Printf("%s\n\n", post.Url)
	}

	return nil
}

func parseDate(date string) time.Time {
	// Mon, 01 Jan 0001 00:00:00 +0000
	fmt.Printf("parsing date: %s\n", date)

	parsed, err := time.Parse(time.RFC1123Z, date)
	if err != nil {
		fmt.Printf("failed to parse date: %s\n", err)
		return time.Time{}
	}

	return parsed
}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get feeds: %s", err)
	}

	if err := s.db.MarkFeedAsFetched(context.Background(), feed.ID); err != nil {
		return fmt.Errorf("failed to mark feed as fetched: %s", err)
	}

	feedData, err := rss.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %s", err)
	}

	for _, item := range feedData.Channel.Item {
		if _, err := s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			PublishedAt: parseDate(item.PubDate),
			FeedID:      feed.ID,
		}); err != nil {
			fmt.Printf("failed to create post: %s\n", err)
		}
	}

	return nil
}

func main() {
	configFile, err := config.Read()
	if err != nil {
		fmt.Printf("Failed to read config: %s\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", configFile.DBUrl)
	if err != nil {
		fmt.Printf("Failed to connect to database: %s\n", err)
		os.Exit(1)
	}

	dbQueries := database.New(db)

	appState := &state{
		db:     dbQueries,
		config: configFile,
	}

	commands := &commands{
		commandMap: make(map[string]func(*state, command) error),
	}

	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerListUsers)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commands.register("feeds", handlerListFeeds)
	commands.register("follow", middlewareLoggedIn(handlerFollowFeed))
	commands.register("following", middlewareLoggedIn(handlerListFollowedFeeds))
	commands.register("unfollow", middlewareLoggedIn(handlerUnfollowFeed))
	commands.register("browse", middlewareLoggedIn(handlerBrowse))

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Please specify a command")
		os.Exit(1)
	}

	command := command{
		name: args[0],
		args: args[1:],
	}

	if err := commands.run(appState, command); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}
