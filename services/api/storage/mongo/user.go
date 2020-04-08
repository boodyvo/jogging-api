package mongo

import (
	"github.com/boodyvo/jogging-api/services/api/storage"
	"github.com/boodyvo/jogging-api/services/api/storage/mongo/filterparser"
	"github.com/google/uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	defaultUsersPerRequest = 10
	userCollection         = "users"
)

func (d *database) SaveUser(user *storage.User) error {
	if err := d.session.DB(d.name).C(userCollection).Insert(user); err != nil {
		return err
	}

	return nil
}

func (d *database) UpdateUser(user *storage.User) error {
	if err := d.session.DB(d.name).C(userCollection).UpdateId(user.ID, user); err != nil {
		if err == mgo.ErrNotFound {
			return storage.ErrNotFound
		}
		return err
	}

	return nil
}

func (d *database) DeleteUser(id uuid.UUID) error {
	if err := d.session.DB(d.name).C(userCollection).RemoveId(id); err != nil {
		if err == mgo.ErrNotFound {
			return nil
		}
		return err
	}

	return nil
}

func (d *database) GetUser(id uuid.UUID) (*storage.User, error) {
	var user storage.User
	if err := d.session.DB(d.name).C(userCollection).FindId(id).One(&user); err != nil {
		if err == mgo.ErrNotFound {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (d *database) GetUserByEmail(email string) (*storage.User, error) {
	var user storage.User
	if err := d.session.DB(d.name).C(userCollection).Find(bson.M{"email": email}).One(&user); err != nil {
		if err == mgo.ErrNotFound {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (d *database) ListUsers(filter *storage.UserFilter) (*storage.ListUsersResponse, error) {
	var users []*storage.User
	var err error
	query := bson.D{}

	if filter.Query != "" {
		query, err = filterparser.ParseUsers(filter.Query)
		if err != nil {
			return nil, err
		}
	}

	mongoQuery := d.session.DB(d.name).C(userCollection).Find(query)
	total, err := mongoQuery.Count()
	if err != nil {
		return nil, err
	}

	if filter.Cursor != "" {
		if !bson.IsObjectIdHex(filter.Cursor) {
			return nil, storage.ErrInvalidCursor
		}

		query = bson.D{{"$and", []bson.D{
			query,
			{{"cursor", bson.D{{"$gt", bson.ObjectIdHex(filter.Cursor)}}}},
		}}}
	}
	mongoQuery = d.session.DB(d.name).C(userCollection).Find(query)
	mongoQuery.Sort("cursor")

	limit := defaultUsersPerRequest
	if filter.PerRequest != 0 {
		limit = int(filter.PerRequest)
	}
	mongoQuery.Limit(limit)

	if err := mongoQuery.All(&users); err != nil {
		return nil, err
	}

	return &storage.ListUsersResponse{
		Total: int64(total),
		Users: users,
	}, nil
}
