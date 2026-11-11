package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
	"github.com/hashicorp/terraform-provider-algolia/internal/algoliautil"
	"golang.org/x/sync/errgroup"
)

func main() {
	appID := os.Getenv("ALGOLIA_APP_ID")
	apiKey := os.Getenv("ALGOLIA_API_KEY")

	log.Printf("[START] Deletes All indices with prefix '%s' in appID: %s", algoliautil.TestIndexNamePrefix, appID)
	algoliaClient, err := search.NewClient(appID, apiKey)
	if err != nil {
		log.Fatal(err)
	}
	listIndicesRes, err := algoliaClient.ListIndices(algoliaClient.NewApiListIndicesRequest())
	if err != nil {
		log.Fatal("Failed to list indices")
	}

	eg := errgroup.Group{}
	for _, index := range listIndicesRes.Items {
		if !strings.HasPrefix(index.Name, algoliautil.TestIndexNamePrefix) {
			continue
		}
		eg.Go(func() error {
			res, err := algoliaClient.DeleteIndex(algoliaClient.NewApiDeleteIndexRequest(index.Name))
			if err != nil {
				return fmt.Errorf("failed to delete %s: %w", index.Name, err)
			}
			_, err = algoliaClient.WaitForTask(index.Name, res.TaskID)
			if err != nil {
				return fmt.Errorf("failed to delete %s: %w", index.Name, err)
			}

			log.Printf("[INFO] Index '%s' is deleted", index.Name)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}

	listAPIKeysRes, err := algoliaClient.ListApiKeys()
	if err != nil {
		log.Fatal("Failed to list api keys")
	}
	for _, apiKey := range listAPIKeysRes.Keys {
		isTestAPIKey := slices.ContainsFunc(apiKey.Indexes, func(index string) bool {
			return strings.HasPrefix(index, algoliautil.TestIndexNamePrefix)
		})
		if !isTestAPIKey {
			continue
		}
		eg.Go(func() error {
			_, err := algoliaClient.DeleteApiKey(algoliaClient.NewApiDeleteApiKeyRequest(apiKey.Value))
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
