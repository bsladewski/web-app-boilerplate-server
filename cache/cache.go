package cache

import (
	"container/heap"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// localCache is used to cache data in application memory.
var localCache = struct {
	mutex   *sync.Mutex
	entries map[string]*cacheEntry
	heap    *priorityQueue
}{
	mutex:   &sync.Mutex{},
	entries: map[string]*cacheEntry{},
	heap:    &priorityQueue{},
}

// cacheEntry stores an item in the cache along with an expiration time. When an
// operation is performed on the cache any expired records will be removed from
// the cache.
type cacheEntry struct {
	key     string
	expires time.Time
	item    interface{}
}

// String returns a string representation of this cache entry.
func (c cacheEntry) String() string {
	return fmt.Sprintf("Key: %s Expires: %v Item: %v", c.key, c.expires, c.item)
}

// SetLocal adds an item to the local cache.
func SetLocal(key string, item interface{}, ttl time.Duration) {

	// lock access to the local cache to prevent concurrent access
	localCache.mutex.Lock()
	defer localCache.mutex.Unlock()

	// remove stale items from the cache
	removeStaleLocal()

	// if the item is present in the cache update it
	if entry, ok := localCache.entries[key]; ok {
		entry.expires = time.Now().Add(ttl)
		entry.item = item
		heap.Init(localCache.heap)
		return
	}

	entry := &cacheEntry{
		key:     key,
		expires: time.Now().Add(ttl),
		item:    item,
	}

	// add the item to the cache
	logrus.Debugf("new cache entry: %v", *entry)
	localCache.entries[key] = entry
	heap.Push(localCache.heap, entry)

}

// GetLocal attempts to retrieve an item from the local cache returning the item
// associated with the supplied key if available and a flag that indicates
// whether the item was found.
func GetLocal(key string) (interface{}, bool) {

	// lock access to the local cache to prevent concurrent access
	localCache.mutex.Lock()
	defer localCache.mutex.Unlock()

	// remove stale items from the cache
	removeStaleLocal()

	// check if the item is present in the cache
	if entry, ok := localCache.entries[key]; ok {
		logrus.Debugf("get cache entry: %v", *entry)
		return entry.item, true
	}

	return nil, false
}

// removeStaleLocal removes any stale entries from the local cache.
func removeStaleLocal() {
	for {
		entry, ok := localCache.heap.Peek().(*cacheEntry)
		if ok && entry.expires.Before(time.Now()) {
			logrus.Debugf("remove cache entry: %v", *entry)
			delete(localCache.entries, entry.key)
			localCache.heap.Pop()
			continue
		}
		break
	}
}

// priorityQueue is used to store cache entries in a way that is optimized for
// removing stale entries.
type priorityQueue []*cacheEntry

// Len gets the current length of the priority queue.
func (p priorityQueue) Len() int {
	return len(p)
}

// Less returns whether the entry at index i expires before the entry at index j.
func (p priorityQueue) Less(i, j int) bool {
	if i < 0 || i >= len(p) || j < 0 || j >= len(p) {
		return false
	}

	return p[i].expires.Before(p[j].expires)
}

// Peek returns the entry with the lowest expiration time without removing it
// from the priority queue.
func (p *priorityQueue) Peek() interface{} {
	if len(*p) == 0 {
		return nil
	}

	return (*p)[len(*p)-1]
}

// Pop removes and returns the entry with the lowest expiration time.
func (p *priorityQueue) Pop() interface{} {
	if len(*p) == 0 {
		return nil
	}

	item := (*p)[len(*p)-1]
	*p = (*p)[0 : len(*p)-1]
	return item
}

// Push adds a new element to the priority queue.
func (p *priorityQueue) Push(x interface{}) {
	entry, ok := x.(*cacheEntry)
	if !ok {
		logrus.Errorf("expected type *cacheEntry, got type %T: %v", x, x)
		return
	}

	*p = append(*p, entry)
}

// Swap exchanges the values at indices i and j.
func (p priorityQueue) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
