package ubi8nodeenginebuildpackextension_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect
	)

	it("It should pass...", func() {

		var hello = true
		Expect(true).To(Equal(hello))

	})
}
