package address

import (
	"bytes"
	"encoding/base32"
	"fmt"
	"strconv"

	"gx/ipfs/QmSKyB5faguXT4NqbrXpnRXqaVj5DhSm7x9BtzFydBY1UK/go-leb128"
	"gx/ipfs/QmZp3eKdYQHHAneECmeK6HhiMwTPufmjC8DuuaGKv3unvx/blake2b-simd"
)

const PayloadHashLength = 20
const ChecksumHashLength = 4

var payloadHashConfig = &blake2b.Config{Size: PayloadHashLength}
var checksumHashConfig = &blake2b.Config{Size: ChecksumHashLength}

/*

There are 2 ways a filecoin address can be represented. An address appearing on
chain will always be formatted as raw bytes. An address may also be encoded to
a string, this encoding includes a checksum and network prefix. An address
encoded as a string will never appear on chain, this format is used for sharing
among humans.

Bytes:
|----------|---------|
| protocol | payload |
|----------|---------|
|  1 byte  | n bytes |

String:
|------------|----------|---------|----------|
|  network   | protocol | payload | checksum |
|------------|----------|---------|----------|
| 'f' or 't' |  1 byte  | n bytes | 4 bytes  |

*/

type Address struct{ str string }

var Undef = Address{}

type Network = byte

const (
	Mainnet Network = iota
	Testnet
)

var MainnetPrefix = "f"
var TestnetPrefix = "t"

type Protocol = byte

const (
	ID Protocol = iota
	SECP256K1
	Actor
	BLS
)

func (a Address) Protocol() Protocol {
	return Protocol(a.str[0])
}

func (a Address) Payload() []byte {
	return []byte(a.str[1:])
}

func newAddress(protocol Protocol, payload []byte) Address {
	if protocol < ID || protocol > BLS {
		panic("invalid protocol")
	}
	explen := 1 + len(payload)
	buf := make([]byte, explen)

	buf[0] = protocol
	if c := copy(buf[1:], payload); c != len(payload) {
		panic("copy data length is inconsistent")
	}
	return Address{string(buf)}
}

func NewIDAddress(id uint64) Address {
	return newAddress(ID, leb128.FromUInt64(id))
}

func NewSecp256k1Address(pubkey []byte) Address {
	return newAddress(SECP256K1, AddressHash(pubkey))
}

func NewActorAddress(data []byte) Address {
	return newAddress(Actor, AddressHash(data))
}

func NewBLSAddress(pubkey []byte) Address {
	return newAddress(BLS, pubkey)
}

func Encode(network Network, addr Address) string {
	var ntwk string
	switch network {
	case Mainnet:
		ntwk = MainnetPrefix
	case Testnet:
		ntwk = TestnetPrefix
	default:
		panic("invalid network byte")
	}

	switch addr.Protocol() {
	case SECP256K1, Actor, BLS:
		cksm := Checksum(append([]byte{addr.Protocol()}, addr.Payload()...))
		return ntwk + fmt.Sprintf("%d", addr.Protocol()) + base32.StdEncoding.EncodeToString(append(addr.Payload(), cksm[:]...))
	case ID:
		return ntwk + fmt.Sprintf("%d", addr.Protocol()) + fmt.Sprintf("%d", leb128.ToUInt64(addr.Payload()))
	default:
		panic("invalid protocol byte")
	}
}

func Decode(a string) Address {
	if len(a) < 3 {
		panic("invalid address length, too short")
	}

	if string(a[0]) != MainnetPrefix && string(a[0]) != TestnetPrefix {
		panic("invalid network prefix")
	}

	var protocol Protocol
	switch a[1] {
	case 48:
		protocol = ID
	case 49:
		protocol = SECP256K1
	case 50:
		protocol = Actor
	case 51:
		protocol = BLS
	default:
		panic("invalid protocol")
	}

	raw := a[2:]
	if Protocol(protocol) == ID {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			panic("invalid ID payload")
		}
		return newAddress(protocol, leb128.FromUInt64(id))
	}

	payload, err := base32.StdEncoding.DecodeString(raw)
	if err != nil {
		panic(err)
	}

	if protocol == SECP256K1 || protocol == Actor {
		if len(payload) != 20 {
			panic("invalid hash payload length")
		}
	}

	return newAddress(protocol, payload)
}

func Checksum(ingest []byte) []byte {
	return hash(ingest, checksumHashConfig)
}

func ValidateChecksum(ingest, expect []byte) bool {
	digest := Checksum(ingest)
	return bytes.Equal(digest, expect)
}

func AddressHash(ingest []byte) []byte {
	return hash(ingest, payloadHashConfig)
}

func hash(ingest []byte, cfg *blake2b.Config) []byte {
	hasher, err := blake2b.New(cfg)
	if err != nil {
		// If this happens sth is very wrong.
		panic(fmt.Sprintf("invalid address hash configuration: %v", err))
	}
	if _, err := hasher.Write(ingest); err != nil {
		// blake2bs Write implementation never returns an error in its current
		// setup. So if this happens sth went very wrong.
		panic(fmt.Sprintf("blake2b is unable to process hashes: %v", err))
	}
	return hasher.Sum(nil)
}
