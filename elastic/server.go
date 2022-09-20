package elastic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/armosec/ca-test/server"
)

type ElasticServer interface {
	server.TestServer
	GetIndexSuffix() string
}

type elasticServer struct {
	server.TestServer
	options options
}

func NewElastic(t *testing.T, esOpts []ElasticServerOption, opts ...server.ServerOption) ElasticServer {
	esOptions, err := makeOptions(esOpts...)
	if err != nil {
		t.Fatal(err)
	}

	return &elasticServer{
		TestServer: createServer(t, *esOptions, opts...),
		options:    *esOptions,
	}
}

func createServer(t *testing.T, esOpts options, opts ...server.ServerOption) server.TestServer {

	commonOptions := []server.ServerOption{
		server.WithHeaders(map[string]string{
			"X-Elastic-Product": "Elasticsearch",
		}),
		server.WithBuiltInHandler(
			server.WithMethod(http.MethodGet),
			server.WithPath("/"),
			server.WithResponse(rootResponse),
		),
		server.WithBuiltInHandler(
			server.WithMethod(http.MethodPut),
			server.WithPath("/_cluster/settings"),
			server.WithResponse(clusterSettingsResp),
		),
		server.WithBuiltInHandler(
			server.WithMethod(http.MethodPut),
			server.WithPath("/*/_settings"),
			server.WithResponse(ackResponse),
		),
	}

	//add custom options
	if opts != nil {
		commonOptions = append(commonOptions, opts...)
	}

	IndicesOptions := addIndicesHandlers(t, esOpts)

	//create and start the mock server
	elasticMock, err := server.NewTestServer(append(commonOptions, IndicesOptions...)...)
	if err != nil {
		t.Fatal(err)
	}
	return elasticMock

}

func (es *elasticServer) GetIndexSuffix() string {
	return es.options.indexSuffix
}

func addIndicesHandlers(t *testing.T, esOpts options) []server.ServerOption {
	var indicesMapping map[string]string
	if len(esOpts.indexPrefix2Mapping) > 0 {
		indicesMapping = esOpts.indexPrefix2Mapping
	} else {
		indicesMapping = make(map[string]string)
		if err := json.Unmarshal(indicesMappingsBytes, &indicesMapping); err != nil {
			t.Fatal(err)
		}
	}

	indexCatResponse := func(indexName string) []byte {
		indexCat := &struct {
			Index              string `json:"index"`
			CreationDateString string `json:"creation.date.string"`
		}{
			Index:              indexName + esOpts.indexSuffix,
			CreationDateString: "2022-09-05T14:16:48.598Z",
		}
		resp := []interface{}{indexCat}
		respBytes, _ := json.Marshal(resp)
		return respBytes
	}
	indexMappingResponse := func(indexName string, indexMapping string) []byte {
		mappingObj := interface{}(nil)
		json.Unmarshal([]byte(indexMapping), &mappingObj)
		mapping := make(map[string]interface{})
		mapping[indexName] = mappingObj
		respBytes, _ := json.Marshal(mapping)
		return respBytes
	}
	options := []server.ServerOption{}

	//create index mapping handlers for supported indices
	for indexName, mapping := range indicesMapping {
		indexName := indexName
		mapping := mapping
		t.Log("load mapping for index:", indexName)
		options = append(options,
			server.WithBuiltInHandler(
				server.WithMethod(http.MethodGet),
				server.WithPath(fmt.Sprintf("/_cat/indices/%s*", indexName)),
				server.WithResponse(indexCatResponse(indexName)),
			))

		options = append(options,
			server.WithBuiltInHandler(
				server.WithMethod(http.MethodGet),
				server.WithPath(fmt.Sprintf("/%s%s/_mapping", indexName, esOpts.indexSuffix)),
				server.WithResponse(indexMappingResponse(indexName+esOpts.indexSuffix, mapping)),
			))
	}
	return options
}
