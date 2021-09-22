package service

import "net"

type RateLimitCheckerMockService struct {
	IpToCalls                map[string]int
	IsLimitExceededForIpFunc func(ipv4Addr net.IP) (bool, error)
	ResetPrefixForIpv4Func   func(ipv4Addr net.IP) error
}

func (m *RateLimitCheckerMockService) IsLimitExceededForIp(ipv4Addr net.IP) (bool, error) {
	return m.IsLimitExceededForIpFunc(ipv4Addr)
}

func (m *RateLimitCheckerMockService) ResetPrefixForIpv4(ipv4Addr net.IP) error {
	return m.ResetPrefixForIpv4Func(ipv4Addr)
}

