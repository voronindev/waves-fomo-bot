package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/wavesplatform/gowaves/pkg/client"
	"github.com/wavesplatform/gowaves/pkg/proto"
)

var defaultOptions = client.Options{
	BaseUrl: "https://nodes.wavesnodes.com",
	Client:  &http.Client{Timeout: 3 * time.Second},
}

var tradisysOptions = client.Options{
	BaseUrl: "https://gamenode.tradisys.com",
	Client:  &http.Client{Timeout: 3 * time.Second},
}

type Listener struct {
	api   []*client.Transactions
	props *ContractProps
}

func NewListener(props *ContractProps) *Listener {
	api := []*client.Transactions{
		client.NewTransactions(defaultOptions),
		client.NewTransactions(tradisysOptions),
	}

	return &Listener{api: api, props: props}
}

func (s *Listener) Launch() {
	height := uint64(0)

	for {
		h := s.getHeight()
		if h >= height {
			height = h
		}
		offset := (height - s.props.BlocksOnGameStart) % s.props.BlocksPerRound
		fmt.Println(s.props.BlocksPerCompetition-offset, "left")

		if s.props.BlocksPerCompetition-offset == 1 {
			fmt.Println("TURN")
			s.turn()
		}

		time.Sleep(1 * time.Second)
	}
}

func (s *Listener) getHeight() uint64 {
	client, err := http.Get(fmt.Sprintf("%s/blocks/height", apiUrlNode))
	if err != nil {
		return 0
	}
	defer client.Body.Close()

	body, err := ioutil.ReadAll(client.Body)
	if err != nil {
		return 0
	}

	h := &struct {
		Height uint64 `json:"height"`
	}{}

	_ = json.Unmarshal(body, h)

	return h.Height
}

func (s *Listener) turn() {
	inv := proto.NewUnsignedInvokeScriptV1(
		proto.MainNetScheme,
		senderPublicKey,
		recipient,
		functionCall,
		payments,
		*assetWaves,
		500000,
		uint64(time.Now().Unix()*1000),
	)

	err := inv.Sign(senderPrivateKey)
	fmt.Println(err)

	for _, api := range s.api {
		_, _ = api.Broadcast(context.Background(), inv)
	}

	fmt.Println(err)
}
