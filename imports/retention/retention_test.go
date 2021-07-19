package retention

import (
	"fmt"
	"strconv"
	// "strconv"
	"testing"
	"time"

	// "fmt"
	// 	// "sync"
	//
	// 	// "github.com/stretchr/testify/require"
	// 	// "github.com/stretchr/testify/assert"
	//
	// "github.com/mattermost/mattermost-server/v5/shared/mlog"
	// 	// "github.com/mattermost/mattermost-server/v5/store/storetest"
	// 	// "github.com/mattermost/mattermost-server/v5/store/sqlstore"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/shared/mlog"

	// 	// "github.com/mattermost/mattermost-server/v5/store"
	"github.com/mattermost/mattermost-server/v5/api4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
// Cases
// root
// root witn threads
// pinned root
// pinned thread
// root witn threads, some with files
// All deleted
// pined/unpined some tines
// edit some times
// react some times
// flag some post
// permanent with some deletion
// rule merge
// general case(all channels include)
// specific case(some channel icnluded)
// empty folder after remove, unempty folder after removal

func TestRetention(t *testing.T) {

	th := api4.Setup(t)
	defer th.TearDown()
	th.Server.Config().SqlSettings.DataSource = mainHelper.Settings.DataSource
	Client := th.Client

	X := th.SystemAdminUser
	Y := th.SystemManagerUser
	Z := th.TeamAdminUser
	U := th.BasicUser
	V := th.BasicUser2
	W := th.CreateUser()
	W, _ = th.App.GetUser(W.Id)
	mainHelper.GetSQLStore().GetMaster().Insert(W)

	// maybe we should delay login the basic user
	// so as to be able to create teams and channels
	A := th.CreateTeam()
	th.LinkUserToTeam(U, A)
	th.LinkUserToTeam(V, A)
	th.LinkUserToTeam(W, A)
	A_1 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_OPEN, A.Id)
	th.App.AddUserToChannel(U, A_1, false)
	th.App.AddUserToChannel(V, A_1, false)
	th.App.AddUserToChannel(W, A_1, false)
	A_2 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
	th.App.AddUserToChannel(U, A_2, false)
	th.App.AddUserToChannel(V, A_2, false)
	th.App.AddUserToChannel(W, A_2, false)
	A_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
	th.App.AddUserToChannel(U, A_3, false)
	th.App.AddUserToChannel(V, A_3, false)
	th.App.AddUserToChannel(W, A_3, false)

	B := th.CreateTeam()
	th.LinkUserToTeam(U, B)
	th.LinkUserToTeam(V, B)
	th.LinkUserToTeam(W, B)
	B_1 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_OPEN, B.Id)
	th.App.AddUserToChannel(U, B_1, false)
	th.App.AddUserToChannel(V, B_1, false)
	th.App.AddUserToChannel(W, B_1, false)
	B_2 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, B.Id)
	th.App.AddUserToChannel(U, B_2, false)
	th.App.AddUserToChannel(V, B_2, false)
	th.App.AddUserToChannel(W, B_2, false)
	B_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, B.Id)
	th.App.AddUserToChannel(U, B_3, false)
	th.App.AddUserToChannel(V, B_3, false)
	th.App.AddUserToChannel(W, B_3, false)

	C := th.CreateTeam()
	th.LinkUserToTeam(U, C)
	th.LinkUserToTeam(V, C)
	th.LinkUserToTeam(W, C)
	C_1 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_OPEN, C.Id)
	th.App.AddUserToChannel(U, C_1, false)
	th.App.AddUserToChannel(V, C_1, false)
	th.App.AddUserToChannel(W, C_1, false)
	C_2 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, C.Id)
	th.App.AddUserToChannel(U, C_2, false)
	th.App.AddUserToChannel(V, C_2, false)
	th.App.AddUserToChannel(W, C_2, false)
	C_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, C.Id)
	th.App.AddUserToChannel(U, C_3, false)
	th.App.AddUserToChannel(V, C_3, false)
	th.App.AddUserToChannel(W, C_3, false)

	CreateDmChannel(th, X, Y)
	CreateDmChannel(th, X, U)

	CreateDmChannel(th, Y, Z)
	CreateDmChannel(th, Y, V)

	CreateDmChannel(th, Z, X)
	CreateDmChannel(th, Z, W)

        _ = mlog.Debug
        _ = fmt.Sprintf

