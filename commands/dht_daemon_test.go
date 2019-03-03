package commands_test

import (
	"testing"
	"time"
	//	"fmt"

	"gx/ipfs/QmPVkJMTeRC6iBByPWdrRkD3BE5UXsj5HPzb4kPqL186mS/testify/assert"
	"gx/ipfs/QmPVkJMTeRC6iBByPWdrRkD3BE5UXsj5HPzb4kPqL186mS/testify/require"
	"gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	mh "gx/ipfs/QmerPMzPk1mJVowm8KgmoknWa4yCYvvugMPsgWmDNUvDLW/go-multihash"

	th "github.com/filecoin-project/go-filecoin/testhelpers"
)

const relayRendevous = "/libp2p/relay"

// Taken from https://github.com/libp2p/go-libp2p-discovery/blob/master/routing.go
// This function takes a name space string and returns the cid used as a
// DHT key by providers of the service given by the namespace.
func nsToCid(ns string) (cid.Cid, error) {
	h, err := mh.Sum([]byte(ns), mh.SHA2_256, -1)
	if err != nil {
		return cid.Undef, err
	}

	return cid.NewCidV1(cid.Raw, h), nil
}

func TestSwarmSeesRelayProvider(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	assert := assert.New(t)

	dClient := th.NewDaemon(t, th.SwarmAddr("/ip4/0.0.0.0/tcp/6000")).Start()
	defer dClient.ShutdownSuccess()

	dProvider := th.NewDaemon(t, th.SwarmAddr("/ip4/0.0.0.0/tcp/6001"), th.IsRelay).Start()
	defer dProvider.ShutdownSuccess()

	dClient.ConnectSuccess(dProvider)

	time.Sleep(30 * time.Second)

	provKey, err := nsToCid(relayRendevous)
	require.NoError(err)

	dhtFindProvOutput := dClient.RunSuccess("dht", "findprovs", provKey.String()).ReadStdoutTrimNewlines()
	assert.Contains(dhtFindProvOutput, dProvider.GetID())
}
