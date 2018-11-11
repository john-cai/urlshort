package database

import (
	"math/rand"
	"testing"
	"time"

	"github.com/john-cai/urlshort/models"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetByOriginal(t *testing.T) {
	db, err := NewTestDB()
	require.NoError(t, err)
	tc := []models.URL{
		models.URL{
			Original: uuid.New(),
		},
		models.URL{
			Original: uuid.New(),
		},
		models.URL{
			Original: uuid.New(),
		},
	}

	for _, test := range tc {
		require.NoError(t, db.db.Insert(&test))
		url, err := db.GetByOriginal(test.Original)
		require.NoError(t, err)
		assert.Equal(t, test.Short, url.Short)
		assert.Equal(t, test.Original, url.Original)
	}
}

func TestGetByShort(t *testing.T) {
	db, err := NewTestDB()
	require.NoError(t, err)
	url := uuid.New()
	short, err := db.InsertURL(url)
	require.NoError(t, err)
	u, err := db.GetByShort(short)
	require.NoError(t, err)
	assert.Equal(t, url, u.Original)
}

func TestAddHits(t *testing.T) {
	db, err := NewTestDB()
	require.NoError(t, err)
	tc := []struct {
		url  string
		hits int
	}{
		{
			url:  "www.google.com",
			hits: rand.Intn(100),
		}, {
			url:  "www.yahoo.com",
			hits: rand.Intn(100),
		}, {
			url:  "www.bing.com",
			hits: rand.Intn(100),
		},
	}

	for _, test := range tc {
		short, err := db.InsertURL(test.url)
		require.NoError(t, err)
		for i := 0; i < test.hits; i++ {
			require.NoError(t, db.AddHit(short, time.Now()))
		}
	}
}