// 	t.Run("tesing merge channel", func(t *testing.T) {
// 
// 		uc, err := th.Server.Store.Channel().GetChannels("", X.Id, true, 0)
// 		require.NoError(t, err, "Get direct channel error")
// 
// 		SetPolicy(SimplePolicy{
// 			period: 1,
// 			team: SimpleSpecificPolicy{
// 				B.Id: 2,
// 				C.Id: 3,
// 			},
// 			channel: SimpleSpecificPolicy{
// 
// 				A_1.Id: 4,
// 				A_2.Id: 5,
// 				B_1.Id: 6,
// 				B_2.Id: 7,
// 				C_1.Id: 8,
// 				C_2.Id: 9,
// 			},
// 			user: SimpleSpecificPolicy{
// 				X.Username: 10,
// 			},
// 		})
// 		prObj, err := New(th.App)
// 		require.NoError(t, err, "should call New succussfully")
// 
// 		type temppl struct {
// 			name   string
// 			id     string
// 			period time.Duration
// 		}
// 		func(pl []temppl) {
// 
// 			for _, ch := range pl {
// 
// 				pa, ok := prObj.merged[ch.id]
// 				assert.Equal(t, true, ok, fmt.Sprintf("%s should be found", ch.name))
// 				assert.Equal(t, ch.period, pa, fmt.Sprintf("%s period wrong", ch.name))
// 
// 			}
// 
// 		}(
// 			[]temppl{
// 				{"A_1", A_1.Id, 4},
// 				{"A_2", A_2.Id, 5},
// 				{"B_1", B_1.Id, 6},
// 				{"B_2", B_2.Id, 7},
// 				{"C_1", C_1.Id, 8},
// 				{"C_2", C_2.Id, 9},
// 				{"B_3", B_3.Id, 2},
// 				{"C_3", C_3.Id, 3},
// 				{"X", []*model.Channel(*uc)[0].Id, 10},
// 				{"X", []*model.Channel(*uc)[1].Id, 10},
// 				{"X", []*model.Channel(*uc)[2].Id, 10},
// 			},
// 		)
// 
// 	})

	t.Run("testing prune root posts and some file", func(t *testing.T) {

		const period = 7
		// SetPolicy(SimplePolicy{
		// 	period: 5,
		// })

		const POSTS_COUNTS = 7

		U.Password = "Pa$$word11"
		LoginWithClient(U, Client)

		fileinfos_backup := []*model.FileInfo{}

		for i := 0; i < POSTS_COUNTS; i++ {
			var fileId string
			if i%2 == 0 {
				fileResp, _ := Client.UploadFile([]byte("data"), A_3.Id, "test"+strconv.Itoa(i))
				fileId = fileResp.FileInfos[0].Id
                                require.NotEmptyf(t, fileId, "upload file should not be empty.", fileResp.ToJson())

			}

			post, _ := Client.CreatePost(&model.Post{
				ChannelId: A_3.Id,
				Message:   strconv.Itoa(i) + ":with files",
				FileIds:   model.StringArray{fileId},
			})

			post, err := th.App.GetSinglePost(post.Id)
                        require.Nilf(t, err, "post should be there.. %v, but there is err:  %v", post.ToJson(), err.ToJson())

			if i%2 == 0 {
				fileinfo, err := th.App.GetFileInfosForPost((*post).Id, true)
				require.Nil(t, err)
				fileinfos_backup = append(fileinfos_backup, fileinfo...)
			}

			time.Sleep(1 * time.Second)
		}


		pr, err := New(th.App)
		require.NoError(t, err)

		endTime := model.GetMillisForTime(time.Now().Add(-time.Second * period))
		err = pr.pruneActions([]string{A_3.Id}, nil, period)
		require.NoError(t, err, "there should be no errors after first pruning.")

		posts, _ := th.App.GetPosts(A_3.Id, 0, 100)
		require.Greaterf(t, len(posts.Posts),0, "there should be some posts after first pruning.")

		for _, post := range posts.Posts {
			assert.Greater(t, post.UpdateAt, endTime, "left posts should be greater than pruning end time")
			if len([]string(post.FileIds)) != 0 {
				fileinfos, err := th.App.GetFileInfosForPost(post.Id, true)
				require.Nil(t, err)
				for _, fileinfo := range fileinfos {
					b, _ := th.App.FileExists(fileinfo.Path)
					assert.Equalf(t, true, b, "file: %v should exist.", fileinfo.Path)

				}
			}
		}

                time.Sleep(3 * time.Second)
		err = pr.pruneActions([]string{A_3.Id}, nil, 1)
		require.NoError(t, err, "there should be no errors after second pruning.")

		for _, fileinfo := range fileinfos_backup {

			b, _ := th.App.FileExists(fileinfo.Path)
			assert.Equalf(t, false, b, "file: %v should not exist after second pruning.", fileinfo.Path)

		}

                dirs, _ := th.App.ListDirectory(".")
                assert.Equalf(t, 0, len(dirs), "there should be no directory after second pruncing. but....%v", dirs)


	})

	// 	t.Run("testing gernal case", func(t *testing.T) {
	//
	// 		const POSTS_COUNTS = 7
	//
	// 		U.Password = "Pa$$word11"
	// 		LoginWithClient(U, Client)
	//
	// 		for i := 0; i < POSTS_COUNTS; i++ {
	// 			fileResp, _ := Client.UploadFile([]byte("data"), A_3.Id, "test"+strconv.Itoa(i))
	// 			fileId := fileResp.FileInfos[0].Id
	//
	// 			Client.CreatePost(&model.Post{
	// 				ChannelId: A_3.Id,
	// 				Message:   strconv.Itoa(i) + ":with files",
	// 				FileIds:   model.StringArray{fileId},
	// 			})
	//
	// 			time.Sleep(1 * time.Second)
	// 		}
	//
	// 		// posts, _ := th.App.GetPosts(A_3.Id, 0, 100)
	// 		// mlog.Debug(fmt.Sprintf("General case created: %v", posts.ToJson()))
	//
	// 		prObj.Prune()
	//
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
	// 		all := []*model.Channel{A_1, A_2, A_3, B_3, C_3}
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

	// U.Password = "Pa$$word11"
	// V.Password = "Pa$$word11"
	// W.Password = "Pa$$word11"

	// System & admin 3 users have logined
	// So we need login another basic users
	// var wg sync.WaitGroup
	// wg.Add(3)
	// go func() {
	// 	LoginWithClient(U, Client)
	// 	wg.Done()
	// }()
	// go func() {
	// 	LoginWithClient(V, Client)
	// 	wg.Done()
	// }()
	// go func() {
	// 	LoginWithClient(W, Client)
	// 	wg.Done()
	// }()
	// wg.Wait()
	// th := api4.Setup(t).InitBasic()
	// defer th.TearDown()
	// Client := th.Client
	// 	fileResp, subResponse := Client.UploadFile([]byte("data"), th.BasicChannel.Id, "test")
	// 	api4.CheckNoError(t, subResponse)
	//         mlog.Debug(fileResp.ToJson())
	// 	fileId := fileResp.FileInfos[0].Id
	//
	// 	postWithFiles, subResponse := Client.CreatePost(&model.Post{
	// 		ChannelId: th.BasicChannel.Id,
	// 		Message:   "with files",
	// 		FileIds:   model.StringArray{fileId},
	// 	})
	// 	api4.CheckNoError(t, subResponse)
	// 	assert.Equal(t, model.StringArray{fileId}, postWithFiles.FileIds)
	//
	// 	actualPostWithFiles, subResponse := Client.GetPost(postWithFiles.Id, "")
	// 	api4.CheckNoError(t, subResponse)
	// 	assert.Equal(t, model.StringArray{fileId}, actualPostWithFiles.FileIds)
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
