package sqlite

var migrate = []string{
	`
	CREATE TABLE IF NOT EXISTS news(
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
	`,
	`
	CREATE TABLE IF NOT EXISTS pictures(
		news_id INTEGER,
		url TEXT,
		caption TEXT
	);
	`,
}

var drop = []string{
	`DROP TABLE IF EXISTS news;`,
	`DROP TABLE IF EXISTS pictures;`,
}
