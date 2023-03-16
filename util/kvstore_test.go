package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewKeyValueStore(t *testing.T) {
	store := NewKeyValueStore[int, string]()
	store.Put(1, "one")
	store.Put(2, "two")
	store.Put(3, "three")
}

func TestKeyValueStoreHas(t *testing.T) {
	store := NewKeyValueStore[int, string]()
	store.Put(1, "one")
	store.Put(2, "two")
	store.Put(3, "three")

	assert.True(t, store.Has(1))
	assert.True(t, store.Has(2))
	assert.True(t, store.Has(3))
}

func TestKeyValueStoreHasOrAdd(t *testing.T) {
	store := NewKeyValueStore[int, string]()
	store.Put(1, "one")
	store.Put(2, "two")
	store.Put(3, "three")

	assert.True(t, !store.HasOrAdd(1, "one"))
	assert.True(t, !store.HasOrAdd(2, "two"))
	assert.True(t, !store.HasOrAdd(3, "three"))

	assert.False(t, !store.HasOrAdd(4, "four"))
	val, exist := store.Get(4)
	assert.True(t, exist)
	assert.Equal(t, "four", val)
}

func TestKeyValueStoreDelete(t *testing.T) {
	store := NewKeyValueStore[int, string]()
	store.Put(1, "one")
	store.Put(2, "two")
	store.Put(3, "three")

	store.Delete(1)
	store.Delete(2)
	store.Delete(3)

	assert.True(t, !store.Has(1))
	assert.True(t, !store.Has(2))
	assert.True(t, !store.Has(3))
}

func TestKeyValueStoreLen(t *testing.T) {
	store := NewKeyValueStore[int, string]()
	store.Put(1, "one")
	store.Put(2, "two")
	store.Put(3, "three")

	assert.Equal(t, 3, store.Len())
}

func TestKeyValueStoreClear(t *testing.T) {
	store := NewKeyValueStore[int, string]()
	store.Put(1, "one")
	store.Put(2, "two")
	store.Put(3, "three")

	store.Clear()
	assert.Equal(t, 0, store.Len())
}

func TestKeyValueStorePopAllKey(t *testing.T) {
	store := NewKeyValueStore[int, string]()
	store.Put(1, "one")
	store.Put(2, "two")
	store.Put(3, "three")

	keys := store.PopAllKey()
	assert.Equal(t, 3, len(keys))
	assert.Equal(t, 0, store.Len())
}

func TestKeyValueStorePopAllValue(t *testing.T) {
	store := NewKeyValueStore[int, string]()
	store.Put(1, "one")
	store.Put(2, "two")
	store.Put(3, "three")

	values := store.PopAllVal()
	assert.Equal(t, 3, len(values))
	assert.Equal(t, 0, store.Len())
}

func TestKeyValueStorePopValByKey(t *testing.T) {
	store := NewKeyValueStore[int, string]()
	store.Put(1, "one")
	store.Put(2, "two")
	store.Put(3, "three")

	val, exist := store.PopValByKey(1)
	assert.True(t, exist)
	assert.Equal(t, "one", val)
	assert.Equal(t, 2, store.Len())
}
