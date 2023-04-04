package store

import (
	"bytes"
	"encoding/json"
	"github.com/lacodon/recoon/pkg/api"
	"github.com/lacodon/recoon/pkg/api/v1/meta"
	"github.com/lacodon/recoon/pkg/config"
	"github.com/lacodon/recoon/pkg/schema"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

// force interface implementation during compile time
var _ GetterSetter = &DefaultStore{}

func NewDefaultStore(opts ...AdaptOption) (*DefaultStore, error) {
	options := &bolt.Options{
		Timeout:      5 * time.Second,
		FreelistType: bolt.FreelistMapType,
	}

	for _, opt := range opts {
		opt(options)
	}

	db, err := bolt.Open(config.Cfg.Store.DatabaseFile, 0664, options)
	if err != nil {
		return nil, err
	}

	return &DefaultStore{
		db:         db,
		eventsChan: make(chan Event, 100),
	}, nil
}

type AdaptOption func(options *bolt.Options)

// WithTempFs uses a temporary filesystem instead of a real one
func WithTempFs(options *bolt.Options) {
	options.OpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		path, err := os.MkdirTemp("", "recoon-test-*")
		if err != nil {
			return nil, err
		}

		filename := filepath.Base(name)
		logrus.Debug("testing with temp bbolt instance at", filename)
		return os.OpenFile(filepath.Join(path, filename), flag, perm)
	}
}

type DefaultStore struct {
	db         *bolt.DB
	eventsChan chan Event
}

func (d *DefaultStore) List(vk metav1.VersionKind, opts ...ListOption) (api.ObjectList, error) {
	typ, err := schema.GetType(vk)
	if err != nil {
		return nil, err
	}

	result := make([]api.Object, 0)

	if err := d.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(d.makeBucketName(vk))
		if bucket == nil {
			return errors.WithMessage(ErrNotFound, vk.String())
		}

		if opts != nil {
			cfg := &listConfig{}
			for _, opt := range opts {
				opt(cfg)
			}

			c := bucket.Cursor()
			prefix := []byte(cfg.Prefix)
			for key, value := c.Seek(prefix); key != nil && bytes.HasPrefix(key, prefix); key, value = c.Next() {
				obj := reflect.New(typ.Elem()).Interface()
				if err := json.Unmarshal(value, obj); err != nil {
					return err
				}
				result = append(result, obj.(api.Object))
			}
			return nil
		}

		return bucket.ForEach(func(key, value []byte) error {
			obj := reflect.New(typ.Elem()).Interface()
			if err := json.Unmarshal(value, obj); err != nil {
				return err
			}
			result = append(result, obj.(api.Object))
			return nil
		})
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (d *DefaultStore) Close() error {
	return d.db.Close()
}

func (d *DefaultStore) Get(namespaceName metav1.NamespaceName, object api.Object) error {
	vk, err := schema.GetVersionKind(object)
	if err != nil {
		return err
	}

	return d.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(d.makeBucketName(vk))
		if bucket == nil {
			return errors.WithMessage(ErrNotFound, vk.String()+"/"+namespaceName.String())
		}

		data := bucket.Get(d.makeKey(namespaceName))
		if data == nil {
			return errors.WithMessage(ErrNotFound, vk.String()+"/"+namespaceName.String())
		}

		return json.Unmarshal(data, object)
	})
}

func (d *DefaultStore) Create(object api.Object) error {
	vk, err := schema.GetVersionKind(object)
	if err != nil {
		return err
	}

	if err := d.validateObj(object); err != nil {
		return err
	}

	// set version and kind
	reflect.ValueOf(object).Elem().FieldByName(reflect.TypeOf(metav1.TypeMeta{}).Name()).Set(reflect.ValueOf(metav1.TypeMeta{
		Kind:    vk.Kind,
		Version: vk.Version,
	}))

	if err := d.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(d.makeBucketName(vk))
		if err != nil {
			return err
		}

		objKey := d.makeKey(object.GetNamespaceName())

		if bucket.Get(objKey) != nil {
			return errors.WithMessage(ErrAlreadyExists, object.GetVersionKind().String()+"/"+object.GetNamespaceName().String())
		}

		// set resource version
		reflect.ValueOf(object).Elem().FieldByName(reflect.TypeOf(metav1.ObjectMeta{}).Name()).Set(reflect.ValueOf(metav1.ObjectMeta{
			Name:             object.GetName(),
			Namespace:        object.GetNamespace(),
			RessourceVersion: 0,
		}))

		data, err := json.Marshal(object)
		if err != nil {
			return err
		}

		return bucket.Put(objKey, data)
	}); err != nil {
		return err
	}

	d.eventsChan <- Event{
		Type:   EventTypeAdd,
		Object: object.DeepCopy(),
	}

	return nil
}

