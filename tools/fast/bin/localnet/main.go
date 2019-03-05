package main

// localnet
//
// localnet is a FAST binary tool that quickly, and easily, sets up a local network
// on the users computer. The network will stay standing till the program is closed.

import (
	"bytes"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gx/ipfs/QmXWZCd8jfaHmt4UDSnjKmGcrQMw95bDGWqEeVLVJjoANX/go-ipfs-files"
	logging "gx/ipfs/QmcuXC5cxs79ro2cUuHs4HQ2bkDLJUYokwL8aivcX6HW3C/go-log"

	"github.com/filecoin-project/go-filecoin/protocol/storage"
	"github.com/filecoin-project/go-filecoin/testhelpers"
	"github.com/filecoin-project/go-filecoin/tools/fast"
	"github.com/filecoin-project/go-filecoin/tools/fast/series"
	lpfc "github.com/filecoin-project/go-filecoin/tools/iptb-plugins/filecoin/local"
)

var (
	workdir   string
	shell     bool
	count     = 5
	blocktime = 5 * time.Second
	err       error
	fil       = 100000
	balance   big.Int

	// SectorSize is the number of bytes that will be used to make power for miners
	SectorSize int64 = 1016
)

func init() {
	logging.SetDebugLogging()

	flag.StringVar(&workdir, "workdir", workdir, "set the working directory")
	flag.BoolVar(&shell, "shell", shell, "drop into a shell")
	flag.IntVar(&count, "count", count, "number of miners")
	flag.DurationVar(&blocktime, "blocktime", blocktime, "duration for blocktime")

	flag.Parse()

	// Set the series global sleep delay to 5 seconds, we will also use this as our
	// block time value.
	series.GlobalSleepDelay = blocktime

	// Set the initial balance
	balance.SetInt64(int64(100 * fil))
}

