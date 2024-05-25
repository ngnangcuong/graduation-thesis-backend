package storage

import (
	"strings"

	"github.com/gocql/gocql"
)

var cassandraClient *gocql.Session

func getSession(hosts []string, keyspace string) *gocql.Session {
	cluster := gocql.NewCluster(strings.Join(hosts, ","))
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}

	return session
}

func GetSession(hosts []string, keyspace string) *gocql.Session {
	if cassandraClient == nil {
		cassandraClient = getSession(hosts, keyspace)
	}

	return cassandraClient
}
