package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"html"

	"github.com/diverdib/gator/internal/config"
	"github.com/diverdib/gator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	registeredCommands map[string]func(*state, command) error
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// register adds a new handler function to the map
func (c *commands) register(name string, f func(*state, command) error) {
	if c.registeredCommands == nil {
		c.registeredCommands = make(map[string]func(*state, command) error)
	}
	c.registeredCommands[name] = f
}

// run executes a command if it exists in the map
func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.registeredCommands[cmd.name]
	if !ok {
		return fmt.Errorf("command %s is not found", cmd.name)
	}
	return handler(s, cmd)
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %v <name>", cmd.name)
	}
	name := cmd.args[0]

	_, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("user %s does not exist", name)
	}

	err = s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("could not set current user: %w", err)
	}

	fmt.Printf("User has been set to: %s\n", name)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %v <name>", cmd.name)
	}

	name := cmd.args[0]

	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
	}

	user, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		return fmt.Errorf("could not create user: %w", err)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("could not set current user: %w", err)
	}

	fmt.Printf("User %s was created successfully!\n", user.Name)
	fmt.Println(user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not reset users: %w", err)
	}
	fmt.Println("All users have been deleted from the database.")
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not get users: %w", err)
	}
	currentUser := s.cfg.CurrentUserName
	for _, user := range users {
		if user.Name == currentUser {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("usage: %s <name> <url>", cmd.name)
	}

	name := cmd.args[0]
	feedURL := cmd.args[1]

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:			uuid.New(),
		CreatedAt:	time.Now().UTC(),
		UpdatedAt:	time.Now().UTC(),
		Name:		name,
		Url:		feedURL,
		UserID:		user.ID,
	})
	if err != nil {
		return fmt.Errorf("could not create feed: %w", err)
	}

	fmt.Println("Feed created successfully:")
	printFeed(feed)
	fmt.Println();
	fmt.Println("=====================================")

	return nil
}

func handlerAgg(s *state, cmd command) error {
	feedURL := "https://www.wagslane.dev/index.xml"
	if len(cmd.args) > 0 {
		feedURL = cmd.args[0]
	}

	_, err := fetchFeed(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("could not fetch feed: %w", err)
	}

	return nil
}

func handlerGetFeed(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("could not get feeds: %w", err)
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds found in the database.")
		return nil
	}

	fmt.Printf("Found %d feeds:\n", len(feeds))
	for _, feed := range feeds {
		fmt.Printf("* Name:			%s\n", feed.Name)
		fmt.Printf("* URL:			 %s\n", feed.Url)
		fmt.Printf("* Created By:	  %s\n", feed.UserName)
		fmt.Println("--------------------")
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.name)
	}

	url := cmd.args[0]

	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("could not find feed with URL %s: %w", url, err)
	}

	ff, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:			uuid.New(),
		CreatedAt:	time.Now(),
		UpdatedAt: 	time.Now(),
		UserID:		user.ID,
		FeedID:		feed.ID,
	})
	if err != nil {
		return fmt.Errorf("could not follow feed: %w", err)
	}
	fmt.Printf("User %s is now following feed: %s\n", ff.UserName, ff.FeedName)
	return nil
}

func printFeed(feed database.Feed) {
	fmt.Printf("* ID: 			 %s\n", feed.ID)
	fmt.Printf("* Created:		 %v\n", feed.CreatedAt)
	fmt.Printf("* Updated:		 %v\n", feed.UpdatedAt)
	fmt.Printf("* Name:			%s\n", feed.Name)
	fmt.Printf("* URL:			 %s\n", feed.Url)
	fmt.Printf("* User ID:		 %s\n", feed.UserID)
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		userName := s.cfg.CurrentUserName
		if userName == "" {
			return fmt.Errorf("no user is currently logged in")
		}

		user, err := s.db.GetUser(context.Background(), userName)
		if err != nil {
			return fmt.Errorf("could not find user: %w", err)
		}
	return handler(s, cmd, user)
	}
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}

	// Set the User-Agent header
	req.Header.Set("User-Agent", "gator")

	// Make the HTTP request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch feed: status code %d", resp.StatusCode)
	}

	// Read the data and Unmarshal into RSSFeed struct
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var feed RSSFeed
	err = xml.Unmarshal(data, &feed)
	if err != nil {
		return nil, err
	}

	// Unescape the top level Channel fields
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Link = html.UnescapeString(feed.Channel.Link)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	// Unescape each Item's fields
	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Link = html.UnescapeString(feed.Channel.Item[i].Link)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
		feed.Channel.Item[i].PubDate = html.UnescapeString(feed.Channel.Item[i].PubDate)
	}
	fmt.Println("Feed fetched successfully:", feed.Channel.Title)
	// Print the entire structure to the console
	fmt.Printf("%+v\n", feed)

	return &feed, nil
}

func main() {

	// Read the config file
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}

	dbQueries := database.New(db)

	// Create the program state
	programState := &state{
		db:  dbQueries,
		cfg: &cfg,
	}

	cmds := commands{
		registeredCommands: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerGetUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerGetFeed)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))

	// Check if enough argumaents were provided
	if len(os.Args) < 2 {
		log.Fatal("Usage: gator <command> [args...]")
	}

	// Build the command struct from os.Args
	cmd := command{
		name: os.Args[1],
		args: os.Args[2:],
	}

	// Run the command
	err = cmds.run(programState, cmd)
	if err != nil {
		log.Fatalf("error running command: %v", err)
	}

}
