package service

import (
	"github.com/asavt7/antibot-developer-trainee/pkg/configs"
	"github.com/asavt7/antibot-developer-trainee/pkg/mocks"
	"net"
	"testing"
)

var (
	conf = configs.Config{
		PrefixSize: 24,
	}
	rateLimitStoreMock = &mocks.RateLimitStoreMock{}
	service            = NewServiceImpl(conf, rateLimitStoreMock)
)

func TestRateLimitCheckerImpl_IsLimitExceededForIp(t *testing.T) {
	testTable := []struct {
		name           string
		prefixSize     int
		ipv4           net.IP
		expectedSubnet string
	}{
		{
			name:           "111.111.111.111",
			prefixSize:     24,
			ipv4:           net.ParseIP("111.111.111.111"),
			expectedSubnet: "111.111.111.0",
		},
		{
			name:           "222.222.222.123",
			prefixSize:     24,
			ipv4:           net.ParseIP("222.222.222.123"),
			expectedSubnet: "222.222.222.0",
		},
		{
			name:           "222.222.222.123",
			prefixSize:     32,
			ipv4:           net.ParseIP("222.222.222.123"),
			expectedSubnet: "222.222.222.123",
		},
		{
			name:           "222.222.222.123",
			prefixSize:     8,
			ipv4:           net.ParseIP("222.222.222.123"),
			expectedSubnet: "222.0.0.0",
		},
	}

	var subnetArg string
	rateLimitStoreMock.CheckFunc = func(subnet string) bool {
		subnetArg = subnet
		return false
	}

	for _, tc := range testTable {
		subnetArg = ""
		t.Run(tc.name, func(t *testing.T) {

			service = NewServiceImpl(configs.Config{PrefixSize: tc.prefixSize}, rateLimitStoreMock)

			isBlocked, err := service.IsLimitExceededForIp(tc.ipv4)
			if err != nil {
				t.Errorf("expected nil error")
			}
			if isBlocked {
				t.Errorf("expected false")
			}
			if subnetArg != tc.expectedSubnet {
				t.Errorf("expected subnet %s != actual %s", tc.expectedSubnet, subnetArg)
			}
		})
	}

	t.Run("invalid arg", func(t *testing.T) {
		service = NewServiceImpl(configs.Config{PrefixSize: 24}, rateLimitStoreMock)
		_, err := service.IsLimitExceededForIp(net.ParseIP("444.444.444.444"))
		if err == nil {
			t.Errorf("expected error for invalid ip addr")
		}
	})

}
func TestRateLimitCheckerImpl_ResetPrefixForIpv4(t *testing.T) {
	var subnetArg string
	rateLimitStoreMock.ResetFunc = func(subnet string) {
		subnetArg = subnet
	}

	t.Run("ok", func(t *testing.T) {
		subnetArg = ""
		service = NewServiceImpl(configs.Config{PrefixSize: 24}, rateLimitStoreMock)

		err := service.ResetPrefixForIpv4(net.ParseIP("123.123.123.123"))
		if err != nil {
			t.Errorf("expected no error")
		}
		if subnetArg != "123.123.123.0" {
			t.Errorf("expected subnet %s != actual %s", "123.123.123.0", subnetArg)
		}
	})

	t.Run("invalid arg", func(t *testing.T) {
		service = NewServiceImpl(configs.Config{PrefixSize: 24}, rateLimitStoreMock)
		err := service.ResetPrefixForIpv4(net.ParseIP("444.444.444.444"))
		if err == nil {
			t.Errorf("expected error for invalid ip addr")
		}
	})

}
