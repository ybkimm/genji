module github.com/asdine/genji/engine/badgerengine

go 1.13

require (
	github.com/asdine/genji v0.2.2
	github.com/dgraph-io/badger/v2 v2.0.0
	github.com/stretchr/testify v1.4.0
	go.etcd.io/bbolt v1.3.3 // indirect
)

replace github.com/asdine/genji v0.2.2 => ../../
