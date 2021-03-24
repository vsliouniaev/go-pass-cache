package cache_test

import (
	"fmt"
	. "github.com/onsi/gomega"
	"github.com/vsliouniaev/go-pass-cache/cache"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	numParallel    = 10
	numConsecutive = 100
)

func TestGetKeyTwice(t *testing.T) {
	// Generate keys and store them
	// Get multiple workers to try getting them
	c := cache.New(time.Second)
	n := 10
	workerNotify := make(map[int]chan string, n)
	for i := 0; i < n; i++ {
		workerNotify[i] = make(chan string)
	}
	numAdded := 0
	tim := time.After(time.Second * 5)
	// Store stuff in the cache and notify workers that they should try to get it
	go func() {
		for {
			select {
			case <-tim:
				for i := range workerNotify {
					close(workerNotify[i])
				}
				return
			default:
				keyVal := strconv.Itoa(rand.Int())
				c.AddKey(keyVal, keyVal)
				// since we're using a map, which worker gets it "first" is random even here
				for i := range workerNotify {
					workerNotify[i] <- keyVal
				}
				numAdded++
			}
		}
	}()

	// Workers store their results in independent maps, that we can correlate afterwards
	resultMaps := make([]map[string]struct{}, n)
	for i := range resultMaps {
		resultMaps[i] = make(map[string]struct{})
	}
	wg := sync.WaitGroup{}
	for i := range workerNotify {
		wg.Add(1)
		go func(i int) {
			for {
				k, more := <-workerNotify[i]
				if !more {
					break
				}
				if val, ok := c.TryGet(k); ok {
					resultMaps[i][val] = struct{}{}
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	// Check that we got all keys and that we got them once
	chk := make(map[string]struct{})
	for i := range resultMaps {
		for j := range resultMaps[i] {
			_, found := chk[j]
			if found {
				t.Errorf("Should not have found same key twice")
			} else {
				chk[j] = struct{}{}
			}
		}
	}

	if numAdded != len(chk) {
		t.Errorf("Added %d found %d\n", numAdded, len(chk))
	} else {
		fmt.Printf("Added %d found %d\n", numAdded, len(chk))
	}
}

// Test overall behaviour with access to different keys
func TestStoreKeyTwice(t *testing.T) {
	sleep := time.Millisecond * 20
	rotate := (sleep * 2) + (sleep / 2)
	g := NewWithT(t)
	c := cache.New(rotate)
	// By sleeping for about 1/2 the cache rotation interval writes, second writes and gets will occur in all
	// permutations around the rotation boundary.
	repeatInParallel(func(_ int) {
		key, val := strconv.Itoa(rand.Int()), strconv.Itoa(rand.Int())
		c.AddKey(key, val)
		time.Sleep(sleep)
		c.AddKey(key, val)
		v, ok := c.TryGet(key)
		g.Expect(ok).To(BeFalse(), "Should not retrieve key after adding twice")
		g.Expect(v).To(BeEmpty(), "Should be empty string")
	})
}

// Test overall behaviour with access to different keys
func TestRetrieveExpiration(t *testing.T) {
	sleep := time.Millisecond * 10
	buffer := time.Millisecond * 100 // Experimentally determined
	g := NewWithT(t)

	// Generate some keys we will then store and retrieve
	keys := make([]string, numConsecutive*numParallel)
	for i := 0; i < len(keys); i++ {
		keys[i] = strconv.Itoa(rand.Int())
	}

	c := cache.New(sleep*numConsecutive + buffer)
	// Store data in cache over the interval it takes to do a rotation.
	repeatInParallel(func(i int) {
		c.AddKey(keys[i], strconv.Itoa(rand.Int()))
		time.Sleep(sleep)
	})

	// Most data will be retrieved from data2
	repeatInParallel(func(i int) {
		v, ok := c.TryGet(keys[i])
		g.Expect(ok).To(BeTrue(), "Should retrieve key after wait")
		g.Expect(v).ToNot(BeEmpty(), "Should not be empty string")
		time.Sleep(sleep)
	})
}

// Retrieving key after cache rotation should work
func TestRetrieveAfterRotation(t *testing.T) {
	g := NewWithT(t)
	c := cache.New(time.Second)
	c.AddKey("test1", "val1") // Expires at t == 1000
	time.Sleep(time.Millisecond * 800)
	c.AddKey("test2", "val2") // Expires at t == 1800

	time.Sleep(time.Millisecond * 100)
	_, ok := c.TryGet("test1") // From before rotation. t = 900
	g.Expect(ok).To(BeTrue())
	time.Sleep(time.Millisecond * 300)
	_, ok = c.TryGet("test2") // From after rotation, t = 1200
	g.Expect(ok).To(BeTrue())
}

func repeatInParallel(f func(int)) {
	wg := sync.WaitGroup{}
	for p := 0; p < numParallel; p++ {
		wg.Add(1)
		lp := p
		go func() {
			for j := 0; j < numConsecutive; j++ {
				f(lp*numConsecutive + j)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
