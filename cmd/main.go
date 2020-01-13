package main

import (
	"log"

	"github.com/wavesplatform/gowaves/pkg/crypto"
	"github.com/wavesplatform/gowaves/pkg/proto"
)

var (
	apiUrlNode          = "https://nodes.wavesnodes.com"
	contractAddress     = "3PDtyStFHhEF5LSqPi4amUUAW6KQQQhNaR7"
	senderPublicKey, _  = crypto.NewPublicKeyFromBase58("")
	senderPrivateKey, _ = crypto.NewSecretKeyFromBase58("")
	assetWaves, _       = proto.NewOptionalAssetFromString("WAVES")
	assetMRT, _         = proto.NewOptionalAssetFromString("4uK8i4ThRGbehENwa6MxyLtxAjAo1Rj9fduborGExarC")
	recipient, _        = proto.NewRecipientFromString(contractAddress)
	functionCall        = proto.FunctionCall{
		Default:   false,
		Name:      "move",
		Arguments: proto.Arguments{},
	}
	payments = proto.ScriptPayments{proto.ScriptPayment{
		Amount: 10000,
		Asset:  *assetMRT,
	}}
)

func main() {
	props, err := GetContractProperties()
	if err != nil {
		log.Fatal(err)
	}

	listener := NewListener(props)
	listener.Launch()
}
