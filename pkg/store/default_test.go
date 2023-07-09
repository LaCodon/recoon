package store_test

import (
	"github.com/lacodon/recoon/pkg/api"
	metav1 "github.com/lacodon/recoon/pkg/api/v1/meta"
	"github.com/lacodon/recoon/pkg/store"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type TestObj struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Data              string `json:"data"`
}

func (t *TestObj) DeepCopy() api.Object {
	return &TestObj{
		TypeMeta:   t.TypeMeta.DeepCopy(),
		ObjectMeta: t.ObjectMeta.DeepCopy(),
		Data:       t.Data,
	}
}

var _ = Describe("DefaultStore", func() {
	Describe("set and get object", Ordered, func() {
		var api *store.DefaultStore
		obj := &TestObj{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Data: "my-data",
		}

		It("should open the database", func() {
			var err error
			api, err = store.NewDefaultStore("./bbolt.db", store.WithTempFs)
			Expect(err).To(BeNil())
		})

		It("should insert the object", func() {
			Expect(api.Create(obj)).To(BeNil())
		})

		It("should get the object", func() {
			getObj := &TestObj{}

			Expect(api.Get(metav1.NamespaceName{
				Name:      obj.Name,
				Namespace: obj.Namespace,
			}, getObj)).To(BeNil())

			Expect(getObj).NotTo(BeNil())

			Expect(getObj.Version).To(BeEquivalentTo("v1"))
			Expect(getObj.Kind).To(BeEquivalentTo("TestObject"))
			Expect(getObj.Name).To(BeEquivalentTo(obj.Name))
			Expect(getObj.Namespace).To(BeEquivalentTo(obj.Namespace))
			Expect(getObj.Data).To(BeEquivalentTo(obj.Data))
		})

		It("should close the database", func() {
			Expect(api.Close()).To(BeNil())
		})
	})

	Describe("fail to get unavailable object", Ordered, func() {
		var api *store.DefaultStore

		It("should open the database", func() {
			var err error
			api, err = store.NewDefaultStore("./bbolt.db", store.WithTempFs)
			Expect(err).To(BeNil())
		})

		It("should not find the object", func() {
			getObj := &TestObj{}

			Expect(api.Get(metav1.NamespaceName{
				Name:      "unavailable",
				Namespace: "unavailable",
			}, getObj)).To(MatchError(store.ErrNotFound))
		})

		It("should close the database", func() {
			Expect(api.Close()).To(BeNil())
		})
	})

	Describe("fail to set invalid object", Ordered, func() {
		var api *store.DefaultStore

		It("should open the database", func() {
			var err error
			api, err = store.NewDefaultStore("./bbolt.db", store.WithTempFs)
			Expect(err).To(BeNil())
		})

		It("should not set the object with invalid name", func() {
			obj := &TestObj{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid/name",
					Namespace: "namespace",
				},
				Data: "data",
			}

			Expect(api.Create(obj)).To(MatchError(store.ErrInvalid))
		})

		It("should not set the object with invalid namespace", func() {
			obj := &TestObj{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "invalid/namespace",
				},
				Data: "data",
			}

			Expect(api.Create(obj)).To(MatchError(store.ErrInvalid))
		})

		It("should not set the object with empty name", func() {
			obj := &TestObj{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "", // empty
					Namespace: "namespace",
				},
				Data: "data",
			}

			Expect(api.Create(obj)).To(MatchError(store.ErrNameEmpty))
		})

		It("should not set the object with empty namespace", func() {
			obj := &TestObj{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "", // empty
				},
				Data: "data",
			}

			Expect(api.Create(obj)).To(MatchError(store.ErrNamespaceEmpty))
		})

		It("should close the database", func() {
			Expect(api.Close()).To(BeNil())
		})
	})

	Describe("get list of objects", Ordered, func() {
		var api *store.DefaultStore

		obj1 := &TestObj{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj-1",
				Namespace: "test",
			},
			Data: "data1",
		}

		obj2 := &TestObj{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj-2",
				Namespace: "test",
			},
			Data: "data2",
		}

		It("should open the database", func() {
			var err error
			api, err = store.NewDefaultStore("./bbolt.db", store.WithTempFs)
			Expect(err).To(BeNil())
		})

		It("should insert objects", func() {
			Expect(api.Create(obj1)).To(BeNil())
			Expect(api.Create(obj2)).To(BeNil())

			Expect(obj1.GetVersionKind().Kind).NotTo(BeEmpty())
			Expect(obj1.GetVersionKind().Version).NotTo(BeEmpty())

			Expect(obj2.GetVersionKind().Kind).NotTo(BeEmpty())
			Expect(obj2.GetVersionKind().Version).NotTo(BeEmpty())
		})

		It("should get list of objects", func() {
			list, err := api.List(metav1.VersionKind{Version: "v1", Kind: "TestObject"})
			Expect(err).To(BeNil())
			Expect(list).To(HaveLen(2))
			Expect(list).To(ContainElements(obj1, obj2))
		})

		It("should close the database", func() {
			Expect(api.Close()).To(BeNil())
		})
	})

	Describe("update object", func() {
		var api *store.DefaultStore

		obj := &TestObj{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj",
				Namespace: "test",
			},
			Data: "data",
		}

		It("should open the database", func() {
			var err error
			api, err = store.NewDefaultStore("./bbolt.db", store.WithTempFs)
			Expect(err).To(BeNil())
		})

		It("should insert the object", func() {
			Expect(api.Create(obj)).To(BeNil())
		})

		It("should fail to insert the object again", func() {
			Expect(api.Create(obj)).To(MatchError(store.ErrAlreadyExists))
		})

		It("should update the object", func() {
			obj.Data = "data-update"
			Expect(api.Update(obj)).To(BeNil())
			Expect(obj.RessourceVersion).To(BeEquivalentTo(1))
		})

		It("should update the object again", func() {
			Expect(api.Update(obj)).To(BeNil())
			Expect(obj.RessourceVersion).To(BeEquivalentTo(2))
		})

		It("should fail to skip resource versions", func() {
			obj.RessourceVersion = 0
			Expect(api.Update(obj)).To(MatchError(store.ErrObjectChanged))
		})

		It("should delete object", func() {
			Expect(api.Delete(obj.GetVersionKind(), obj.GetNamespaceName())).To(BeNil())
		})

		It("should close the database", func() {
			Expect(api.Close()).To(BeNil())
		})
	})
})
