package mongo

import (
	"github.com/boodyvo/jogging-api/services/api/storage"

	"gopkg.in/mgo.v2"
)

type database struct {
	session *mgo.Session
	name    string
}

type Index struct {
	CollectionName string
	Index          []mgo.Index
}

var (
	indexes = []Index{
		{
			CollectionName: "users",
			Index: []mgo.Index{
				{
					Key:    []string{"email"},
					Unique: true,
				},
			},
		},
		{
			CollectionName: "trackings",
			Index: []mgo.Index{
				{
					Key:    []string{"user_id"},
					Unique: false,
				},
				// for pagination
				{
					Key:    []string{"cursor"},
					Unique: true,
				},
				// for pagination
				{
					Key:    []string{"user_id", "cursor"},
					Unique: true,
				},
				// for reports
				{
					Key:    []string{"user_id", "date"},
					Unique: false,
				},
			},
		},
		{
			CollectionName: "tokens",
			Index: []mgo.Index{
				{
					Key:    []string{"refresh_token"},
					Unique: true,
				},
			},
		},
	}
)

func New(url, name string) (storage.Storage, error) {
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}

	for _, index := range indexes {
		collection := session.DB(name).C(index.CollectionName)
		for _, ix := range index.Index {
			if err := collection.EnsureIndex(ix); err != nil {
				return nil, err
			}
		}
	}

	return &database{
		session: session,
		name:    name,
	}, nil
}
