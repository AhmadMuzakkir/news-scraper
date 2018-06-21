package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ahmadmuzakkir/scrapenews/model"
	_ "github.com/go-sql-driver/mysql"
)

//Database encapsulates database
type Store struct {
	db *sql.DB
}

//Begins a transaction
func (s *Store) begin() (tx *sql.Tx) {
	tx, err := s.db.Begin()
	if err != nil {
		log.Println(err)
		return nil
	}
	return tx
}

func (s *Store) prepare(q string) (stmt *sql.Stmt) {
	stmt, err := s.db.Prepare(q)
	if err != nil {
		log.Println(err)
		return nil
	}
	return stmt
}

func (s *Store) query(q string, args ...interface{}) (rows *sql.Rows) {
	rows, err := s.db.Query(q, args...)
	if err != nil {
		log.Println(err)
		return nil
	}
	return rows
}

func NewStore(address, username, password, database string) (*Store, error) {
	connstr := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?parseTime=true",
		username, password, address, database,
	)

	db, err := sql.Open("mysql", connstr)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	store := &Store{db: db}

	err = store.Migrate()
	if err != nil {
		return nil, err
	}

	return store, nil
}

// Migrate migrates the store database.
func (s *Store) Migrate() error {
	for _, q := range migrate {
		_, err := s.db.Exec(q)
		if err != nil {
			return fmt.Errorf("sql exec error: %s; query: %q", err, q)
		}
	}
	return nil
}

func (s *Store) Insert(news []*model.News) error {
	tx := s.begin()

	stmt, err := tx.Prepare("INSERT IGNORE INTO news(gen_id,author,datetime,title,location,content,tags,url," +
		"newspaper_name,newspaper_id,newspaper_category,newspaper_subcategory,newspaper_tags,newspaper_url) " +
		"VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	stmtPicture, err := tx.Prepare("INSERT INTO pictures(news_id, url, caption) VALUES (?,?,?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmtPicture.Close()

	for _, n := range news {
		var tags string
		if n.Tags != nil && len(n.Tags) > 0 {
			for _, t := range n.Tags {
				tags += t
				tags += ","
			}

			tags = strings.TrimSuffix(tags, ",")
		}

		var newspaperTags string
		if n.Source.Tags != nil && len(n.Source.Tags) > 0 {
			for _, t := range n.Source.Tags {
				newspaperTags += t
				newspaperTags += ","
			}

			newspaperTags = strings.TrimSuffix(newspaperTags, ",")
		}

		res, err := stmt.Exec(n.Id, n.Author, n.Datetime, n.Title, n.Location, n.Content, tags, n.Url,
			n.Source.NewspaperName, n.Source.NewspaperId, n.Source.OriginalCategory, n.Source.OriginalSubcategory, newspaperTags, n.Source.Url)
		if err != nil {
			tx.Rollback()
			return err
		}

		id, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			return err
		}

		for _, pic := range n.Pictures {
			stmtPicture.Exec(id, pic.ImageUrl, pic.Caption)
		}

	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAll(from time.Time, until time.Time) ([]*model.News, error) {
	var err error
	var rows *sql.Rows

	stmt, err := s.db.Prepare("SELECT news.id, news.gen_id, news.author, news.datetime, news.title, news.location, news.content, news.tags, news.url, news.newspaper_name, news.newspaper_id, news.newspaper_category, news.newspaper_subcategory, news.newspaper_tags, news.newspaper_url, pictures.url, pictures.caption FROM news WHERE news.datetime >= $1 and news.datetime <= $2 INNER JOIN pictures ON pictures.news_id = news.id ORDER BY news.id")
	if err != nil {
		return nil, err
	}

	if !from.IsZero() && !until.IsZero() {
		rows, err = stmt.Query(from, until)
	} else if !from.IsZero() {
		rows, err = stmt.Query(from, nil)
	} else if !until.IsZero() {
		rows, err = stmt.Query(nil, until)
	} else {
		rows, err = stmt.Query(nil, nil)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []*model.News
	var lastNews *model.News
	var lastNewsPk int64
	for rows.Next() {
		var pk int64
		var tags string
		n := model.News{}
		newsSource := model.NewsSource{}
		var sourceTags string

		picture := model.Picture{}
		rows.Scan(&pk, &n.Id, &n.Author, &n.Datetime, &n.Title, &n.Location, &n.Content, &tags, &n.Url, &newsSource.NewspaperName, &newsSource.NewspaperId, &newsSource.OriginalCategory, &newsSource.OriginalSubcategory, &sourceTags, &newsSource.Url, &picture.ImageUrl, &picture.Caption)

		newsSource.Tags = strings.Split(sourceTags, ",")
		n.Tags = strings.Split(tags, ",")
		n.Source = newsSource

		if lastNewsPk == pk && picture.ImageUrl != "" {
			lastNews.Pictures = append(lastNews.Pictures, &picture)
			continue
		}

		if picture.ImageUrl != "" {
			n.Pictures = append(n.Pictures, &picture)
		}

		list = append(list, &n)
		lastNews = &n
		lastNewsPk = pk
	}

	return list, nil
}

func (s *Store) GetByKeywords(keywords []string) ([]*model.News, error) {
	return nil, nil
}

func (s *Store) GetByProvider(provider string) ([]*model.News, error) {
	return nil, nil
}
