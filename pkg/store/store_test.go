package store

import (
	"github.com/asavt7/antibot-developer-trainee/pkg/configs"
	"testing"
	"time"
)

const subnet = "subnet"

func initStore(config configs.Config) (*InMemoryStoreRateLimitStore, func()) {
	inMemStore := NewInMemoryStoreRateLimitStore(config)
	inMemStore.InitStore()
	return inMemStore, func() {
		inMemStore.CloseStore()
	}
}

func TestNewInMemoryStoreRateLimitStoreName(t *testing.T) {

	t.Run("ok case ", func(t *testing.T) {
		inMemStore, closeStore := initStore(configs.Config{
			RequestLimit:    1,
			TimeInterval:    300 * time.Millisecond,
			BlockingTimeout: 3 * time.Second,
		})
		defer closeStore()

		res := inMemStore.Check(subnet)
		if res {
			t.Errorf("expected false for 1 req of 1 max per 1 second")
		}
		time.Sleep(500 * time.Millisecond)
		res = inMemStore.Check(subnet)
		if res {
			t.Errorf("expected false for 1 req of 1 max per 1 second")
		}
		time.Sleep(500 * time.Millisecond)
		res = inMemStore.Check(subnet)
		if res {
			t.Errorf("expected false for 1 req of 1 max per 1 second")
		}
	})

	t.Run("block ", func(t *testing.T) {
		inMemStore, closeStore := initStore(configs.Config{
			RequestLimit:    1,
			TimeInterval:    3 * time.Second,
			BlockingTimeout: 3 * time.Second,
		})
		defer closeStore()

		res := inMemStore.Check(subnet)
		if res {
			t.Errorf("expected false for 1 req of 1 max per 1 second")
		}
		inMemStore.Check(subnet)

		time.Sleep(100 * time.Millisecond)

		res = inMemStore.Check(subnet)
		if !res {
			t.Errorf("expected false for 1 req of 1 max per 1 second")
		}
	})

	t.Run("block on blocking timeout ", func(t *testing.T) {
		inMemStore, closeStore := initStore(configs.Config{
			RequestLimit:    1,
			TimeInterval:    300 * time.Millisecond,
			BlockingTimeout: 3 * time.Second,
		})
		defer closeStore()

		res := inMemStore.Check(subnet)
		if res {
			t.Errorf("expected false for 1 req of 1 max per 1 second")
		}
		inMemStore.Check(subnet)
		inMemStore.Check(subnet)

		time.Sleep(100 * time.Millisecond)

		ticker := time.NewTicker(800 * time.Millisecond)
		timer := time.NewTimer(2950 * time.Millisecond)

		for loop := true; loop; {
			select {
			case <-ticker.C:
				res := inMemStore.Check(subnet)
				if !res {
					t.Errorf("expected blocked")
				}
			case <-timer.C:
				loop = false
				ticker.Stop()
				break
			}
		}

		res = inMemStore.Check(subnet)
		if res {
			t.Errorf("expected unblocked after timeout")
		}
	})

}
