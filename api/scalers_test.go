package api

import (
	"encoding/json"
	"testing"

	"github.com/Zenika/zscaler/core"
	"github.com/stretchr/testify/assert"
)

func configAndBuildScaler(t *testing.T, input string) error {
	var sb ScalerBuilder
	core.Config = MockConf
	err := json.Unmarshal([]byte(input), &sb)
	failIfErr(t, err)
	_, err = sb.Build()
	return err
}

func TestCreateComposeScaler(t *testing.T) {
	const input = `{
    "type":"docker-compose",
    "name":"testing",
    "args": {
        "service":"whoami",
        "config":"/dummy/path"
    }
}`
	err := configAndBuildScaler(t, input)
	assert.Nil(t, err)
}

func TestCreateComposeScalerBadType(t *testing.T) {
	const input = `{
    "type":"badtype",
    "name":"testing",
    "args": {
        "service":"whoami",
        "config":"/dummy/path"
    }
}`
	err := configAndBuildScaler(t, input)
	assert.Error(t, err)
}
