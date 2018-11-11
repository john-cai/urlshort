package database

import (
	"time"

	"github.com/go-pg/pg"

	"github.com/john-cai/urlshort/models"
)

// Database is a top level database object we can tie database methods to
type Database struct {
	db *pg.DB
}

// New creates a new database object
func New(user, addr, database string) (*Database, error) {
	db := pg.Connect(&pg.Options{
		User:     user,
		Addr:     addr,
		Database: database,
	})
	return &Database{
		db: db,
	}, nil
}

func NewTestDB() (*Database, error) {
	return New("postgres", "localhost:5432", "urlshort_test")
}

// GetByOriginal gets the link by the original url
func (d *Database) GetByOriginal(url string) (*models.URL, error) {
	var shortenedURL models.URL
	_, err := d.db.QueryOne(&shortenedURL, `select * from urls where original = ?`, url)
	if err != nil {
		return nil, err
	}
	return &shortenedURL, nil
}

// GetByShort gets the link by the shortened url
func (d *Database) GetByShort(short string) (*models.URL, error) {
	var shortenedURL models.URL

	id, err := Decode(short)
	if err != nil {
		return nil, err
	}
	_, err = d.db.QueryOne(&shortenedURL, `select * from urls where id = ?`, id)
	if err != nil {
		return nil, err
	}
	return &shortenedURL, nil
}

// Insert is a proxy to the pg insert method
func (d *Database) Insert(i interface{}) error {
	return d.db.Insert(i)
}

// InsertURL inserts a url into the database, returning the shortened version
func (d *Database) InsertURL(url string) (string, error) {
	u := models.URL{Original: url, CreatedAt: time.Now()}
	if err := d.db.Insert(&u); err != nil {
		return "", err
	}
	return Encode(u.ID), nil
}

// UpdateURL inserts a url into the database, returning the shortened version
func (d *Database) UpdateURL(url *models.URL) error {
	return d.db.Update(url)
}

// InsertCustomURL inserts a url into the database
func (d *Database) InsertCustomURL(url, custom string) error {
	id, err := Decode(custom)
	if err != nil {
		return err
	}
	u := models.URL{Original: url, ID: id}
	if err := d.db.Insert(&u); err != nil {
		return err
	}
	return nil
}

// AddHit takes a url and increments that day's hits
func (d *Database) AddHit(short string, now time.Time) error {
	return d.db.RunInTransaction(func(tx *pg.Tx) error {
		var stats models.URLStats

		id, err := Decode(short)
		if err != nil {
			return err
		}
		_, err = tx.QueryOne(&stats, `select * from url_stats where url_id = ? and date = ?`, id, now)
		if err != nil {
			if err != pg.ErrNoRows {
				return err
			}
			stats = models.URLStats{
				Hits:  1,
				UrlID: id,
				Date:  now,
			}
			if err = tx.Insert(&stats); err != nil {
				return err
			}
			return nil
		}
		stats.Hits++
		stats.UpdatedAt = time.Now()
		if err = tx.Update(&stats); err != nil {
			return err
		}
		return nil
	})
}

// GetTotalHits gets the total count of the hits for a given short url
func (d *Database) GetTotalHits(short string) (int, error) {
	var hits int
	id, err := Decode(short)
	if err != nil {
		return 0, err
	}
	_, err = d.db.Query(&hits, `select sum(hits) from url_stats where url_id = ?`, id)
	if err != nil {
		return 0, err
	}
	return hits, nil
}

// GetLast7Days returns a list of hits summarized by day for the past 7 days
func (d *Database) GetLast7Days(short string, now time.Time) ([]models.URLStats, error) {
	var stats []models.URLStats
	id, err := Decode(short)
	if err != nil {
		return nil, err
	}
	_, err = d.db.Query(&stats, `select * from url_stats where url_id = ? and date >= ?`, id, now.Add(-7*24*time.Hour))
	if err != nil {
		return nil, err
	}
	return stats, nil
}
