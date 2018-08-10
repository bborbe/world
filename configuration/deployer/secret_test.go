package deployer

import (
	"bytes"

	"github.com/bborbe/world"
	"github.com/go-yaml/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SecretDeployer", func() {
	It("secret", func() {
		secretDeployer := &SecretDeployer{
			Namespace: "banana",
			Secrets: world.Secrets{
				"secret": &world.SecretValueStatic{
					Content: []byte("hello world"),
				},
			},
		}
		b := &bytes.Buffer{}
		secret, err := secretDeployer.secret()
		Expect(err).NotTo(HaveOccurred())
		err = yaml.NewEncoder(b).Encode(secret)
		Expect(err).NotTo(HaveOccurred())
		Expect(b.String()).To(Equal(`apiVersion: v1
kind: Secret
metadata:
  namespace: banana
  name: banana
  labels:
    app: banana
type: Opaque
data:
  secret: aGVsbG8gd29ybGQ=
`))
	})
})
