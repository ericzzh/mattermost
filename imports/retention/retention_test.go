package retention

import (
// 	"testing"
// 	"time"
// 	// "fmt"
// 	// "sync"
// 
// 	// "github.com/stretchr/testify/require"
// 	// "github.com/stretchr/testify/assert"
// 
// 	// "github.com/mattermost/mattermost-server/v5/shared/mlog"
// 	// "github.com/mattermost/mattermost-server/v5/store/storetest"
// 	// "github.com/mattermost/mattermost-server/v5/store/sqlstore"
// 	"github.com/mattermost/mattermost-server/v5/model"
// 	// "github.com/mattermost/mattermost-server/v5/store"
// 	"github.com/mattermost/mattermost-server/v5/api4"
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
//    Y <- Z,V specific situation
//    Z <- X,W permanent situation
//
// Data setup( following a specifical function call )
// General retention period 5 seconds
// Specific team retention period 3 seconds
// Specific channel retention period 1 seconds
// Specific user direct message 3 seconds
// Permanent period = ""
//
// Root messages
// U send 7 root messages to all channels(interval 1 second)
// Some message is pinned( better first serveral message )
// Odd message will attach a file
// Call retention method
// Expect all messages' time left in the channels are right
// Expect all deleted messages' file is deleted
// Expect all pinned message is not deleted
//
// Thread messages
// U send 7 theads ( 1 root + 6 threads) to all channels(interval 1 second)
// some message is pinned
// some message is not pinned but its thread is pinned
// Call retention method
// Wait extra 3 seconds
// All left message should be with all threads
// Wait extra 3 seconds
// All message and threads should be deleted
// All files with deleted message should be removed
// All Pinned message as a whole ( root + thread) should be left

// func TestRention(t *testing.T) {
// 
// 	th := api4.Setup(t)
// 	defer th.TearDown()
// 	Client := th.Client
// 
// 	X := th.SystemAdminUser
// 	Y := th.SystemManagerUser
// 	Z := th.TeamAdminUser
// 	U := th.BasicUser
// 	V := th.BasicUser2
// 	W := th.CreateUser()
// 	W, _ = th.App.GetUser(W.Id)
// 	// mainHelper.GetSQLStore().GetMaster().Insert(W)
// 
// 	// maybe we should delay login the basic user
// 	// so as to be able to create teams and channels
// 	A := th.CreateTeam()
// 	th.LinkUserToTeam(U, A)
// 	th.LinkUserToTeam(V, A)
// 	th.LinkUserToTeam(W, A)
// 	A_1 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_OPEN, A.Id)
// 	th.App.AddUserToChannel(U, A_1, false)
// 	th.App.AddUserToChannel(V, A_1, false)
// 	th.App.AddUserToChannel(W, A_1, false)
// 	A_2 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
// 	th.App.AddUserToChannel(U, A_2, false)
// 	th.App.AddUserToChannel(V, A_2, false)
// 	th.App.AddUserToChannel(W, A_2, false)
// 	A_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
// 	th.App.AddUserToChannel(U, A_3, false)
// 	th.App.AddUserToChannel(V, A_3, false)
// 	th.App.AddUserToChannel(W, A_3, false)
// 
// 	B := th.CreateTeam()
// 	th.LinkUserToTeam(U, B)
// 	th.LinkUserToTeam(V, B)
// 	th.LinkUserToTeam(W, B)
// 	B_1 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_OPEN, B.Id)
// 	th.App.AddUserToChannel(U, B_1, false)
// 	th.App.AddUserToChannel(V, B_1, false)
// 	th.App.AddUserToChannel(W, B_1, false)
// 	B_2 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, B.Id)
// 	th.App.AddUserToChannel(U, B_2, false)
// 	th.App.AddUserToChannel(V, B_2, false)
// 	th.App.AddUserToChannel(W, B_2, false)
// 	B_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, B.Id)
// 	th.App.AddUserToChannel(U, B_3, false)
// 	th.App.AddUserToChannel(V, B_3, false)
// 	th.App.AddUserToChannel(W, B_3, false)
// 
// 	C := th.CreateTeam()
// 	th.LinkUserToTeam(U, C)
// 	th.LinkUserToTeam(V, C)
// 	th.LinkUserToTeam(W, C)
// 	C_1 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_OPEN, C.Id)
// 	th.App.AddUserToChannel(U, C_1, false)
// 	th.App.AddUserToChannel(V, C_1, false)
// 	th.App.AddUserToChannel(W, C_1, false)
// 	C_2 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, C.Id)
// 	th.App.AddUserToChannel(U, C_2, false)
// 	th.App.AddUserToChannel(V, C_2, false)
// 	th.App.AddUserToChannel(W, C_2, false)
// 	C_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, C.Id)
// 	th.App.AddUserToChannel(U, C_3, false)
// 	th.App.AddUserToChannel(V, C_3, false)
// 	th.App.AddUserToChannel(W, C_3, false)
// 
// 	CreateDmChannel(th, X, Y)
// 	CreateDmChannel(th, X, U)
// 
// 	CreateDmChannel(th, Y, Z)
// 	CreateDmChannel(th, Y, V)
// 
// 	CreateDmChannel(th, Z, X)
// 	CreateDmChannel(th, Z, W)
// 
// 	SetPolicy(SimplePolicy{
// 		period: 5,
// 		specific: []SimpleSpecificPolicy{
// 			{SIMPLE_RETENTION_KIND_TEAM, B.Id, "special team B", 3},
// 			{SIMPLE_RETENTION_KIND_TEAM, C.Id, "permanent team C", 0},
// 			{SIMPLE_RETENTION_KIND_CHANNEL, A_1.Id, "specific channel A_1", 1},
// 			{SIMPLE_RETENTION_KIND_CHANNEL, B_1.Id, "specific channel B_1", 1},
// 			{SIMPLE_RETENTION_KIND_CHANNEL, C_1.Id, "specific channel C_1", 1},
// 			{SIMPLE_RETENTION_KIND_CHANNEL, A_2.Id, "permanent channel A_2", 0},
// 			{SIMPLE_RETENTION_KIND_CHANNEL, B_2.Id, "permanent channel B_2", 0},
// 			{SIMPLE_RETENTION_KIND_CHANNEL, C_2.Id, "permanent channel C_2", 0},
// 			{SIMPLE_RETENTION_KIND_USER, Y.Username, "specific user Y", 3},
// 			{SIMPLE_RETENTION_KIND_USER, Z.Username, "permanent user Z", 0},
// 		},
// 	})
// 	t.Run("testing root message", func(t *testing.T) {
// 
// 		const POSTS_COUNTS = 7
// 
// 		U.Password = "Pa$$word11"
// 		LoginWithClient(U, Client)
// 
// 		//send root post to all channels
// 		// all := []*model.Channel{A_1, A_2, A_3, B_1, B_2, B_3, C_1, C_2, C_3}
// 		all := []*model.Channel{A_1, A_2, A_3, B_1, B_2, B_3, C_1, C_2, C_3}
// 		for _, v := range all {
// 			for i := 0; i < POSTS_COUNTS; i++ {
// 				if i == 0 || i == POSTS_COUNTS {
// 					fileResp, _ := Client.UploadFile([]byte("data"), v.Id, "test"+string(i))
// 					fileId := fileResp.FileInfos[0].Id
// 
// 					Client.CreatePost(&model.Post{
// 						ChannelId: v.Id,
// 						Message:   string(i) + ":with files",
// 						FileIds:   model.StringArray{fileId},
// 					})
// 				} else {
// 
// 					Client.CreatePost(&model.Post{
// 						ChannelId: v.Id,
// 						Message:   string(i) + ":without files",
// 					})
// 				}
// 
// 				if i == 1 {
// 
// 					Client.CreatePost(&model.Post{
// 						ChannelId: v.Id,
// 						Message:   string(i) + ":without files",
// 						IsPinned:  true,
// 					})
// 				}
// 
// 				time.Sleep(1 * time.Second)
// 
// 			}
// 		}
// 
// 	})
// 
// 	// U.Password = "Pa$$word11"
// 	// V.Password = "Pa$$word11"
// 	// W.Password = "Pa$$word11"
// 
// 	// System & admin 3 users have logined
// 	// So we need login another basic users
// 	// var wg sync.WaitGroup
// 	// wg.Add(3)
// 	// go func() {
// 	// 	LoginWithClient(U, Client)
// 	// 	wg.Done()
// 	// }()
// 	// go func() {
// 	// 	LoginWithClient(V, Client)
// 	// 	wg.Done()
// 	// }()
// 	// go func() {
// 	// 	LoginWithClient(W, Client)
// 	// 	wg.Done()
// 	// }()
// 	// wg.Wait()
// 	// th := api4.Setup(t).InitBasic()
// 	// defer th.TearDown()
// 	// Client := th.Client
// 	// 	fileResp, subResponse := Client.UploadFile([]byte("data"), th.BasicChannel.Id, "test")
// 	// 	api4.CheckNoError(t, subResponse)
// 	//         mlog.Debug(fileResp.ToJson())
// 	// 	fileId := fileResp.FileInfos[0].Id
// 	//
// 	// 	postWithFiles, subResponse := Client.CreatePost(&model.Post{
// 	// 		ChannelId: th.BasicChannel.Id,
// 	// 		Message:   "with files",
// 	// 		FileIds:   model.StringArray{fileId},
// 	// 	})
// 	// 	api4.CheckNoError(t, subResponse)
// 	// 	assert.Equal(t, model.StringArray{fileId}, postWithFiles.FileIds)
// 	//
// 	// 	actualPostWithFiles, subResponse := Client.GetPost(postWithFiles.Id, "")
// 	// 	api4.CheckNoError(t, subResponse)
// 	// 	assert.Equal(t, model.StringArray{fileId}, actualPostWithFiles.FileIds)
// 	// StoreTestWithSqlStore(t, testRention)
// }

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
