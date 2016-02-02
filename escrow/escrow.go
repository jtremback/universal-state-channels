package escrow

import (
	"errors"
	"github.com/agl/ed25519"
	"github.com/golang/protobuf/proto"
	"github.com/jtremback/upc/schema"
	"github.com/jtremback/upc/wire"
	"math"
	"math/big"
)

type Channel schema.Channel
type Identity schema.Identity

// func ChannelFromOpeningTx(ev *wire.Envelope) (*Channel, error) {
// 	err := VerifySignatures(ev)
// 	if err != nil {
// 		return nil, err
// 	}

// 	otx := wire.OpeningTx{}
// 	err = proto.Unmarshal(ev.Payload, &otx)
// 	if err != nil {
// 		return nil, err
// 	}
// }

func sliceTo64Byte(slice []byte) *[64]byte {
	var array [64]byte
	copy(array[:], slice[:64])
	return &array
}

func sliceTo32Byte(slice []byte) *[32]byte {
	var array [32]byte
	copy(array[:], slice[:32])
	return &array
}

// VerifyOpeningTx checks the signatures of an OpeningTx, unmarshals it and returns it.
func VerifyOpeningTx(ev *wire.Envelope) (*wire.OpeningTx, error) {
	otx := wire.OpeningTx{}
	err := proto.Unmarshal(ev.Payload, &otx)
	if err != nil {
		return nil, err
	}

	// Check signatures
	if !ed25519.Verify(sliceTo32Byte(otx.Pubkey1), ev.Payload, sliceTo64Byte(ev.Signature1)) {
		return nil, errors.New("signature 1 invalid")
	}
	if !ed25519.Verify(sliceTo32Byte(otx.Pubkey2), ev.Payload, sliceTo64Byte(ev.Signature2)) {
		return nil, errors.New("signature 2 invalid")
	}

	return &otx, nil
}

// VerifyUpdateTx checks the signatures of an UpdateTx, unmarshals it and returns it.
func (ch *Channel) VerifyUpdateTx(ev *wire.Envelope) (*wire.UpdateTx, error) {
	// Check signatures
	if !ed25519.Verify(sliceTo32Byte(ch.OpeningTx.Pubkey1), ev.Payload, sliceTo64Byte(ev.Signature1)) {
		return nil, errors.New("signature 1 invalid")
	}
	if !ed25519.Verify(sliceTo32Byte(ch.OpeningTx.Pubkey2), ev.Payload, sliceTo64Byte(ev.Signature2)) {
		return nil, errors.New("signature 2 invalid")
	}

	utx := wire.UpdateTx{}
	err := proto.Unmarshal(ev.Payload, &utx)
	if err != nil {
		return nil, err
	}

	return &utx, nil
}

func (ch *Channel) StartClose(utx *wire.UpdateTx) error {
	if ch.State != schema.Channel_PendingClosed {
		if ch.LastFullUpdateTx.SequenceNumber > utx.SequenceNumber {
			return errors.New("update tx with higher sequence number exists")
		}
	}

	ch.State = schema.Channel_PendingClosed
	ch.LastFullUpdateTx = utx
	return nil
}

// AddFulfillment verifies a fulfillment's signature and adds it to the Channel's
// Fulfillments array.
func (ch *Channel) AddFulfillment(ev *wire.Envelope, eval func()) error {
	if ch.State != schema.Channel_PendingClosed {
		return errors.New("channel must be pending closed")
	}

	if !ed25519.Verify(sliceTo32Byte(ch.OpeningTx.Pubkey1), ev.Payload, sliceTo64Byte(ev.Signature1)) ||
		!ed25519.Verify(sliceTo32Byte(ch.OpeningTx.Pubkey2), ev.Payload, sliceTo64Byte(ev.Signature1)) {
		return errors.New("signature invalid")
	}

	ful := wire.Fulfillment{}
	err := proto.Unmarshal(ev.Payload, &ful)
	if err != nil {
		return err
	}

	ch.Fulfillments = append(ch.Fulfillments, &ful)
	return nil
}

// func parseConditionalMultiplier(s string) (*big.Rat, error) {
// 	rat := big.NewRat(0, 1)
// 	rat.SetString(s)
// 	if rat.Cmp(big.NewRat(1, 1)) > 0 {
// 		return rat, errors.New("conditional multiplier is larger than 1")
// 	}
// 	return rat, nil
// }

// func round(input float64) float64 {
// 	if input < 0 {
// 		return math.Ceil(input - 0.5)
// 	}
// 	return math.Floor(input + 0.5)
// }

// EvaluateConditions takes a function `fn` that takes 3 string arguments-
// Name, ConditionalData, and FulfillmentData. Name is the name of the Condition,
// ConditionalData was supplied by the condition (signed by both parties), and
// FulfillmentData was supplied by the Fulfillment (signed by both parties).
// fn must return a ConditionalMultiplier, a number in string form between
// 1 and 0, expressed either as a fraction ("1/2"), or a decimal ("0.5")
//
// It then iterates through the Channel's fulfillments and calls fn on each of
// them and their corresponding conditions. It multiplies the returned conditional
// multiplier by the Condition's conditional transfer and adds the result to the
// Condition's NetTransfer. If there are 2 Fulfillments for one condition, the
// Fulfillment evaluating to the higher ConditionalMultiplier is used.
// func (ch *Channel) EvaluateConditions(fn func(string, string, string) string) (int64, error) {
// 	var fulMap map[string]*big.Rat

// 	for _, ful := range ch.Fulfillments {
// 		// Get corresponding condition
// 		cond := ch.LastFullUpdateTx.Conditions[ful.Condition]
// 		// Evaluate to get conditional multiplier
// 		cm, err := parseConditionalMultiplier(fn(cond.PresetCondition, cond.Data, ful.Data))
// 		if err != nil {
// 			return 0, err
// 		}

// 		// Get previous conditional multiplier, if any
// 		prevCm, ok := fulMap[ful.Condition]
// 		if ok {
// 			// If prevCm is lower, replace
// 			if prevCm.Cmp(cm) < 0 {
// 				fulMap[ful.Condition] = cm
// 			}
// 		} else {
// 			// If prevCm did not exist, set to cm
// 			fulMap[ful.Condition] = cm
// 		}
// 	}

// 	nt := big.NewRat(ch.LastFullUpdateTx.NetTransfer, 1)

// 	var r big.Rat
// 	for _, ful := range ch.Fulfillments {
// 		cond := ch.LastFullUpdateTx.Conditions[ful.Condition]
// 		cm, _ := fulMap[ful.Condition]
// 		nt.Add(nt, r.Mul(big.NewRat(cond.ConditionalTransfer, 1), cm))
// 	}

// 	n, _ := nt.Float64()

// 	return int64(round(n)), nil
// }
