package manager

import (
	"time"

	"github.com/Sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/utils"
)

type UserManager struct {
	manager *Manager
	col     *mgo.Collection // default collection
}

func (m *UserManager) Init() error {
	logrus.Infof("Initialize user indexes")
	err := m.col.EnsureIndex(mgo.Index{
		Key:        []string{"email"},
		Unique:     true,
		Background: false,
	})
	if err != nil {
		return err
	}

	// TODO (m0sth8): extract system users creation to project initialization
	agent := &user.User{
		Email:    "agent@barbudo.net",
		Password: "",
	}
	if _, err := m.Create(agent); err != nil {
		if !m.manager.IsDup(err) {
			return err
		}
	}

	return err
}

func (m *UserManager) GetById(id string) (*user.User, error) {
	obj := &user.User{}
	if err := m.col.FindId(bson.ObjectIdHex(id)).One(obj); err != nil {
		return nil, err
	}
	if obj.Avatar == "" {
		obj.Avatar = utils.GetGravatar(obj.Email, 38, utils.AvatarRetro)
	}
	return obj, nil
}

func (m *UserManager) GetByEmail(email string) (*user.User, error) {
	u := &user.User{}
	if err := m.col.Find(bson.D{{"email", email}}).One(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (m *UserManager) All() ([]*user.User, int, error) {
	results := []*user.User{}

	query := &bson.M{}
	q := m.col.Find(query)
	if err := q.All(&results); err != nil {
		return nil, 0, err
	}
	count, err := q.Count()
	if err != nil {
		return nil, 0, err
	}
	return results, count, nil
}

func (m *UserManager) Create(raw *user.User) (*user.User, error) {
	// TODO (m0sth8): add validation
	raw.Id = bson.NewObjectId()
	raw.Created = time.Now()
	raw.Updated = raw.Created
	raw.Avatar = utils.GetGravatar(raw.Email, 38, utils.AvatarRetro)
	if err := m.col.Insert(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (m *UserManager) Update(obj *user.User) error {
	obj.Updated = time.Now()
	if obj.Avatar == "" {
		obj.Avatar = utils.GetGravatar(obj.Email, 38, utils.AvatarRetro)
	}
	return m.col.UpdateId(obj.Id, obj)
}