func (d *DefaultStore) Update(object api.Object) error {
	vk, err := schema.GetVersionKind(object)
	if err != nil {
		return err
	}

	if err := d.validateObj(object); err != nil {
		return err
	}

	if object.GetVersionKind().Version != vk.Version || object.GetVersionKind().Kind != vk.Kind {
		return errors.WithMessage(ErrInvalid, "object has wrong version or kind")
	}

	currentObj := object.DeepCopy()

	if err := d.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(d.makeBucketName(vk))
		if bucket == nil {
			return ErrNotFound
		}

		objKey := d.makeKey(object.GetNamespaceName())

		currentObjData := bucket.Get(objKey)
		if currentObjData == nil {
			return errors.WithMessage(ErrNotFound, object.GetVersionKind().String()+"/"+object.GetNamespaceName().String())
		}

		if err := json.Unmarshal(currentObjData, currentObj); err != nil {
			return errors.WithMessage(err, "failed to unmarshal current object")
		}

		if currentObj.GetRessourceVersion() != object.GetRessourceVersion() {
			return ErrObjectChanged
		}

		// set resource version
		reflect.ValueOf(object).Elem().FieldByName(reflect.TypeOf(metav1.ObjectMeta{}).Name()).Set(reflect.ValueOf(metav1.ObjectMeta{
			Name:             object.GetName(),
			Namespace:        object.GetNamespace(),
			RessourceVersion: object.GetRessourceVersion() + 1,
		}))

		data, err := json.Marshal(object)
		if err != nil {
			return err
		}

		return bucket.Put(objKey, data)
	}); err != nil {
		return err
	}

	d.eventsChan <- Event{
		Type:           EventTypeUpdate,
		PreviousObject: currentObj.DeepCopy(),
		Object:         object.DeepCopy(),
	}

	return nil
}

func (d *DefaultStore) Delete(vk metav1.VersionKind, namespaceName metav1.NamespaceName) error {
	var data []byte

	if err := d.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(d.makeBucketName(vk))
		if bucket == nil {
			return nil
		}

		data = bucket.Get(d.makeKey(namespaceName))
		if data == nil {
			return nil
		}

		return bucket.Delete(d.makeKey(namespaceName))
	}); err != nil {
		return err
	}

	d.eventsChan <- Event{
		Type: EventTypeDelete,
		Object: &api.GenericObject{
			TypeMeta: metav1.TypeMeta{
				Version: vk.Version,
				Kind:    vk.Kind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      namespaceName.Name,
				Namespace: namespaceName.Namespace,
			},
			Data: data,
		},
	}

	return nil
}

func (d *DefaultStore) EventsChan() <-chan Event {
	return d.eventsChan
}

func (d *DefaultStore) makeBucketName(vk metav1.VersionKind) []byte {
	return []byte(vk.Kind + "/" + vk.Version)
}

func (d *DefaultStore) makeKey(namespaceName metav1.NamespaceName) []byte {
	return []byte(namespaceName.Namespace + "/" + namespaceName.Name)
}

func (d *DefaultStore) validateObj(object api.Object) error {
	if object.GetNamespace() == "" {
		return ErrNamespaceEmpty
	}

	if object.GetName() == "" {
		return ErrNameEmpty
	}

	if strings.Contains(object.GetName(), "/") || strings.Contains(object.GetNamespace(), "/") {
		return errors.WithMessage(ErrInvalid, "name and namespace must not contain '/'")
	}

	return nil
}
