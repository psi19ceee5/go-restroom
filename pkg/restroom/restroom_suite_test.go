package restroom_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRestroom(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Restroom Suite")
}
