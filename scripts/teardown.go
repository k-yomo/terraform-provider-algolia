package main

import (
	"fmt"
	"log"
	"os"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"golang.org/x/sync/errgroup"
)

func main() {
	appID := os.Getenv("ALGOLIA_APP_ID")
	apiKey := os.Getenv("ALGOLIA_API_KEY")

	log.Printf("[START] Deletes All indices, appID: %s", appID)
	algoliaClient := search.NewClient(appID, apiKey)
	res, err := algoliaClient.ListIndices()
	if err != nil {
		log.Fatal("Failed to list indices")
	}

	eg := errgroup.Group{}
	for _, index := range res.Items {
		index := index
		eg.Go(func() error {
			res, err := algoliaClient.InitIndex(index.Name).Delete()
			if err != nil {
				return fmt.Errorf("failed to delete %s", index.Name)
			}
			if err := res.Wait(); err != nil {
				return fmt.Errorf("failed to delete %s: %w", index.Name, err)
			}

			log.Printf("[INFO] Index '%s' is deleted", index.Name)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
	log.Println("[END] All indices are deleted")
}
