DROP TABLE IF EXISTS news;

CREATE TABLE news(
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
);

DROP TABLE IF EXISTS pictures;

CREATE TABLE pictures(
    id bigint not null auto_increment,
    news_id bigint not null references news(gen_id),
    url varchar(255),
    caption varchar(255),

    primary key (id),
    CONSTRAINT pictures_news_id_foreign FOREIGN KEY (news_id) REFERENCES news(gen_id) ON DELETE CASCADE
);