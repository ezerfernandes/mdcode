# Database Schema

## Version 1.0

```sql version="1.0"
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
```

## Version 2.0

```sql version="2.0"
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE
);

CREATE TABLE posts (
    id INTEGER PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    title TEXT NOT NULL,
    body TEXT
);
```

## Version 3.0

```sql version="3.0"
CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email TEXT UNIQUE
);

CREATE TABLE posts (
    id INTEGER PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    title TEXT NOT NULL,
    body TEXT,
    created_at TIMESTAMP NOT NULL
);
```
