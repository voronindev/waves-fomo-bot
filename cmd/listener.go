package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/wavesplatform/gowaves/pkg/client"
	"github.com/wavesplatform/gowaves/pkg/proto"
)

var (
	errIncomingDataAreWrong = errors.New("incoming data are wrong")
	tradisysOptions         = client.Options{
		BaseUrl: "https://gamenode.tradisys.com",
		Client:  &http.Client{Timeout: 3 * time.Second},
	}
	defaultOptions = client.Options{
		BaseUrl: "https://nodes.wavesnodes.com",
		Client:  &http.Client{Timeout: 3 * time.Second},
	}
)

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
		h, err := s.getHeight()
		if err != nil {
			continue
		}

		if h >= height {
			height = h
		}

		err = s.getGameState()
		if err != nil {
			continue
		}

		offset := (height - s.props.BlocksOnGameStart) % s.props.BlocksPerRound
		fmt.Println(s.props.BlocksPerCompetition-offset, "offset")

		if s.props.BlocksPerCompetition-offset == 1 {
			fmt.Println("turn: last block")
			s.turn()
			time.Sleep(1 * time.Second)

			continue
		}

		diff := int64(s.state.heightToGetMoney) - int64(h)
		if diff > 0 {
			fmt.Println(diff, "blocks left")
		}

		if diff > 0 && diff <= 4 && s.state.last != senderAddress.String() {
			fmt.Println("turn: prolong")
			s.turn()
			time.Sleep(15 * time.Second)

			continue
		}

		time.Sleep(10 * time.Second)
	}
}

func (s *Listener) getHeight() (uint64, error) {
	c, err := http.Get(fmt.Sprintf("%s/blocks/height", apiUrlNode))
	if err != nil {
		return 0, err
	}
	defer c.Body.Close()

	body, err := ioutil.ReadAll(c.Body)
	if err != nil {
		return 0, err
	}

	h := &struct {
		Height uint64 `json:"height"`
	}{}

	err = json.Unmarshal(body, h)
	if err != nil {
		return 0, err
	}

	return h.Height, nil
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

	err = json.Unmarshal(body, v)
	if err != nil {
		return err
	}

	spl := strings.Split(v.Value, "_")
	if len(spl) == 3 {
		stateHeightToGetMoney, err := strconv.Atoi(spl[0])
		if err != nil {
			return err
		}

		participants := strings.Split(spl[2], "-")
		if len(participants) >= 2 {
			stateLast := participants[1]
			s.state.heightToGetMoney = uint64(stateHeightToGetMoney)
			s.state.last = stateLast
		} else {
			return errIncomingDataAreWrong
		}
	} else {
		return errIncomingDataAreWrong
	}

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
