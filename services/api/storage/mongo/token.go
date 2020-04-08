package mongo

import (
	"github.com/boodyvo/jogging-api/services/api/storage"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const tokenCollection = "tokens"

func (d *database) SaveToken(token *storage.Token) error {
	if err := d.session.DB(d.name).C(tokenCollection).Insert(token); err != nil {
		return err
	}

	return nil
}

func (d *database) DeleteToken(token *storage.Token) error {
	if err := d.session.DB(d.name).C(tokenCollection).
		Remove(bson.M{"refresh_token": token.Refresh}); err != nil {
		if err == mgo.ErrNotFound {
			return nil
		}
		return err
	}

	return nil
}

func (d *database) GetToken(refreshToken string) (*storage.Token, error) {
	var token storage.Token
	if err := d.session.DB(d.name).C(tokenCollection).
		Find(bson.M{"refresh_token": refreshToken}).One(&token); err != nil {
		if err == mgo.ErrNotFound {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	return &token, nil
}
