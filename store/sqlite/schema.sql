DROP TABLE IF EXISTS news;

CREATE TABLE news(
    gen_id TEXT NOT NULL UNIQUE,
    author TEXT,
    datetime TIMESTAMP,
    title TEXT,
    location TEXT,
    content TEXT,
    tags TEXT,
    url TEXT,
    newspaper_name TEXT,
    newspaper_id TEXT,
    newspaper_category TEXT,
    newspaper_subcategory TEXT,
    newspaper_tags TEXT, 
    newspaper_url TEXT
);

DROP TABLE IF EXISTS pictures;

CREATE TABLE pictures(
    news_id INTEGER,
    url TEXT,
    caption TEXT
);