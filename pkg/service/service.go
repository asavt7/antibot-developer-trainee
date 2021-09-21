package service

import (
	"errors"
	"github.com/asavt7/antibot-developer-trainee/pkg/configs"
	"github.com/asavt7/antibot-developer-trainee/pkg/store"
	"log"
	"net"
	"time"
)

type RateLimitChecker interface {
	IsLimitExceededForIp(ipv4Addr net.IP) (bool, error)
	ResetPrefixForIpv4(ipv4Addr net.IP) error
}

type Service struct {
	RateLimitChecker
}

type RateLimitCheckerImpl struct {
	prefixSize  int
	mask        net.IPMask
	limit       int
	waitingTime time.Duration
	store       store.RateLimitStore
}

func parseSubnetSizeToMask(size int) (net.IPMask, error) {
	if size > 32 || size < 0 {
		return nil, errors.New("Incorrect subnet size")
	}

	res := make([]byte, 4)
	for i := 1; i < 5; i++ {
		if size/8 >= i {
			res[i-1] = 255
			continue
		} else {
			if (size - i*8) >= 0 {
				res[i-1] = byte(size - i*8)
			} else {
				res[i-1] = 0
			}
		}
	}
	return net.IPv4Mask(res[0], res[1], res[2], res[3]), nil
}

func NewServiceImpl(conf configs.Config, store store.RateLimitStore) *Service {
	mask, err := parseSubnetSizeToMask(conf.PrefixSize)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return &Service{
		&RateLimitCheckerImpl{prefixSize: conf.PrefixSize, limit: conf.RequestLimit, mask: mask, waitingTime: conf.BlockingTimeout, store: store},
	}
}

func (s *RateLimitCheckerImpl) IsLimitExceededForIp(ipv4Addr net.IP) (bool, error) {
	subnet, err := s.parseIpToSubnet(ipv4Addr)
	if err != nil {
		return false, err
	}
	isBlocked := s.store.Check(subnet)
	return isBlocked, nil
}


func (s *RateLimitCheckerImpl) ResetPrefixForIpv4(ipv4Addr net.IP) error {
	subnet, err := s.parseIpToSubnet(ipv4Addr)
	if err != nil {
		return err
	}
	s.store.Reset(subnet)
	return nil
}

func (s *RateLimitCheckerImpl) parseIpToSubnet(ip net.IP) (string, error) {
	subnetIp := ip.Mask(s.mask)
	if subnetIp == nil {
		return "", errors.New("invalid ip provided")
	}
	return subnetIp.String(), nil
}
