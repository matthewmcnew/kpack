package registry_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"

	"github.com/sclevine/spec"

	"build-service/image"
	"build-service/lifecycle_image"
)

func TestRemoteImage(t *testing.T) {
	spec.Run(t, "Test Remote Image", testRemoteImageAccess)
}

func testRemoteImageAccess(t *testing.T, when spec.G, it spec.S) {
	var (
		Expect        func(interface{}, ...interface{}) GomegaAssertion
		handler       *http.ServeMux
		server        *httptest.Server
		accessChecker *lifecycle_image.RemoteImageAccessChecker
		tagName       string
	)

	it.Before(func() {
		Expect = NewGomegaWithT(t).Expect

		handler = http.NewServeMux()
		server = httptest.NewServer(handler)
		tagName = fmt.Sprintf("%s/some/image:tag", server.URL[7:])

		accessChecker = &lifecycle_image.RemoteImageAccessChecker{
			KeychainFactory: &testKeychainFactory{expectedImage: tagName, t: t},
		}
	})

	when("HasWriteAccess", func() {
		it("true when has permission", func() {
			handler.HandleFunc("/v2/some/image/blobs/uploads/", func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(201)
			})

			handler.HandleFunc("/v2/", func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(200)
			})

			hasAccess, err := accessChecker.HasWriteAccess(image.NewNoAuthImageRef(tagName))
			Expect(err).NotTo(HaveOccurred())
			Expect(hasAccess).To(BeTrue())
		})

		it("requests scope push permission", func() {
			handler.HandleFunc("/unauthorized-token/", func(writer http.ResponseWriter, request *http.Request) {
				values, err := url.ParseQuery(request.URL.RawQuery)
				Expect(err).NotTo(HaveOccurred())
				Expect(values.Get("scope")).To(Equal("repository:some/image:push,pull"))
			})

			handler.HandleFunc("/v2/", func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Add("WWW-Authenticate", fmt.Sprintf("bearer realm=%s/unauthorized-token/", server.URL))
				writer.WriteHeader(401)
			})

			_, _ = accessChecker.HasWriteAccess(image.NewNoAuthImageRef(tagName))
		})

		it("uses the keychain auth", func() {
			handler.HandleFunc("/unauthorized-token/", func(writer http.ResponseWriter, request *http.Request) {
				Expect(request.Header.Get("Authorization")).To(Equal(testAuthValue))
			})

			handler.HandleFunc("/v2/", func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Add("WWW-Authenticate", fmt.Sprintf("bearer realm=%s/unauthorized-token/", server.URL))
				writer.WriteHeader(401)
			})

			_, _ = accessChecker.HasWriteAccess(image.NewNoAuthImageRef(tagName))
		})

		it("false when fetching token is unauthorized", func() {
			handler.HandleFunc("/unauthorized-token/", func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(401)
				writer.Write([]byte(`{"errors": [{"code":  "UNAUTHORIZED"}]}`))
			})

			handler.HandleFunc("/v2/", func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Add("WWW-Authenticate", fmt.Sprintf("bearer realm=%s/unauthorized-token/", server.URL))
				writer.WriteHeader(401)
			})

			tagName := fmt.Sprintf("%s/some/image:tag", server.URL[7:])

			hasAccess, err := accessChecker.HasWriteAccess(image.NewNoAuthImageRef(tagName))
			Expect(err).NotTo(HaveOccurred())
			Expect(hasAccess).To(BeFalse())
		})

		it("false when server responds with unauthorized but without a code such as on artifactory", func() {
			handler.HandleFunc("/unauthorized-token/", func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(401)
				writer.Write([]byte(`{"statusCode":401,"details":"BAD_CREDENTIAL"}`))
			})

			handler.HandleFunc("/v2/", func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Add("WWW-Authenticate", fmt.Sprintf("bearer realm=%s/unauthorized-token/", server.URL))
				writer.WriteHeader(401)
			})

			tagName := fmt.Sprintf("%s/some/image:tag", server.URL[7:])

			hasAccess, err := accessChecker.HasWriteAccess(image.NewNoAuthImageRef(tagName))
			Expect(err).NotTo(HaveOccurred())
			Expect(hasAccess).To(BeFalse())
		})

		it("false when does not have permission", func() {
			handler.HandleFunc("/v2/some/image/blobs/uploads/", func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(403)
			})

			handler.HandleFunc("/v2/", func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(200)
			})

			tagName := fmt.Sprintf("%s/some/image:tag", server.URL[7:])

			hasAccess, err := accessChecker.HasWriteAccess(image.NewNoAuthImageRef(tagName))
			Expect(err).NotTo(HaveOccurred())
			Expect(hasAccess).To(BeFalse())
		})

		it("false when cannot reach server with an error", func() {
			handler.HandleFunc("/v2/", func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(404)
			})

			tagName := fmt.Sprintf("%s/some/image:tag", server.URL[7:])

			hasAccess, err := accessChecker.HasWriteAccess(image.NewNoAuthImageRef(tagName))
			Expect(err).To(HaveOccurred())
			Expect(hasAccess).To(BeFalse())
		})
	})
}

type testKeychainFactory struct {
	expectedImage string
	t             *testing.T
}

type testKeychain struct {
}

type testAuth struct {
}

const testAuthValue = "test auth"

func (testAuth) Authorization() (string, error) {
	return testAuthValue, nil
}

func (testKeychain) Resolve(name.Registry) (authn.Authenticator, error) {
	return &testAuth{}, nil
}

func (kf *testKeychainFactory) KeychainForImageRef(ref image.ImageRef) authn.Keychain {
	if ref.RepoName() != kf.expectedImage {
		kf.t.Fatalf("expected %s got %s", kf.expectedImage, ref.RepoName())
	}

	return &testKeychain{}
}
