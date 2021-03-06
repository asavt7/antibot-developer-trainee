package store

import (
	"context"
	"github.com/asavt7/antibot-developer-trainee/pkg/configs"
	"log"
	"sync"
	"time"
)

type RateLimitStore interface {
	Check(subnet string) bool
	Reset(subnet string)
}

type SubnetBlocksMap struct {
	sync.RWMutex
	m map[string]bool
}

type SubnetCountMap struct {
	sync.Mutex
	m map[string]int
}

type SubnetTimer struct {
	timer  *time.Timer
	subnet string
}

type InMemoryStoreRateLimitStore struct {
	subnetBlocksMap SubnetBlocksMap
	subnetCountMap  SubnetCountMap
	reqLimit        int
	timeLimit       time.Duration
	timeout         time.Duration

	reqEventCh       chan string
	resetTimeLimitCh chan SubnetTimer
	resetCounterCh   chan string

	unlockTimeLimitCh chan SubnetTimer
	resetTimerCh      chan SubnetTimer

	blockSubnetCh   chan string
	unblockSubnetCh chan string

	ctx    context.Context
	cancel context.CancelFunc
}

func (i *InMemoryStoreRateLimitStore) Check(subnet string) bool {
	i.subnetBlocksMap.RLock()
	isBlocked, _ := i.subnetBlocksMap.m[subnet]
	i.subnetBlocksMap.RUnlock()
	log.Printf("Request for subnet %s isBlocked: %t", subnet, isBlocked)
	if !isBlocked {
		i.reqEventCh <- subnet
	}
	return isBlocked
}

func (i *InMemoryStoreRateLimitStore) startReqEventListener() {
	go func() {
		for {
			select {
			case subnet := <-i.reqEventCh:
				count, inMap := i.subnetCountMap.m[subnet]
				if !inMap {
					log.Printf("reqEventListener create timer to reset counter for subnet %s", subnet)
					i.resetTimerCh <- SubnetTimer{
						timer:  time.NewTimer(i.timeLimit),
						subnet: subnet,
					}
				}
				log.Printf("reqEventListener got event : subnet %s requestCount %d/%d", subnet, count+1, i.reqLimit)
				i.subnetCountMap.m[subnet] = count + 1

				if i.subnetCountMap.m[subnet] > i.reqLimit {
					i.blockSubnet(subnet)
				}
			case subnet := <-i.resetCounterCh:
				_, inMap := i.subnetCountMap.m[subnet]
				if inMap {
					i.subnetCountMap.m[subnet] = 0
				}
			case <-i.ctx.Done():
				log.Println("Cancelled ReqEventListener")
				return
			}
		}
	}()
}

func (i *InMemoryStoreRateLimitStore) startResetListener() {
	go func() {
		for {
			select {
			case subnetTimer := <-i.resetTimerCh:
				<-subnetTimer.timer.C
				log.Printf("resetting counter for subnet %s", subnetTimer.subnet)
				i.resetCounterCh <- subnetTimer.subnet
				i.resetTimerCh <- SubnetTimer{
					timer:  time.NewTimer(i.timeLimit),
					subnet: subnetTimer.subnet,
				}
			case <-i.ctx.Done():
				log.Println("Cancelled ResetListener")
				return
			}
		}
	}()
}

func (i *InMemoryStoreRateLimitStore) startBlockListener() {
	go func() {
		for {
			select {
			case subnet := <-i.blockSubnetCh:
				log.Printf("blocking for subnet %s", subnet)
				i.subnetBlocksMap.Lock()
				i.subnetBlocksMap.m[subnet] = true
				i.subnetBlocksMap.Unlock()
			case subnet := <-i.unblockSubnetCh:
				log.Printf("unblocking for subnet %s", subnet)
				i.subnetBlocksMap.Lock()
				i.subnetBlocksMap.m[subnet] = false
				i.subnetBlocksMap.Unlock()
			case <-i.ctx.Done():
				log.Println("Cancelled BlockListener")
				return
			}
		}
	}()
}

func (i *InMemoryStoreRateLimitStore) Reset(subnet string) {
	log.Printf("resetting blocking and request counter for subnet %s", subnet)
	i.resetCounterCh <- subnet
	i.unblockSubnetCh <- subnet
}

func (i *InMemoryStoreRateLimitStore) blockSubnet(subnet string) {
	i.blockSubnetCh <- subnet
	i.unlockTimeLimitCh <- SubnetTimer{
		timer:  time.NewTimer(i.timeout),
		subnet: subnet,
	}
}

func (i *InMemoryStoreRateLimitStore) startUnBlockTimerListener() {

	go func() {
		for {
			select {
			case subnetTimer := <-i.unlockTimeLimitCh:
				<-subnetTimer.timer.C
				i.unblockSubnetCh <- subnetTimer.subnet
			case <-i.ctx.Done():
				log.Println("Cancelled UnBlockTimerListener")
				return
			}
		}
	}()
}

func (i *InMemoryStoreRateLimitStore) InitStore() {
	i.startReqEventListener()
	i.startBlockListener()
	i.startUnBlockTimerListener()
	i.startResetListener()
}

func (i *InMemoryStoreRateLimitStore) CloseStore() {
	i.cancel()
}

func NewInMemoryStoreRateLimitStore(conf configs.Config) *InMemoryStoreRateLimitStore {
	ctx, cancel := context.WithCancel(context.Background())
	return &InMemoryStoreRateLimitStore{
		subnetBlocksMap:   SubnetBlocksMap{m: make(map[string]bool)},
		subnetCountMap:    SubnetCountMap{m: make(map[string]int)},
		reqLimit:          conf.RequestLimit,
		timeLimit:         conf.TimeInterval,
		timeout:           conf.BlockingTimeout,
		reqEventCh:        make(chan string, 1000),
		resetTimeLimitCh:  make(chan SubnetTimer, 1000),
		resetCounterCh:    make(chan string, 1000),
		unlockTimeLimitCh: make(chan SubnetTimer, 1000),
		resetTimerCh:      make(chan SubnetTimer, 1000),
		blockSubnetCh:     make(chan string, 1000),
		unblockSubnetCh:   make(chan string, 1000),
		ctx:               ctx,
		cancel:            cancel,
	}
}
