package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"github.com/john-cai/urlshort/database"
	log "github.com/sirupsen/logrus"

	"github.com/john-cai/urlshort/models"
)

type Server struct {
	*mux.Router
	database *database.Database
	prefix   string
}

func NewServer() *Server {
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresAddr := os.Getenv("POSTGRES_ADDR")
	postgresDB := os.Getenv("POSTGRES_DB")
	database, err := database.New(
		postgresUser,
		postgresAddr,
		postgresDB,
	)
	if err != nil {
		panic(err)
	}

	s := &Server{
		Router:   mux.NewRouter(),
		database: database,
		prefix:   os.Getenv("PREFIX"),
	}
	s.configureRoutes()

	return s
}

func (s *Server) configureRoutes() {
	s.HandleFunc("/shorten", s.Shorten)
	s.HandleFunc("/links/{short}", s.Redirect)
	s.HandleFunc("/links/{short}/stats", s.Stats)
}

// Redirect takes a shortened link and redirects the http request to the full url
func (s *Server) Redirect(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	link := vars["short"]
	shortUrl, err := s.database.GetByShort(link)
	if err != nil {
		if err != pg.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	http.Redirect(w, r, shortUrl.Original, http.StatusSeeOther)
	go func() {
		if err := s.database.AddHit(link, time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)); err != nil {
			log.Errorf("error when updating hits in database: %v", err)
		}
	}()
	return
}

// Shorten takes a shorten url request and tries to shorten it
func (s *Server) Shorten(w http.ResponseWriter, r *http.Request) {
	var err error
	var url models.URLRequest
	if err = json.NewDecoder(r.Body).Decode(&url); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("could not read request"))
		return
	}

	// validate input
	if len(url.Url) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("url cannot be empty"))
		return
	}
	if !(strings.HasPrefix(url.Url, "http://") || strings.HasPrefix(url.Url, "https://")) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("not a valid url"))
		return
	}
	var shortenedURL *models.URL
	// check to see if this url has already been shortened
	shortenedURL, err = s.database.GetByOriginal(url.Url)
	if err == nil {
		w.WriteHeader(http.StatusBadRequest)
		shortenedURL.Short = fmt.Sprintf("%s/links/%s", s.prefix, database.Encode(shortenedURL.ID))
		if err = json.NewEncoder(w).Encode(shortenedURL); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
	if url.Custom != nil {
		if _, err := database.Decode(*url.Custom); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("not a valid url"))
			return
		}

		if shortenedURL, err = s.database.GetByShort(*url.Custom); err != nil {
			if err != pg.ErrNoRows {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			if shortenedURL.Original == url.Url {
				w.WriteHeader(http.StatusOK)
				if err = json.NewEncoder(w).Encode(shortenedURL); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	var short string
	if url.Custom != nil {
		if err = s.database.InsertCustomURL(url.Url, *url.Custom); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("sorry something went wrong"))
			return
		}
		short = *url.Custom
	} else {
		if short, err = s.database.InsertURL(url.Url); err != nil {
			log.Errorf("error when inserting shortened url into database: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("sorry something went wrong"))
			return
		}
	}

	shortenedURL = &models.URL{Original: url.Url, Short: fmt.Sprintf("%s/links/%s", s.prefix, short)}

	if err = json.NewEncoder(w).Encode(shortenedURL); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("sorry something went wrong"))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Stats returns statistics about the link
func (s *Server) Stats(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	link := vars["short"]
	hits, err := s.database.GetTotalHits(link)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("sorry something went wrong"))
		return
	}
	histogram, err := s.database.GetLast7Days(link, time.Now())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("sorry something went wrong"))
		return
	}
	url, err := s.database.GetByShort(link)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("sorry something went wrong"))
		return
	}
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(&models.Stats{
		Original:  url.Original,
		Short:     fmt.Sprintf("%s/links/%s", s.prefix, database.Encode(url.ID)),
		Total:     hits,
		Histogram: histogram,
		CreatedAt: url.CreatedAt,
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	return
}
