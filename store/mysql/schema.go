package mysql

var migrate = []string{
	`
	CREATE TABLE IF NOT EXISTS news(
		id bigint not null auto_increment,
		gen_id varchar(255) not null,
		author varchar(255),
		datetime timestamp,
		title varchar(255),
		location varchar(255),
		content TEXT,
		tags TEXT,
		url varchar(255),
		newspaper_name varchar(255),
		newspaper_id varchar(255),
		newspaper_category varchar(255),
		newspaper_subcategory varchar(255),
		newspaper_tags TEXT,
		newspaper_url varchar(255),
	
		primary key (id),
		unique (gen_id)
	) default charset = utf8mb4;
	`,
	`
	CREATE TABLE IF NOT EXISTS pictures(
		id bigint not null auto_increment,
		news_id varchar(255) not null references news(gen_id),
		url varchar(255),
		caption varchar(255),
	
		primary key (id),
		CONSTRAINT pictures_news_id_foreign FOREIGN KEY (news_id) REFERENCES news(gen_id) ON DELETE CASCADE
	) default charset = utf8mb4;
	`,
}

var drop = []string{
	`DROP TABLE IF EXISTS news;`,
	`DROP TABLE IF EXISTS pictures;`,
}
