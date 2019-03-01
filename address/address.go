package address

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"gx/ipfs/QmSKyB5faguXT4NqbrXpnRXqaVj5DhSm7x9BtzFydBY1UK/go-leb128"
	"gx/ipfs/QmZp3eKdYQHHAneECmeK6HhiMwTPufmjC8DuuaGKv3unvx/blake2b-simd"
	logging "gx/ipfs/QmbkT7eMTyXfpeyB3ZMxxcxg7XH8t6uXp49jqzz4HB7BGF/go-log"
	cbor "gx/ipfs/QmcZLyosDwMKdB6NLRsiss9HXzDPhVhhRtPy67JFKTDQDX/go-ipld-cbor"
	"gx/ipfs/QmdBzoMxsBpojBfN1cv5GnKtB7sfYBMoLH7p9qSyEVYXcu/refmt/obj/atlas"
)

var log = logging.Logger("address")

func init() {
	cbor.RegisterCborType(addressAtlasEntry)

	TestAddress = NewActorAddress([]byte("satoshi"))
	TestAddress2 = NewActorAddress([]byte("nakamoto"))

	NetworkAddress = NewActorAddress([]byte("filecoin"))
	StorageMarketAddress = NewActorAddress([]byte("storage"))
	PaymentBrokerAddress = NewActorAddress([]byte("payments"))
}

var addressAtlasEntry = atlas.BuildEntry(Address{}).Transform().
	TransformMarshal(atlas.MakeMarshalTransformFunc(
		func(a Address) ([]byte, error) {
			return []byte(a.str), nil
		})).
	TransformUnmarshal(atlas.MakeUnmarshalTransformFunc(
		func(x []byte) (Address, error) {
			return Address{string(x)}, nil
		})).
	Complete()

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

func (a Address) Bytes() []byte {
	return []byte(a.str)
}

func (a Address) String() string {
	str, err := encode(Mainnet, a)
	if err != nil {
		panic(err)
	}
	return str
}

func (a Address) Equal(b Address) bool {
	return bytes.Equal(a.Bytes(), b.Bytes())
}

func (a Address) Unmarshal(b []byte) error {
	return cbor.DecodeInto(b, &a)
}

func (a Address) Marshal() ([]byte, error) {
	return cbor.DumpObject(a)
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
	log.Debugf("new address protocol: %x, payload: %v", protocol, payload)
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

func NewFromString(addr string) (Address, error) {
	return decode(addr)
}

func NewFromBytes(addr []byte) Address {
	return Address{string(addr)}
}

func encode(network Network, addr Address) (string, error) {
	var ntwk string
	switch network {
	case Mainnet:
		ntwk = MainnetPrefix
	case Testnet:
		ntwk = TestnetPrefix
	default:
		panic("invalid network byte")
	}

	var strAddr string
	switch addr.Protocol() {
	case SECP256K1, Actor, BLS:
		cksm := Checksum(append([]byte{addr.Protocol()}, addr.Payload()...))
		strAddr = ntwk + fmt.Sprintf("%d", addr.Protocol()) + AddressEncoding.WithPadding(-1).EncodeToString(append(addr.Payload(), cksm[:]...))
	case ID:
		strAddr = ntwk + fmt.Sprintf("%d", addr.Protocol()) + fmt.Sprintf("%d", leb128.ToUInt64(addr.Payload()))
	default:
		return "", errors.New("invalid protocol byte")
	}
	log.Debugf("encoded address: %s", strAddr)
	return strAddr, nil
}

func decode(a string) (Address, error) {
	log.Debugf("decoding address: %s", a)
	if len(a) < 3 {
		return Undef, errors.New("invalid address length, too short")
	}

	if string(a[0]) != MainnetPrefix && string(a[0]) != TestnetPrefix {
		return Undef, errors.New("invalid network prefix")
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
		return Undef, errors.New("invalid protocol")
	}

	raw := a[2:]
	if Protocol(protocol) == ID {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return Undef, errors.New("invalid ID payload")
		}
		return newAddress(protocol, leb128.FromUInt64(id)), nil
	}

	payloadcksm, err := AddressEncoding.WithPadding(-1).DecodeString(raw)
	if err != nil {
		return Undef, err
	}
	payload := payloadcksm[:len(payloadcksm)-ChecksumHashLength]
	cksm := payloadcksm[len(payloadcksm)-ChecksumHashLength:]

	if protocol == SECP256K1 || protocol == Actor {
		if len(payload) != 20 {
			return Undef, errors.New(fmt.Sprintf("invalid hash payload length: %d, payload: %v", len(payload), payload))
		}
	}

	if !ValidateChecksum(append([]byte{protocol}, payload...), cksm) {
		return Undef, errors.New("invalid checksum")
	}

	return newAddress(protocol, payload), nil
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
