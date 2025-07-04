package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	elastic "github.com/elastic/go-elasticsearch/v9"
)

type searchClient struct {
	*elastic.Client
}

func connectElasticsearch() searchClient {
	address := fmt.Sprintf("http://%s:%d", Config.Elasticsearch.Host, Config.Elasticsearch.Port)
	log.Printf("Connecting to Elasticsearch at %s\n", address)

	cfg := elastic.Config{
		Addresses: []string{address},
	}

	if Config.Elasticsearch.SecureConnection {
		cfg.Addresses = []string{strings.Replace(address, "http://", "https://", 1)}
	}

	if Config.Elasticsearch.Username != BLANK && Config.Elasticsearch.Password != BLANK {
		cfg.Username = Config.Elasticsearch.Username
		cfg.Password = Config.Elasticsearch.Password
	}

	client, err := elastic.NewClient(cfg)
	if err != nil {
		log.Fatalf("Cannot connect to Elasticsearch: %v", err)
	}

	// Check if Elasticsearch is running
	res, err := client.Info()
	if err != nil {
		log.Fatalf("Elasticsearch is not running: %v", err)
	}
	defer res.Body.Close()

	var info map[string]any
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		log.Fatalf("Error parsing the response body: %v", err)
	}

	log.Printf("Elasticsearch returned with code %d and version %s", res.StatusCode, info["version"].(map[string]any)["number"])

	return searchClient{
		Client: client,
	}
}

// CreateIndex creates a new index with the given name and mapping
func CreateSearchIndex(ctx Context, indexName string, mapping string) Error {
	res, err := esClient.Indices.Exists([]string{indexName})
	if err != nil {
		ctx.LogError("Failed to check if index exists: %v", err)
		return NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to check if index exists: %v", err))
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		res, err := esClient.Indices.Create(
			indexName,
			esClient.Indices.Create.WithBody(strings.NewReader(mapping)),
		)
		if err != nil {
			ctx.LogError("Failed to create index: %v", err)
			return NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to create index: %v", err))
		}
		defer res.Body.Close()

		if res.IsError() {
			ctx.LogError("Failed to create index: %s", res.String())
			return NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to create index: %s", res.String()))
		}
	}

	ctx.LogInfo("Index %s is created", indexName)
	return nil
}

// IndexDocument indexes a document in Elasticsearch
func IndexSearchDocument(ctx Context, indexName string, id string, document any) Error {
	data, err := json.Marshal(document)
	if err != nil {
		return NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to marshal document: %v", err))
	}

	res, err := esClient.Index(
		indexName,
		bytes.NewReader(data),
		esClient.Index.WithDocumentID(id),
		esClient.Index.WithContext(ctx),
	)
	if err != nil {
		return NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to index document: %v", err))
	}
	defer res.Body.Close()

	if res.IsError() {
		return NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to index document: %s", res.String()))
	}
	return nil
}

// Search performs a search query on the specified index
func Search(ctx Context, indexName string, query any) (map[string]any, Error) {
	data, err := json.Marshal(query)
	if err != nil {
		return nil, NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to marshal query: %v", err))
	}

	res, err := esClient.Search(
		esClient.Search.WithIndex(indexName),
		esClient.Search.WithBody(bytes.NewReader(data)),
		esClient.Search.WithContext(ctx),
	)
	if err != nil {
		return nil, NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to search: %v", err))
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Search error: %s", res.String()))
	}

	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to parse response: %v", err))
	}

	return result, nil
}

// DeleteSearchDocument deletes a document from the index
func DeleteSearchDocument(ctx Context, indexName string, id string) Error {
	res, err := esClient.Delete(
		indexName,
		id,
		esClient.Delete.WithContext(ctx),
	)
	if err != nil {
		return NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to delete document: %v", err))
	}
	defer res.Body.Close()

	if res.IsError() {
		return NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to delete document: %s", res.String()))
	}
	return nil
}

// UpdateSearchDocument updates a document in the index
func UpdateSearchDocument(ctx Context, indexName string, id string, update any) Error {
	data, err := json.Marshal(map[string]any{
		"doc": update,
	})
	if err != nil {
		return NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to marshal update: %v", err))
	}

	res, err := esClient.Update(
		indexName,
		id,
		bytes.NewReader(data),
		esClient.Update.WithContext(ctx),
	)
	if err != nil {
		return NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to update document: %v", err))
	}
	defer res.Body.Close()

	if res.IsError() {
		return NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to update document: %s", res.String()))
	}
	return nil
}

// Close closes the Elasticsearch client
func (c *searchClient) Close() {
	// The official client doesn't require explicit closing
	// It uses Go's standard http.Client which manages its own connections
}

// SearchIndexExists checks if an index exists in Elasticsearch
func SearchIndexExists(ctx Context, indexName string) bool {
	res, err := esClient.Indices.Exists([]string{indexName})
	if err != nil {
		ctx.LogError("Failed to check if index exists: %v", err)
		return false
	}
	defer res.Body.Close()

	return res.StatusCode == 200
}

// AppendSearchDocument appends a document to an index, creating the index if it doesn't exist
func AppendSearchDocument(ctx Context, indexName string, id string, document any) Error {
	// Check if index exists
	if !SearchIndexExists(ctx, indexName) {
		// Create index with default mapping if it doesn't exist
		defaultMapping := `{
			"mappings": {
				"properties": {
					"@timestamp": { "type": "date" }
				}
			}
		}`
		if err := CreateSearchIndex(ctx, indexName, defaultMapping); err != nil {
			ctx.LogError("Failed to create index %s: %v", indexName, err)
			return err
		}
	}

	// Add timestamp to document if it's a map
	if docMap, ok := document.(map[string]any); ok {
		docMap["@timestamp"] = time.Now().UTC().Format(time.RFC3339)
		document = docMap
	}

	// Index the document
	return IndexSearchDocument(ctx, indexName, id, document)
}
