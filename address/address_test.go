package address

import (
	"crypto/ecdsa"
	"testing"

	"gx/ipfs/QmPVkJMTeRC6iBByPWdrRkD3BE5UXsj5HPzb4kPqL186mS/testify/require"

	"github.com/filecoin-project/go-filecoin/crypto"
)

func TestIDAddress(t *testing.T) {
	require := require.New(t)

	addr := NewIDAddress(100)
	require.Equal(ID, addr.Protocol())

	maybe := Decode(Encode(Mainnet, addr))
	require.Equal(addr, maybe)

}

func TestSecp256k1Address(t *testing.T) {
	require := require.New(t)

	sk, err := crypto.GenerateKey()
	require.NoError(err)

	addr := NewSecp256k1Address(crypto.ECDSAPubToBytes(sk.Public().(*ecdsa.PublicKey)))
	require.Equal(SECP256K1, addr.Protocol())

	maybe := Decode(Encode(Mainnet, addr))
	require.Equal(addr, maybe)

}
