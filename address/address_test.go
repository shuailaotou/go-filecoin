package address

import (
	//"fmt"
	"testing"

	"gx/ipfs/QmPVkJMTeRC6iBByPWdrRkD3BE5UXsj5HPzb4kPqL186mS/testify/require"
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

	addr := NewSecp256k1Address()
}
