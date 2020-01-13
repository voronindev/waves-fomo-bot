package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type ContractProps struct {
	BlocksOnGameStart    uint64 `json:"$blocksOnGameStart"`
	BlocksPerCompetition uint64 `json:"$blocksPerCompetition"`
	BlocksPerRound       uint64 `json:"$blocksPerRound"`
	WithdrawPeriod       uint64 `json:"$withdrawPeriod"`
	MaxRounds            uint64 `json:"$maxRounds"`
	PaymentAmount        uint64 `json:"$pmtStep"`
	HeightStep           uint64 `json:"$heightStep"`
	Asset                string `json:"$mrtAssetId"`
}

func GetContractProperties() (*ContractProps, error) {
	client, err := http.Get(fmt.Sprintf("%s/addresses/data/%s?matches=\\$.*", apiUrlNode, contractAddress))
	if err != nil {
		return nil, err
	}
	defer client.Body.Close()

	body, err := ioutil.ReadAll(client.Body)
	if err != nil {
		return nil, err
	}

	var raw []map[string]string
	err = json.Unmarshal(body, &raw)
	if err != nil {
		return nil, err
	}
	mapped := make(map[string]interface{}, len(raw))

	for _, prop := range raw {
		value, err := strconv.Atoi(prop["value"])
		if err == nil {
			mapped[prop["key"]] = value
		} else {
			mapped[prop["key"]] = prop["value"]
		}
	}

	body, err = json.Marshal(mapped)
	if err != nil {
		return nil, err
	}

	props := ContractProps{}
	err = json.Unmarshal(body, &props)
	if err != nil {
		return nil, err
	}

	return &props, nil
}
