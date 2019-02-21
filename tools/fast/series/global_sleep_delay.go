package series

import (
	"time"

	"github.com/filecoin-project/go-filecoin/mining"
)

var GlobalSleepDelay = mining.DefaultBlockTime

// SleepDelay is just a handle method to make sure people
// don't call `time.Sleep` themselves.
func SleepDelay() {
	time.Sleep(GlobalSleepDelay)
}
