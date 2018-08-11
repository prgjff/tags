package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type (
	Tag struct {
		ID   int64  `db:"rowid" json:"id"`
		Data string `db:"data" json:"data"`
	}

	Tags []Tag
)

var (
	db *sqlx.DB
)

func init() {
	var err error

	if db, err = sqlx.Open("sqlite3", ":memory:"); err != nil {
		panic(err)
	}

	if _, err = db.Exec(`CREATE TABLE tags (data)`); err != nil {
		panic(err)
	}
}

func New(t *Tag) (interface{}, error) {
	if res, err := db.Exec("INSERT INTO tags (data) VALUES (?)", t.Data); err != nil {
		return t, err
	} else if t.ID, err = res.LastInsertId(); err != nil {
		return t, err
	}
	return t, nil
}

func Read(id int64) (interface{}, error) {
	ts := &Tags{}
	if id > 0 {
		return ts, db.Select(ts, "SELECT rowid, data FROM tags WHERE rowid = ?", id)
	} else {
		return ts, db.Select(ts, "SELECT rowid, data FROM tags")
	}
}

func Update(t *Tag) (interface{}, error) {
	_, err := db.Exec("UPDATE tags SET data = ? WHERE rowid = ?", t.Data, t.ID)
	return nil, err
}

func Delete(id int64) (interface{}, error) {
	_, err := db.Exec("DELETE FROM tags WHERE rowid = ?", id)
	return nil, err
}

type (
	Response struct {
		Data  interface{} `json:"data,omitempty"`
		Error string      `json:"error,omitempty"`
	}
)

func NewResponse(d interface{}, e error) Response {
	if e != nil {
		return Response{Data: d, Error: e.Error()}
	}
	return Response{Data: d}
}

func (r Response) Send(w http.ResponseWriter) {
	j, err := json.Marshal(r)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprint(w, string(j))
}

func handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rcv := recover(); rcv != nil {
			Response{Error: fmt.Sprint(rcv)}.Send(w)
		}
	}()

	t := &Tag{}
	if (r.Method == http.MethodPost) || (r.Method == http.MethodPut) {
		err := json.NewDecoder(r.Body).Decode(t)
		if err != nil {
			NewResponse(nil, err).Send(w)
			return
		}
	}

	q := r.URL.Query()
	id, _ := strconv.ParseInt(q.Get("id"), 10, 64)

	switch r.Method {
	case http.MethodPost:
		NewResponse(New(t)).Send(w)
	case http.MethodGet:
		NewResponse(Read(id)).Send(w)
	case http.MethodPut:
		NewResponse(Update(t)).Send(w)
	case http.MethodDelete:
		NewResponse(Delete(id)).Send(w)
	}
}

func main() {
	http.HandleFunc("/api/v1/tags", handler)
	panic(http.ListenAndServe(":80", nil))
}
