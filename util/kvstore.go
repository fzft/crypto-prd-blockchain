package util

type KeyValueStore[K comparable, V any] struct {
	data        map[K]V
	getCh       chan *GetRequest[K, V]
	setCh       chan *SetRequest[K, V]
	hasCh       chan *HasRequest[K]
	deleteCh    chan *DeleteRequest[K]
	hasOrAddCh  chan *HasOrAddRequest[K, V]
	lenCh       chan *LenRequest
	clearCh     chan *ClearRequest
	popAllValCh chan *PopAllValRequest[V]
	popAllKeyCh chan *PopAllKeyRequest[K]
	popValByKey chan *PopValByKeyRequest[K, V]
}

type PopValByKeyRequest[K comparable, V any] struct {
	key            K
	response       chan V
	existsResponse chan bool
}

type PopAllValRequest[V any] struct {
	response chan []V
}

type PopAllKeyRequest[K comparable] struct {
	response chan []K
}

type ClearRequest struct {
	done chan struct{}
}

type LenRequest struct {
	response chan int
}

type GetRequest[K comparable, V any] struct {
	key            K
	valueResponse  chan V
	existsResponse chan bool
}

type SetRequest[K comparable, V any] struct {
	key   K
	value V
}

type HasRequest[K comparable] struct {
	key      K
	response chan bool
}

type DeleteRequest[K comparable] struct {
	key K
}

type HasOrAddRequest[K comparable, V any] struct {
	key      K
	value    V
	response chan bool
}

func NewKeyValueStore[K comparable, V any]() *KeyValueStore[K, V] {
	store := &KeyValueStore[K, V]{
		data:        make(map[K]V),
		getCh:       make(chan *GetRequest[K, V]),
		setCh:       make(chan *SetRequest[K, V]),
		hasCh:       make(chan *HasRequest[K]),
		hasOrAddCh:  make(chan *HasOrAddRequest[K, V]),
		deleteCh:    make(chan *DeleteRequest[K]),
		lenCh:       make(chan *LenRequest),
		clearCh:     make(chan *ClearRequest),
		popAllValCh: make(chan *PopAllValRequest[V]),
		popAllKeyCh: make(chan *PopAllKeyRequest[K]),
		popValByKey: make(chan *PopValByKeyRequest[K, V]),
	}

	go store.run()
	return store
}

func (store *KeyValueStore[K, V]) run() {
	for {
		select {
		case req := <-store.getCh:
			value, exists := store.data[req.key]
			req.valueResponse <- value
			req.existsResponse <- exists
		case req := <-store.setCh:
			store.data[req.key] = req.value
		case req := <-store.hasCh:
			_, exists := store.data[req.key]
			req.response <- exists
		case req := <-store.hasOrAddCh:
			_, exists := store.data[req.key]
			if !exists {
				store.data[req.key] = req.value
			}
			req.response <- !exists
		case req := <-store.deleteCh:
			delete(store.data, req.key)
		case req := <-store.lenCh:
			req.response <- len(store.data)
		case req := <-store.clearCh:
			store.data = make(map[K]V)
			req.done <- struct{}{}
		case req := <-store.popAllValCh:
			var values []V
			for _, value := range store.data {
				values = append(values, value)
			}
			store.data = make(map[K]V)
			req.response <- values
		case req := <-store.popAllKeyCh:
			var keys []K
			for key := range store.data {
				keys = append(keys, key)
			}
			store.data = make(map[K]V)
			req.response <- keys
		case req := <-store.popValByKey:
			value, exists := store.data[req.key]
			if exists {
				delete(store.data, req.key)
			}
			req.response <- value
			req.existsResponse <- exists
		}
	}
}

func (store *KeyValueStore[K, V]) Get(key K) (V, bool) {
	req := &GetRequest[K, V]{
		key:            key,
		valueResponse:  make(chan V),
		existsResponse: make(chan bool),
	}
	store.getCh <- req
	value := <-req.valueResponse
	exists := <-req.existsResponse
	return value, exists
}

func (store *KeyValueStore[K, V]) Put(key K, value V) {
	req := &SetRequest[K, V]{
		key:   key,
		value: value,
	}
	store.setCh <- req
}

// Has returns true if the key exists.
func (store *KeyValueStore[K, V]) Has(key K) bool {
	req := &HasRequest[K]{
		key:      key,
		response: make(chan bool),
	}
	store.hasCh <- req
	return <-req.response
}

// HasOrAdd returns true if the key not exists and add the key-value pair. otherwise returns false.
func (store *KeyValueStore[K, V]) HasOrAdd(key K, value V) bool {
	req := &HasOrAddRequest[K, V]{
		key:      key,
		value:    value,
		response: make(chan bool),
	}
	store.hasOrAddCh <- req
	return <-req.response
}

func (store *KeyValueStore[K, V]) Delete(key K) {
	req := &DeleteRequest[K]{
		key: key,
	}
	store.deleteCh <- req
}

func (store *KeyValueStore[K, V]) Clear() {
	req := &ClearRequest{
		done: make(chan struct{}),
	}
	store.clearCh <- req
	<-req.done
}

func (store *KeyValueStore[K, V]) Len() int {
	req := &LenRequest{
		response: make(chan int),
	}
	store.lenCh <- req
	return <-req.response
}

// PopAllVal returns all values and clear the store.
func (store *KeyValueStore[K, V]) PopAllVal() []V {
	req := &PopAllValRequest[V]{
		response: make(chan []V),
	}
	store.popAllValCh <- req
	return <-req.response
}

// PopAllKey returns all keys and clear the store.
func (store *KeyValueStore[K, V]) PopAllKey() []K {
	req := &PopAllKeyRequest[K]{
		response: make(chan []K),
	}
	store.popAllKeyCh <- req
	return <-req.response
}

// PopValByKey returns the value of the key and delete the key-value pair.
func (store *KeyValueStore[K, V]) PopValByKey(key K) (V, bool) {
	req := &PopValByKeyRequest[K, V]{
		key:            key,
		response:       make(chan V),
		existsResponse: make(chan bool),
	}
	store.popValByKey <- req
	value := <-req.response
	exists := <-req.existsResponse
	return value, exists
}
