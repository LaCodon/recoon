package store_test

import (
	metav1 "github.com/lacodon/recoon/pkg/api/v1/meta"
	"github.com/lacodon/recoon/pkg/schema"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Store Suite")
}

var _ = BeforeSuite(func() {
	schema.Register(metav1.VersionKind{
		Version: "v1",
		Kind:    "TestObject",
	}, &TestObj{})
})
