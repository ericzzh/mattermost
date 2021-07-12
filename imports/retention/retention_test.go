package retention
import (
	"testing"
	// "fmt"

	// "github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/v5/shared/mlog"
	// "github.com/mattermost/mattermost-server/v5/store/storetest"
	// "github.com/mattermost/mattermost-server/v5/store/sqlstore"
	"github.com/mattermost/mattermost-server/v5/model"
	// "github.com/mattermost/mattermost-server/v5/store"
	"github.com/mattermost/mattermost-server/v5/api4"
)

// Enviroment setup
// Create a team A for general situation
//    a channel A-1 for specific situation
//    a channel A-2 for permanent situation
//    a channel A-3 for general situation
// Create a team B for specific situation
//    a channel B-1 for specific situation
//    a channel B-2 for permanent situation
//    a channel B-3 for general situation
// Create a team C for permanent situation
//    a channel C-1 for specific situation
//    a channel C-2 for permanent situation
//    a channel C-3 for general situation
// Create 6 users
//    X <- Y,U general situation
//    X <- Z,V specific situation
//    Y <- Z,W permanent situation
//
// Data setup( following a specifical function call )
// General retention period 5 seconds
// Specific team retention period 3 seconds
// Specific channel retention period 1 seconds
// Specific user direct message 3 seconds
// Permanent period = ""
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
// All left message should be with all threads
// Wait extra 3 seconds
// All message and threads should be deleted

func TestRention(t *testing.T) {

	th := api4.Setup(t).InitBasic()
	defer th.TearDown()
	Client := th.Client
	fileResp, subResponse := Client.UploadFile([]byte("data"), th.BasicChannel.Id, "test")
	api4.CheckNoError(t, subResponse)
        mlog.Debug(fileResp.ToJson())
	fileId := fileResp.FileInfos[0].Id

	postWithFiles, subResponse := Client.CreatePost(&model.Post{
		ChannelId: th.BasicChannel.Id,
		Message:   "with files",
		FileIds:   model.StringArray{fileId},
	})
	api4.CheckNoError(t, subResponse)
	assert.Equal(t, model.StringArray{fileId}, postWithFiles.FileIds)

	actualPostWithFiles, subResponse := Client.GetPost(postWithFiles.Id, "")
	api4.CheckNoError(t, subResponse)
	assert.Equal(t, model.StringArray{fileId}, actualPostWithFiles.FileIds)
	// StoreTestWithSqlStore(t, testRention)
}

// func testRention(t *testing.T, ss store.Store, s storetest.SqlStore) {
// 	setup(ss, s)
// 	mlog.Debug("Enting testRention")
// }

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
// func setup(ss store.Store, s storetest.SqlStore) {
//
// 	o1 := model.Team{}
// 	o1.DisplayName = "DisplayName"
// 	o1.Name = NewTestId()
// 	o1.Email = MakeEmail()
// 	o1.Type = model.TEAM_OPEN
//
// 	ss.Team().Save(&o1)
// }
