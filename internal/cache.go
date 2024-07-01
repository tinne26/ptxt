package internal

import "fmt"
import "sync"

import "github.com/tinne26/ptxt/core"

import "github.com/tinne26/ggfnt"

// Default cache size value, in bytes.
const DefaultCacheSize = 8*1024*1024 // 8 MiB

// Package level cache. We could export the cache type and add
// methods for setting the cache on the renderer, just like etxt
// does, but caches are simpler on ptxt (no quantization variations,
// no size variations, no rasterizer variations), so I think having
// a few exposed package-level methods instead is good enough.
var DefaultCache Cache = *NewCache(DefaultCacheSize)

const noEntry32 uint32 = uint32(0b11111111_11111111_11111111_11111111)

// TODO: remove lru/mru *cachedMaskEntry and use cachedMaskEntries instead
// of pointers, and depend on mruIndex and lruIndex.

type CachedMaskEntry struct {
	Mask core.GlyphMask // Read-only.
	FontKey uint64 // Read-only.
	GlyphIndex ggfnt.GlyphIndex // Read-only.
	ByteSize uint32 // Read-only.
	PrevEntryIndex uint32 // for LRU. none if == noEntry32. also used for nextFreeEntryIndex
	NextEntryIndex uint32 // for LRU. none if == noEntry32
}

type Cache struct {
	masksMap map[[2]uint64]uint32 // key is made with the font UID and the glyph index at [4]
	maskEntries []*CachedMaskEntry
	mru *CachedMaskEntry
	lru *CachedMaskEntry
	mruIndex uint32
	lruIndex uint32
	nextFreeEntryIndex uint32 // none if == noEntry32

	mutex sync.RWMutex
	capacity uint64
	currentSize uint64
	peakSize uint64 // (max ever size)
}

func NewCache(capacity int) *Cache {
	const maxCapacity = 1*1024*1024*1024 // 1 GiB

	if capacity < 0 { panic("can't create cache with negative capacity") }
	if capacity > maxCapacity {
		capacity = maxCapacity
		fmt.Print("[ptxt.cache] Excessive cache capacity requested, limited to 1GiB\n")
	}
	return &Cache{
		capacity: uint64(capacity),
		masksMap: make(map[[2]uint64]uint32, 64),
		maskEntries: make([]*CachedMaskEntry, 0, 64),
		nextFreeEntryIndex: noEntry32,
		mruIndex: noEntry32,
	}
}

func (self *Cache) SetCapacity(bytes int) {
	if bytes < 0 { panic("can't cache.SetMaxSize(bytes) with bytes < 0") }
	self.mutex.Lock()
	if bytes == 0 {
		clear(self.masksMap)
		self.maskEntries = self.maskEntries[ : 0]
		self.mru, self.lru = nil, nil
		self.currentSize = 0
	} else {
		for self.currentSize > uint64(bytes) {
			self.removeOldestEntry()
		}
	}
	self.capacity = uint64(bytes)
	self.mutex.Unlock()
}

func (self *Cache) PeakSize() uint64 {
	self.mutex.RLock()
	peakSize := self.peakSize
	self.mutex.RUnlock()
	return peakSize
}

// Returns the number of cached masks currently in the cache.
func (self *Cache) NumEntries() int {
	self.mutex.RLock()
	numEntries := len(self.masksMap)
	self.mutex.RUnlock()
	return numEntries
}

func (self *Cache) SetGlyphMask(fontKey uint64, glyphIndex ggfnt.GlyphIndex, mask core.GlyphMask) {
	key := [2]uint64{fontKey, uint64(glyphIndex)}
	maskSize := glyphMaskByteSize(mask)

	self.mutex.Lock()
	entryIndex, found := self.masksMap[key]
	if found { // update existing mask case
		entry := self.maskEntries[entryIndex]

		// ensure free space
		if !self.getFreeSpace(uint64(maskSize), entry, entryIndex) {
			self.mutex.Unlock()
			return // can't fit mask into cache
		}
		self.currentSize += uint64(maskSize)

		// reassign mask and byte size
		entry.Mask = mask
		entry.ByteSize = maskSize

		// bump most recently used
		if entry != self.mru {
			if entry == self.lru {
				self.lruIndex = entry.NextEntryIndex
				self.lru = self.maskEntries[self.lruIndex]
			}
			entry.PrevEntryIndex = self.mruIndex
			entry.NextEntryIndex = noEntry32
			self.mru.NextEntryIndex = entryIndex
			self.mruIndex = entryIndex
			self.mru = entry
		}
	} else { // new mask case
		// ensure free space
		if !self.getFreeSpace(uint64(maskSize), nil, noEntry32) {
			self.mutex.Unlock()
			return // can't fit mask into cache
		}

		// create new entry
		entry := &CachedMaskEntry{
			Mask: mask,
			FontKey: fontKey,
			GlyphIndex: glyphIndex,
			ByteSize: maskSize,
			PrevEntryIndex: noEntry32,
			NextEntryIndex: noEntry32,
		}
		self.currentSize += uint64(entry.ByteSize)

		// assign entry to maskEntries and masksMap
		if self.nextFreeEntryIndex == noEntry32 {
			entryIndex = uint32(len(self.maskEntries))
			self.maskEntries = append(self.maskEntries, entry)
		} else {
			entryIndex = self.nextFreeEntryIndex
			self.nextFreeEntryIndex = self.maskEntries[self.nextFreeEntryIndex].PrevEntryIndex
		}
		self.masksMap[key] = entryIndex

		// set new entry as mru
		if self.mru == nil {
			self.lru, self.mru = entry, entry
			self.lruIndex, self.mruIndex = entryIndex, entryIndex
		} else {
			entry.PrevEntryIndex = self.mruIndex
			self.mru = entry
			self.mruIndex = entryIndex
		}
	}

	// update peak size if necessary
	if self.currentSize > self.peakSize {
		self.peakSize = self.currentSize
	}

	// unlock and we are done
	self.mutex.Unlock()
}

