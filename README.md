# gator
RSS feed aggregator and browser 

## Prerequisites

To run `gator`, you'll need:
- PostgreSQL installed and running
- Go 1.19 or later

## Installation

Install the `gator` CLI by running:

```bash
go install github.com/nrbernard/gator@latest
```

## Configuration

Before using `gator`, you need to set up your configuration file. Create a `.gatorconfig.json` file in your home directory with the following structure:

```json
{
    "db_url": "postgres://username:password@localhost:5432/gator",
    "current_user_name": ""
}
```

Replace the `db_url` with your PostgreSQL connection string. The `current_user_name` will be set automatically when you register or login.

## Database Setup

1. Create a PostgreSQL database named `gator`:
```bash
createdb gator
```

2. Run the database migrations:
```bash
goose -dir sql/schema postgres "postgres://username:password@localhost:5432/gator" up
```

## Usage

### User Management

1. Register a new user:
```bash
gator register <username>
```

2. Login as a user:
```bash
gator login <username>
```

3. List all users:
```bash
gator users
```

### Feed Management

1. Add a new feed:
```bash
gator addfeed <feed_name> <feed_url>
```

2. List all available feeds:
```bash
gator feeds
```

3. Follow a feed:
```bash
gator follow <feed_url>
```

4. List your followed feeds:
```bash
gator following
```

5. Unfollow a feed:
```bash
gator unfollow <feed_url>
```

### Reading Feeds

1. Browse your feed items:
```bash
gator browse [count]
```
The optional `count` parameter specifies how many items to show (default is 2).

### Feed Aggregation

To start the feed aggregator that periodically fetches new items:
```bash
gator agg <interval>
```
The `interval` parameter specifies how often to fetch feeds (e.g., "1h" for every hour, "30m" for every 30 minutes).
