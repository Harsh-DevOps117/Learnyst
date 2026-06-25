package cassandra

import (
	"log"
	"time"

	"github.com/gocql/gocql"
)

var Session *gocql.Session

func InitCassandra() {
	// In production, fetch from env variable CASSANDRA_HOSTS
	cluster := gocql.NewCluster("cassandra") // Service name in docker-compose
	cluster.Keyspace = "lms_analytics"
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 10 * time.Second
	cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 3}

	var err error
	// Attempt to connect. Note: We might need to create the keyspace first if it doesn't exist
	// For resilience, let's connect to system keyspace, create keyspace, then connect to lms_analytics
	sysCluster := gocql.NewCluster("cassandra")
	sysCluster.Keyspace = "system"
	sysCluster.Timeout = 10 * time.Second

	sysSession, err := sysCluster.CreateSession()
	if err != nil {
		log.Printf("Failed to connect to Cassandra: %v", err)
		return
	}
	defer sysSession.Close()

	err = sysSession.Query(`
		CREATE KEYSPACE IF NOT EXISTS lms_analytics 
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
	`).Exec()
	if err != nil {
		log.Printf("Failed to create keyspace: %v", err)
	}

	Session, err = cluster.CreateSession()
	if err != nil {
		log.Printf("Failed to connect to Cassandra keyspace: %v", err)
		return
	}

	// Create Video Progress Table
	err = Session.Query(`
		CREATE TABLE IF NOT EXISTS video_progress (
			tenant_id int,
			video_id int,
			student_id uuid,
			timestamp timestamp,
			watch_seconds int,
			PRIMARY KEY ((tenant_id, video_id), student_id, timestamp)
		) WITH CLUSTERING ORDER BY (student_id ASC, timestamp DESC);
	`).Exec()

	if err != nil {
		log.Printf("Failed to create video_progress table: %v", err)
	} else {
		log.Println("Cassandra Initialized and video_progress table ready!")
	}
}
