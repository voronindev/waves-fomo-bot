package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
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

type GameState struct {
	heightToGetMoney uint64
	last             string
}

type Listener struct {
	api   []*client.Transactions
	state *GameState
	props *ContractProps
}

func NewListener(props *ContractProps) *Listener {
	api := []*client.Transactions{
		client.NewTransactions(defaultOptions),
		client.NewTransactions(tradisysOptions),
	}
	state := &GameState{
		heightToGetMoney: 0,
		last:             "",
	}

	return &Listener{api: api, state: state, props: props}
}

func (s *Listener) Launch() {
	height := uint64(0)

	for {
		h := s.getHeight()
		if h >= height {
			height = h
		}
		err := s.getGameState()
		if err != nil {
			continue
		}

		fmt.Println(int64(s.state.heightToGetMoney), int64(h))

		offset := (height - s.props.BlocksOnGameStart) % s.props.BlocksPerRound
		fmt.Println(s.props.BlocksPerCompetition-offset, "left")

		if s.props.BlocksPerCompetition-offset == 1 {
			s.turn()
			time.Sleep(1 * time.Second)

			continue
		}

		diff := int64(s.state.heightToGetMoney) - int64(h)
		if diff > 0 && diff <= 3 && s.state.last != senderAddress.String() {
			s.turn()
			time.Sleep(30 * time.Second)

			continue
		}

		time.Sleep(1 * time.Second)
	}
}

func (s *Listener) getHeight() uint64 {
	c, err := http.Get(fmt.Sprintf("%s/blocks/height", apiUrlNode))
	if err != nil {
		return 0
	}
	defer c.Body.Close()

	body, err := ioutil.ReadAll(c.Body)
	if err != nil {
		return 0
	}

	h := &struct {
		Height uint64 `json:"height"`
	}{}

	_ = json.Unmarshal(body, h)

	return h.Height
}

func (s *Listener) getGameState() error {
	c, err := http.Get(fmt.Sprintf("%s/addresses/data/%s/%s", apiUrlNode, contractAddress, "RoundsSharedState"))
	if err != nil {
		return err
	}
	defer c.Body.Close()

	body, err := ioutil.ReadAll(c.Body)
	if err != nil {
		return err
	}

	v := &struct {
		Value string `json:"value"`
	}{}

	_ = json.Unmarshal(body, v)
	spl := strings.Split(v.Value, "_")
	stateHeightToGetMoney, _ := strconv.Atoi(spl[0])
	stateLast := strings.Split(spl[2], "-")[1]

	s.state.heightToGetMoney = uint64(stateHeightToGetMoney)
	s.state.last = stateLast

	return nil
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
