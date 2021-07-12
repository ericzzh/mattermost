package retention

import (
	"testing"
	// "fmt"

	// "github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v5/shared/mlog"
	"github.com/mattermost/mattermost-server/v5/store/storetest"
	// "github.com/mattermost/mattermost-server/v5/store/sqlstore"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

func TestRention(t *testing.T) {
	StoreTestWithSqlStore(t, testRention)
}

func testRention(t *testing.T, ss store.Store, s storetest.SqlStore) {
	setup(ss, s)
	mlog.Debug("Enting testRention")
}

// setup is to setup the backgroup of testing
//
// Enviroment setup
// Create a team A for general situation
//    a channel A-1 for special situation
//    a channel A-2 for general situation
// Create a team B for specific situation
//    a channel B-1 for special situation
//    a channel B-2 for general situation
// Create 3 users
//    X, Y general
//    Z specific
//
// Data setup( following a specifical function call )
// General retention period 5 seconds
// Specific team retention period 3 seconds
// Specific channel retention period 1 seconds
// Specific user direct message 3 seconds
//
// Root messages
// X send 7 root messages to all channels(interval 1 second)
// Odd message will attach a file
// Call retention method
// Expect all messages' time left in the channels are right
//
// Thread messages
// X send 7 theads ( 1 root + 6 threads) to all channels(interval 1 second)
// Call retention method
// Wait extra 3 seconds
// All message s except A-2 should be deleted
// Wait extra 3 seconds
// All thread should be deleted
//
func setup(ss store.Store, s storetest.SqlStore) {

	o1 := model.Team{}
	o1.DisplayName = "DisplayName"
	o1.Name = NewTestId()
	o1.Email = MakeEmail()
	o1.Type = model.TEAM_OPEN

	ss.Team().Save(&o1)
}
