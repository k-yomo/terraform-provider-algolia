package algoliautil

import (
	"fmt"
)

func IndexExistsInReplicas(replicas []string, indexName string, isVirtual bool) bool {
	replicaIndexName := getReplicaIndexName(indexName, isVirtual)
	for _, replica := range replicas {
		if replica == replicaIndexName {
			return true
		}
	}
	return false
}

func RemoveIndexFromReplicas(replicas []string, indexName string, isVirtual bool) []string {
	replicaIndexName := getReplicaIndexName(indexName, isVirtual)

	var newReplicas []string
	for _, replica := range replicas {
		if replica == replicaIndexName {
			continue
		}
		newReplicas = append(newReplicas, replica)
	}
	return newReplicas
}

func getReplicaIndexName(indexName string, isVirtual bool) string {
	if isVirtual {
		return fmt.Sprintf("virtual(%s)", indexName)
	}
	return indexName
}
