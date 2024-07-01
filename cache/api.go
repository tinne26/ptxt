package cache

import "github.com/tinne26/ptxt/internal"

// Default cache size value, in bytes.
const DefaultSize = 8*1024*1024 // 8 MiB

// cache size constant verification
func init() {
	if DefaultSize != internal.DefaultCacheSize {
		panic("DefaultSize != internal.DefaultCacheSize")
	}
}

// Returns the current cache capacity. It's either [DefaultSize] or
// the last value set by the user through [SetCapacity]().
func GetCapacity() int {
	return internal.DefaultCache.Capacity()
}

// Sets the maximum cache size, in bytes. The default value is [DefaultSize].
// Values above 1GiB are not allowed. Values below 32KiB are not recommended.
// 
// To fully clear the cache, you can set the capacity to zero and then bring
// it up again. That being said, the cache will automatically evict entries
// with an LRU policy as needed, so you rarely need to clear anything manually.
func SetCapacity(bytes int) {
	internal.DefaultCache.SetCapacity(bytes)
}

// Returns an approximation of the number of bytes taken by the glyph masks
// currently stored in the cache.
//
// In Ebitengine this estimation is not particularly reliable, as images
// might or might not include borders, mipmaps, and their internal structure
// might change between versions, causing more or less overhead.
func GetCurrentSize() int {
	return internal.DefaultCache.CurrentSize()
}

// Returns an approximation of the maximum amount of bytes that the cache
// has been filled with at any point of its life.
// 
// This method can be useful to determine the actual cache usage within
// your application and set its capacity to a reasonable value.
func GetPeakSize() int {
	return int(internal.DefaultCache.PeakSize())
}

// Returns the number of mask entries currently cached. Entries have an
// overhead of about 32 bytes each, which someone obsessive enough might
// be interested in analyzing in more detail. If you are that someone, just
// write your own package already.
func GetNumEntries() int {
	return internal.DefaultCache.NumEntries()
}
