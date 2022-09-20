package elastic

import (
	_ "embed"
)

//*********************** hard-coded indices mapping used if the server is created without WithIndicesMapping option ***********************
//go:embed mock_settings/mapping.json
var indicesMappingsBytes []byte

////*********************** Responses for ES initialization ***********************
//go:embed mock_settings/clusterSettingsResp.json
var clusterSettingsResp []byte

//go:embed mock_settings/rootResponse.json
var rootResponse []byte

//go:embed mock_settings/ackResponse.json
var ackResponse []byte
