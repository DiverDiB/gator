#gator
Gator is a command-line blog aggregator that allows you to register users, subscribe to RSS feeds, and aggregate posts into a local database for easy reading.

Prerequisites
Before installing and running Gator, ensure you have the following installed on your system:

Go: Version 1.22 or higher.

PostgreSQL: A running instance to store users, feeds, and posts.

Installation
Since Go programs are statically compiled, you can install the gator binary directly to your $GOPATH/bin using the go install command:

Bash
go install github.com/your-username/gator@latest
(Note: Replace your-username with your actual GitHub username.)

Setup & Configuration
Gator requires a configuration file located at ~/.gatorconfig.json to manage your database connection and current user session.

Create the config file in your home directory:

JSON
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
Initialize the Database: Ensure you have created a database named gator in your PostgreSQL instance.

Usage
Once installed and configured, you can run gator from anywhere in your terminal. Here are some primary commands to get started:

User Management
Register a new user:

Bash
gator register <username>
Login as an existing user:

Bash
gator login <username>
Feed Management
Add a new RSS feed:

Bash
gator addfeed <name> <url>
List all feeds:

Bash
gator feeds
Aggregation
Start the aggregator:

Bash
gator aggregate 1m
(This will fetch new posts from all feeds every 1 minute.)