func (self *Cache) GetGlyphMask(fontKey uint64, glyphIndex ggfnt.GlyphIndex) (core.GlyphMask, bool) {
	var mask core.GlyphMask
	key := [2]uint64{fontKey, uint64(glyphIndex)}
	self.mutex.RLock()
	maskIndex, found := self.masksMap[key]
	if found { mask = self.maskEntries[maskIndex].Mask }
	self.mutex.RUnlock()
	return mask, found
}

// precondition: must be called with the cache locked
// additionally, if replacementEntry != nil and the result is true, the capacity
// will have already had the replacementEntry.ByteSize subtracted, so no need to
// do it manually.
func (self *Cache) getFreeSpace(requiredSpace uint64, replacementEntry *CachedMaskEntry, replacementEntryIndex uint32) bool {
	// trivial case
	if requiredSpace > self.capacity { return false }

	// detect replacement entry being considered
	if replacementEntry != nil {
		self.currentSize -= uint64(replacementEntry.ByteSize) // *don't restore later*
	}

	// iterate from lru deleting until we have enough space
	var restoreReplacementEntryAsLRU bool
	for self.currentSize + requiredSpace > self.capacity {
		if self.lru == replacementEntry {
			if self.lru.NextEntryIndex == noEntry32 { return true } // lru is mru and replacementEntry
			self.lruIndex = self.lru.NextEntryIndex
			self.lru = self.maskEntries[self.lru.NextEntryIndex]
			self.lru.PrevEntryIndex = noEntry32
			restoreReplacementEntryAsLRU = true
		} else {
			if uint64(self.lru.ByteSize) > self.currentSize {
				panic("broken code") // discretionary safety check
			}
			self.currentSize -= uint64(self.lru.ByteSize)
			key := [2]uint64{self.lru.FontKey, uint64(self.lru.GlyphIndex)}
			
			lruIndex, found := self.masksMap[key]
			if !found { panic("broken code") } // discretionary safety check
			delete(self.masksMap, key)
			self.lru.PrevEntryIndex = self.nextFreeEntryIndex
			self.lru.Mask = nil // allow mask to be GC'd
			self.nextFreeEntryIndex = lruIndex

			if self.lru != self.mru {
				self.lruIndex = self.lru.NextEntryIndex
				self.lru = self.maskEntries[self.lruIndex]
				self.lru.PrevEntryIndex = noEntry32
			} else { // self.lru == self.mru
				self.lru = nil
				self.mru = nil
				self.lruIndex = noEntry32
				self.mruIndex = noEntry32
				if self.currentSize + requiredSpace > self.capacity {
					panic("broken code") // discretionary safety check
				}
			}
		}
	}

	if restoreReplacementEntryAsLRU {
		replacementEntry.NextEntryIndex = self.lruIndex
		self.lru = replacementEntry
		self.lruIndex = replacementEntryIndex
		if self.mru == nil {
			self.mru = replacementEntry
			self.mruIndex = replacementEntryIndex
		}
	}

	return true
}

// Precondition: the cache is not empty.
// If there's nothing to remove, this method will panic.
func (self *Cache) removeOldestEntry() {
	// optional safety check
	if self.lru.PrevEntryIndex != noEntry32 {
		panic("broken code")
	}
	
	// delete oldest
	delete(self.masksMap, [2]uint64{self.lru.FontKey, uint64(self.lru.GlyphIndex)})

	// re-link LRU
	if self.lru.NextEntryIndex != noEntry32 {
		self.lru = self.maskEntries[self.lru.NextEntryIndex]
	} else {
		if self.mru != nil { panic("broken code") } // optional safety check
		if self.mru == self.lru { self.mru = nil }
		self.lru = nil
	}
}
