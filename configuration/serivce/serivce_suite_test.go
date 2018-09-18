package service_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSerivce(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Serivce Suite")
}
