package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/hashicorp/terraform-provider-algolia/internal/algoliautil"
	"golang.org/x/sync/errgroup"
)

func main() {
	appID := os.Getenv("ALGOLIA_APP_ID")
	apiKey := os.Getenv("ALGOLIA_API_KEY")

	log.Printf("[START] Deletes All indices with prefix '%s' in appID: %s", algoliautil.TestIndexNamePrefix, appID)
	algoliaClient := search.NewClient(appID, apiKey)
	listIndicesRes, err := algoliaClient.ListIndices()
	if err != nil {
		log.Fatal("Failed to list indices")
	}

	eg := errgroup.Group{}
	for _, index := range listIndicesRes.Items {
		if !strings.HasPrefix(index.Name, algoliautil.TestIndexNamePrefix) {
			continue
		}
		index := index
		eg.Go(func() error {
			res, err := algoliaClient.InitIndex(index.Name).Delete()
			if err != nil {
				return fmt.Errorf("failed to delete %s: %w", index.Name, err)
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

	listAPIKeysRes, err := algoliaClient.ListAPIKeys()
	if err != nil {
		log.Fatal("Failed to list api keys")
	}
	for _, apiKey := range listAPIKeysRes.Keys {
		apiKey := apiKey
		isTestAPIKey := slices.ContainsFunc(apiKey.Indexes, func(index string) bool {
			return strings.HasPrefix(index, algoliautil.TestIndexNamePrefix)
		})
		if !isTestAPIKey {
			continue
		}
		eg.Go(func() error {
			_, err := algoliaClient.DeleteAPIKey(apiKey.Value)
			if err != nil {
				return fmt.Errorf("failed to delete %s: %w", apiKey.Value, err)
			}
			log.Printf("[INFO] API key '%s' is deleted", apiKey.Value)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}

	log.Println("[END] All indices and API keys are deleted")
}
