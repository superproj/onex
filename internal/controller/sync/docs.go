// Package sync is used to synchronize resources from the onex-apiserver to the MySQL database.

// When there is a large amount of data in Etcd, there is a query bottleneck, especially with the
// List interface. Therefore, the Sync package can be used to continuously synchronize data from Etcd
// to the MySQL database in real-time. When onex-gateway performs list queries, it can directly query
// from the MySQL database.
package sync