func main() {
	ctx := context.Background()

	if len(workdir) == 0 {
		workdir, err = ioutil.TempDir("", "localnet")
		if err != nil {
			handleError(err)
		}
	}

	if ok, err := isEmpty(workdir); !ok {
		handleError(err, "workdir is not empty;")
		return
	}

	env, err := fast.NewEnvironmentMemoryGenesis(&balance, workdir)
	if err != nil {
		handleError(err)
		return
	}

	// Defer the teardown, this will shuteverything down for us
	defer env.Teardown(ctx)

	binpath, err := testhelpers.GetFilecoinBinary()
	if err != nil {
		handleError(err, "no binary was found, please build go-filecoin;")
		return
	}

	// Setup localfilecoin plugin options
	options := make(map[string]string)
	options[lpfc.AttrLogJSON] = "0"            // Disable JSON logs
	options[lpfc.AttrLogLevel] = "4"           // Set log level to Info
	options[lpfc.AttrUseSmallSectors] = "true" // Enable small sectors
	options[lpfc.AttrFilecoinBinary] = binpath // Use the repo binary

	genesisURI := env.GenesisCar()
	genesisMiner, err := env.GenesisMiner()
	if err != nil {
		handleError(err, "failed to retrieve miner information from genesis;")
		return
	}
	fastenvOpts := fast.EnvironmentOpts{
		InitOpts:   []fast.ProcessInitOption{fast.POGenesisFile(genesisURI)},
		DaemonOpts: []fast.ProcessDaemonOption{fast.POBlockTime(series.GlobalSleepDelay)},
	}

	// The genesis process is the filecoin node that loads the miner that is
	// define with power in the genesis block, and the prefunnded wallet
	genesis, err := env.NewProcess(ctx, lpfc.PluginName, options, fastenvOpts)
	if err != nil {
		handleError(err, "failed to create genesis process;")
		return
	}

	err = series.SetupGenesisNode(ctx, genesis, genesisMiner.Address, files.NewReaderFile(genesisMiner.Owner))
	if err != nil {
		handleError(err, "failed series.SetupGenesisNode;")
		return
	}

	// Create the processes that we will use to become miners
	var miners []*fast.Filecoin
	for i := 0; i < count; i++ {
		miner, err := env.NewProcess(ctx, lpfc.PluginName, options, fastenvOpts)
		if err != nil {
			handleError(err, "failed to create miner process;")
			return
		}

		miners = append(miners, miner)
	}

	// We will now go through the process of creating miners
	// InitAndStart
	// 1. Initialize node
	// 2. Start daemon
	//
	// Connect
	// 3. Connect to genesis
	//
	// SendFilecoinDefaults
	// 4. Issue FIL to node
	//
	// CreateMinerWithAsk
	// 5. Create a new miner
	// 6. Set the miner price, and get ask
	//
	// ImportAndStore
	// 7. Generated some random data and import it to genesis
	// 8. Genesis propposes a storage deal with miner
	//
	// WaitForDealState
	// 9. Query deal till posted

	var deals []*storage.DealResponse

	for _, miner := range miners {
		err = series.InitAndStart(ctx, miner)
		if err != nil {
			handleError(err, "failed series.InitAndStart;")
			return
		}

		err = series.Connect(ctx, genesis, miner)
		if err != nil {
			handleError(err, "failed series.Connect;")
			return
		}

		err = series.SendFilecoinDefaults(ctx, genesis, miner, fil)
		if err != nil {
			handleError(err, "failed series.SendFilecoinDefaults;")
			return
		}

		pledge := uint64(10)                    // sectors
		collateral := big.NewInt(500)           // FIL
		price := big.NewFloat(0.000000001)      // price per byte/block
		expiry := big.NewInt(24 * 60 * 60 / 30) // ~24 hours

		ask, err := series.CreateMinerWithAsk(ctx, miner, pledge, collateral, price, expiry)
		if err != nil {
			handleError(err, "failed series.CreateMinerWithAsk;")
			return
		}

		var data bytes.Buffer
		dataReader := io.LimitReader(rand.Reader, SectorSize)
		dataReader = io.TeeReader(dataReader, &data)
		_, deal, err := series.ImportAndStore(ctx, genesis, ask, files.NewReaderFile(dataReader))
		if err != nil {
			handleError(err, "failed series.ImportAndStore;")
			return
		}

		deals = append(deals, deal)

	}

	for _, deal := range deals {
		err = series.WaitForDealState(ctx, genesis, deal, storage.Posted)
		if err != nil {
			handleError(err, "failed series.WaitForDealState;")
			return
		}
	}

	if shell {
		client, err := env.NewProcess(ctx, lpfc.PluginName, options, fastenvOpts)
		if err != nil {
			handleError(err, "failed to create client process;")
			return
		}

		err = series.InitAndStart(ctx, client)
		if err != nil {
			handleError(err, "failed series.InitAndStart;")
			return
		}

		err = series.Connect(ctx, genesis, client)
		if err != nil {
			handleError(err, "failed series.Connect;")
			return
		}

		err = series.SendFilecoinDefaults(ctx, genesis, client, fil)
		if err != nil {
			handleError(err, "failed series.SendFilecoinDefaults;")
			return
		}

		interval, err := client.StartLogCapture()
		if err != nil {
			handleError(err, "failed to start log capture;")
			return
		}

		if err := client.Shell(); err != nil {
			handleError(err, "failed to run client shell;")
			return
		}

		interval.Stop()
		fmt.Println("===================================")
		fmt.Println("===================================")
		io.Copy(os.Stdout, interval)
		fmt.Println("===================================")
		fmt.Println("===================================")
	}

	fmt.Println("Finished!")
	fmt.Println("Ctrl-C to handleError")

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	<-signals
}

func handleError(err error, msg ...string) {
	if err == nil {
		return
	}

	if len(msg) != 0 {
		fmt.Println(msg[0], err)
	} else {
		fmt.Println(err)
	}
}

// https://stackoverflow.com/a/3070891://stackoverflow.com/a/30708914
func isEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
