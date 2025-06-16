package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

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
func (c *searchClient) CreateIndex(ctx Context, indexName string, mapping string) Error {
	res, err := c.Client.Indices.Exists([]string{indexName})
	if err != nil {
		return NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to check if index exists: %v", err))
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		res, err := c.Client.Indices.Create(
			indexName,
			c.Client.Indices.Create.WithBody(strings.NewReader(mapping)),
		)
		if err != nil {
			return NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to create index: %v", err))
		}
		defer res.Body.Close()

		if res.IsError() {
			return NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to create index: %s", res.String()))
		}
	}
	return nil
}

// IndexDocument indexes a document in Elasticsearch
func (c *searchClient) IndexDocument(ctx Context, indexName string, id string, document any) Error {
	data, err := json.Marshal(document)
	if err != nil {
		return NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to marshal document: %v", err))
	}

	res, err := c.Client.Index(
		indexName,
		bytes.NewReader(data),
		c.Client.Index.WithDocumentID(id),
		c.Client.Index.WithContext(ctx),
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
func (c *searchClient) Search(ctx Context, indexName string, query any) (map[string]any, Error) {
	data, err := json.Marshal(query)
	if err != nil {
		return nil, NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to marshal query: %v", err))
	}

	res, err := c.Client.Search(
		c.Client.Search.WithIndex(indexName),
		c.Client.Search.WithBody(bytes.NewReader(data)),
		c.Client.Search.WithContext(ctx),
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

// DeleteDocument deletes a document from the index
func (c *searchClient) DeleteDocument(ctx Context, indexName string, id string) Error {
	res, err := c.Client.Delete(
		indexName,
		id,
		c.Client.Delete.WithContext(ctx),
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

// UpdateDocument updates a document in the index
func (c *searchClient) UpdateDocument(ctx Context, indexName string, id string, update any) Error {
	data, err := json.Marshal(map[string]any{
		"doc": update,
	})
	if err != nil {
		return NewError(ERROR_CODE_FROM_ELASTICSEARCH, fmt.Sprintf("Failed to marshal update: %v", err))
	}

	res, err := c.Client.Update(
		indexName,
		id,
		bytes.NewReader(data),
		c.Client.Update.WithContext(ctx),
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

// IndexExists checks if an index exists in Elasticsearch
func (c *searchClient) IndexExists(ctx Context, indexName string) bool {
	res, err := c.Client.Indices.Exists([]string{indexName})
	if err != nil {
		ctx.LogError("Failed to check if index exists: %v", err)
		return false
	}
	defer res.Body.Close()

	return res.StatusCode == 200
}
