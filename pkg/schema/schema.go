package schema

import (
	"fmt"
	"github.com/lacodon/recoon/pkg/api"
	"github.com/lacodon/recoon/pkg/api/v1/meta"
	"github.com/sirupsen/logrus"
	"reflect"
)

var schema = &Schema{
	typeToKind: make(map[reflect.Type]metav1.VersionKind),
	kindToType: make(map[metav1.VersionKind]reflect.Type),
}

// Register a new type
var Register = schema.Register

// GetVersionKind of object
var GetVersionKind = schema.GetVersionKind

// GetType of object
var GetType = schema.GetType

type Schema struct {
	typeToKind map[reflect.Type]metav1.VersionKind
	kindToType map[metav1.VersionKind]reflect.Type
}

func (s *Schema) Register(vk metav1.VersionKind, object api.Object) {
	valueOf := reflect.ValueOf(object)
	if valueOf.Kind() != reflect.Ptr || valueOf.IsNil() {
		panic("you can only register non-nil pointer types")
	}

	if vk.Version == "" || vk.Kind == "" {
		panic("version and kind must be set")
	}

	if _, ok := s.typeToKind[reflect.TypeOf(object)]; ok {
		panic(fmt.Sprintf("type '%s' already registered", vk.String()))
	}

	s.typeToKind[reflect.TypeOf(object)] = vk
	s.kindToType[vk] = reflect.TypeOf(object)

	logrus.WithField("type", vk.String()).Debug("registered new schema type")
}

func (s *Schema) GetVersionKind(object api.Object) (metav1.VersionKind, error) {
	typ := reflect.TypeOf(object)
	kind, ok := s.typeToKind[typ]
	if !ok {
		return metav1.VersionKind{}, fmt.Errorf("unknown object type: %s/%s", typ.PkgPath(), typ.String())
	}

	return kind, nil
}

func (s *Schema) GetType(vk metav1.VersionKind) (reflect.Type, error) {
	typ, ok := s.kindToType[vk]
	if !ok {
		return nil, fmt.Errorf("unknown object kind: %s", vk)
	}

	return typ, nil
}
