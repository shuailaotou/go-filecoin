package address

import (
	"encoding/base32"

	"gx/ipfs/QmZp3eKdYQHHAneECmeK6HhiMwTPufmjC8DuuaGKv3unvx/blake2b-simd"
)

var (
	// TODO remove TestAddresses
	// TestAddress is an account with some initial funds in it
	TestAddress Address
	// TestAddress2 is an account with some initial funds in it
	TestAddress2 Address

	// NetworkAddress is the filecoin network
	NetworkAddress Address
	// StorageMarketAddress is the hard-coded address of the filecoin storage market
	StorageMarketAddress Address
	// PaymentBrokerAddress is the hard-coded address of the filecoin storage market
	PaymentBrokerAddress Address
)

const PayloadHashLength = 20
const ChecksumHashLength = 4

var payloadHashConfig = &blake2b.Config{Size: PayloadHashLength}
var checksumHashConfig = &blake2b.Config{Size: ChecksumHashLength}

const encodeStd = "abcdefghijklmnopqrstuvwxyz234567"

var AddressEncoding = base32.NewEncoding(encodeStd)
