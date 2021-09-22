package server_test

import (
	"errors"
	"fmt"
	"github.com/asavt7/antibot-developer-trainee/pkg/configs"
	"github.com/asavt7/antibot-developer-trainee/pkg/server"
	"github.com/asavt7/antibot-developer-trainee/pkg/service"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	StaticContent = "STATIC_CONTENT"
)

type mockHandler struct {
	CallsCount int
}

func (m *mockHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	m.CallsCount += 1
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(StaticContent))
	return
}

var (
	mockRateLimitService = &service.RateLimitCheckerMockService{}
	mockService          = &service.Service{mockRateLimitService}
	mockProtectedHandler = &mockHandler{}
	serv                 = server.NewServer(configs.Config{}, mockService, mockProtectedHandler)
	setupTestCase        = func() {
		mockProtectedHandler.CallsCount = 0
	}
)

func TestMainHandler(t *testing.T) {
	testServ := httptest.NewServer(serv.Handler)
	defer testServ.Close()

	t.Run("not found", func(t *testing.T) {
		res, err := http.Get(fmt.Sprintf("%s/strange/url", testServ.URL))
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, actual %d", res.StatusCode)
		}
	})

	t.Run("no X-Forwarded-For header", func(t *testing.T) {
		res, err := http.Get(fmt.Sprintf("%s/", testServ.URL))
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, actual %d", res.StatusCode)
		}
	})

	t.Run("invalid X-Forwarded-For header", func(t *testing.T) {
		r, err := http.NewRequest("GET", fmt.Sprintf("%s/", testServ.URL), nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("X-Forwarded-For", "qwe.qwe.qwe.123")
		res, err := testServ.Client().Do(r)

		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, actual %d", res.StatusCode)
		}
	})

	t.Run("invalid X-Forwarded-For header", func(t *testing.T) {
		r, err := http.NewRequest("GET", fmt.Sprintf("%s/", testServ.URL), nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("X-Forwarded-For", "2001:0db8:85a3:0000:0000:8a2e:0370:7334")
		res, err := testServ.Client().Do(r)

		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, actual %d", res.StatusCode)
		}
	})

	t.Run("ok, allowed access", func(t *testing.T) {
		setupTestCase()
		mockRateLimitService.IsLimitExceededForIpFunc = func(ipv4Addr net.IP) (bool, error) {
			return false, nil
		}

		r, err := http.NewRequest("GET", fmt.Sprintf("%s", testServ.URL), nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("X-Forwarded-For", "111.111.111.111")
		res, err := testServ.Client().Do(r)

		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, actual %d", res.StatusCode)
		}

		if h := res.Header.Get("Content-Type"); h != "text/plain; charset=utf-8" {
			t.Errorf("Content-Type header should == text/plain; charset=utf-8, actual header %s", h)
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), StaticContent) {
			t.Errorf("shoutl return html static content, actual : %s", body)
		}
		if mockProtectedHandler.CallsCount < 1 {
			t.Errorf("static content handler was not called")
		}
	})

	t.Run("subnet blocked", func(t *testing.T) {
		setupTestCase()
		mockRateLimitService.IsLimitExceededForIpFunc = func(ipv4Addr net.IP) (bool, error) {
			return true, nil
		}

		r, err := http.NewRequest("GET", fmt.Sprintf("%s", testServ.URL), nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("X-Forwarded-For", "111.111.111.111")
		res, err := testServ.Client().Do(r)

		if res.StatusCode != http.StatusTooManyRequests {
			t.Errorf("expected status 429, actual %d", res.StatusCode)
		}
		if mockProtectedHandler.CallsCount > 0 {
			t.Errorf("static content handler was called, but should not")
		}
	})

	t.Run("error from service layer", func(t *testing.T) {
		setupTestCase()
		mockRateLimitService.IsLimitExceededForIpFunc = func(ipv4Addr net.IP) (bool, error) {
			return false, errors.New("error")
		}

		r, err := http.NewRequest("GET", fmt.Sprintf("%s", testServ.URL), nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("X-Forwarded-For", "111.111.111.111")
		res, err := testServ.Client().Do(r)

		if res.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, actual %d", res.StatusCode)
		}
		if mockProtectedHandler.CallsCount > 0 {
			t.Errorf("static content handler was called, but should not")
		}
	})

}

func TestResetHandler(t *testing.T)  {

	testServ := httptest.NewServer(serv.Handler)
	defer testServ.Close()

	t.Run("no X-Forwarded-For header", func(t *testing.T) {
		res, err := http.Get(fmt.Sprintf("%s/reset", testServ.URL))
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, actual %d", res.StatusCode)
		}
	})

	t.Run("invalid X-Forwarded-For header", func(t *testing.T) {
		r, err := http.NewRequest("GET", fmt.Sprintf("%s/reset", testServ.URL), nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("X-Forwarded-For", "qwe.qwe.qwe.123")
		res, err := testServ.Client().Do(r)

		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, actual %d", res.StatusCode)
		}
	})

	t.Run("invalid X-Forwarded-For header", func(t *testing.T) {
		r, err := http.NewRequest("GET", fmt.Sprintf("%s/reset", testServ.URL), nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("X-Forwarded-For", "2001:0db8:85a3:0000:0000:8a2e:0370:7334")
		res, err := testServ.Client().Do(r)

		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, actual %d", res.StatusCode)
		}
	})

	t.Run("ok", func(t *testing.T) {
		setupTestCase()
		mockRateLimitService.ResetPrefixForIpv4Func = func(ipv4Addr net.IP) error {
			return nil
		}

		r, err := http.NewRequest("GET", fmt.Sprintf("%s/reset", testServ.URL), nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("X-Forwarded-For", "111.111.111.111")
		res, err := testServ.Client().Do(r)

		if res.StatusCode != http.StatusNoContent {
			t.Errorf("expected status 204, actual %d", res.StatusCode)
		}
	})
	t.Run("error from service layer", func(t *testing.T) {
		setupTestCase()
		mockRateLimitService.ResetPrefixForIpv4Func = func(ipv4Addr net.IP)  error {
			return errors.New("error")
		}

		r, err := http.NewRequest("GET", fmt.Sprintf("%s/reset", testServ.URL), nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("X-Forwarded-For", "111.111.111.111")
		res, err := testServ.Client().Do(r)

		if res.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, actual %d", res.StatusCode)
		}
	})

}