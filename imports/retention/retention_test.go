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

	_ = mlog.Debug
	_ = fmt.Sprintf

	t.Run("testing merge channel", func(t *testing.T) {

		// Data preparing
		A := th.CreateTeam()
		th.LinkUserToTeam(U, A)
		A_1 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_OPEN, A.Id)
		th.App.AddUserToChannel(U, A_1, false)
		A_2 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
		th.App.AddUserToChannel(U, A_2, false)
		A_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
		th.App.AddUserToChannel(U, A_3, false)

		B := th.CreateTeam()
		th.LinkUserToTeam(U, B)
		B_1 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_OPEN, B.Id)
		th.App.AddUserToChannel(U, B_1, false)
		B_2 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, B.Id)
		th.App.AddUserToChannel(U, B_2, false)
		B_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, B.Id)
		th.App.AddUserToChannel(U, B_3, false)

		C := th.CreateTeam()
		th.LinkUserToTeam(U, C)
		C_1 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_OPEN, C.Id)
		th.App.AddUserToChannel(U, C_1, false)
		C_2 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, C.Id)
		th.App.AddUserToChannel(U, C_2, false)
		C_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, C.Id)
		th.App.AddUserToChannel(U, C_3, false)

		CreateDmChannel(th, X, Y)
		CreateDmChannel(th, X, U)

		CreateDmChannel(th, Y, Z)
		CreateDmChannel(th, Y, V)

		CreateDmChannel(th, Z, X)
		CreateDmChannel(th, Z, W)

		// testing
		uc, err := th.Server.Store.Channel().GetChannels("", X.Id, true, 0)
		require.NoError(t, err, "Get direct channel error")

		SetPolicy(SimplePolicy{
			period: 1,
			team: SimpleSpecificPolicy{
				B.Id: 2,
				C.Id: 3,
			},
			channel: SimpleSpecificPolicy{

				A_1.Id: 4,
				A_2.Id: 5,
				B_1.Id: 6,
				B_2.Id: 7,
				C_1.Id: 8,
				C_2.Id: 9,
			},
			user: SimpleSpecificPolicy{
				X.Username: 10,
			},
		})
		prObj, err := New(th.App)
		require.NoError(t, err, "should call New succussfully")

		type temppl struct {
			name   string
			id     string
			period time.Duration
		}
		func(pl []temppl) {

			for _, ch := range pl {

				pa, ok := prObj.merged[ch.id]
				assert.Equal(t, true, ok, fmt.Sprintf("%s should be found", ch.name))
				assert.Equal(t, ch.period, pa, fmt.Sprintf("%s period wrong", ch.name))

			}

		}(
			[]temppl{
				{"A_1", A_1.Id, 4},
				{"A_2", A_2.Id, 5},
				{"B_1", B_1.Id, 6},
				{"B_2", B_2.Id, 7},
				{"C_1", C_1.Id, 8},
				{"C_2", C_2.Id, 9},
				{"B_3", B_3.Id, 2},
				{"C_3", C_3.Id, 3},
				{"X", []*model.Channel(*uc)[0].Id, 10},
				{"X", []*model.Channel(*uc)[1].Id, 10},
				{"X", []*model.Channel(*uc)[2].Id, 10},
			},
		)

	})

	t.Run("testing prune root posts and some file", func(t *testing.T) {

		A := th.CreateTeam()
		th.LinkUserToTeam(U, A)
		A_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
		th.App.AddUserToChannel(U, A_3, false)

		const period = 7

		const POSTS_COUNTS = 7

		U.Password = "Pa$$word11"
		LoginWithClient(U, Client)

		fileinfos_backup := []*model.FileInfo{}

		for i := 0; i < POSTS_COUNTS; i++ {
			var fileId string
			if i%2 == 0 {
				fileResp, resp := Client.UploadFile([]byte("data"), A_3.Id, "test"+strconv.Itoa(i))
				require.Nil(t, resp.Error, "expected no error")
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
		stats, err := pr.pruneActions([]string{A_3.Id}, nil, period)
		require.NoError(t, err, "there should be no errors after first pruning.")

		posts, _ := th.App.GetPosts(A_3.Id, 0, 100)
		require.Greaterf(t, len(posts.Posts), 0, "there should be some posts after first pruning.")

		fsNoDel := make(map[string]bool)
		for _, post := range posts.Posts {
			assert.Greater(t, post.UpdateAt, endTime, "left posts should be greater than pruning end time")
			if len([]string(post.FileIds)) != 0 {
				fileinfos, err := th.App.GetFileInfosForPost(post.Id, true)
				require.Nil(t, err)
				for _, fileinfo := range fileinfos {
					b, _ := th.App.FileExists(fileinfo.Path)
					assert.Equalf(t, true, b, "file: %v should exist.", fileinfo.Path)
					fsNoDel[fileinfo.Id] = true

				}
			}
		}

		for _, fileinfo := range fileinfos_backup {

			if _, ok := fsNoDel[fileinfo.Id]; !ok {

				b, err := th.App.FileExists(fileinfo.Path)
				require.Nil(t, err, "expected no error")
				assert.Equalf(t, false, b, "file: %v should be deleted.", fileinfo.Path)

			}

		}

		time.Sleep(3 * time.Second)
		stat2, err := pr.pruneActions([]string{A_3.Id}, nil, 1)
		require.NoError(t, err, "there should be no errors after second pruning.")

		posts, err = th.App.GetPosts(A_3.Id, 0, 100)
		require.Nil(t, err, "expected no error")
		require.Equalf(t, len(posts.Posts), 0, "there should be no posts after second pruning.")

		for _, fileinfo := range fileinfos_backup {

			b, err := th.App.FileExists(fileinfo.Path)
			require.Nil(t, err, "expected no error")
			assert.Equalf(t, false, b, "file: %v should not exist after second pruning.", fileinfo.Path)

		}

		dirs, err := th.App.ListDirectory(".")
		require.Nil(t, err, "expected no error")
		assert.Equalf(t, 0, len(dirs), "there should be no directory after second pruncing. but....%v", dirs)

		// +1 system message
		assert.Equalf(t, 8, stats.OrgRoots+stat2.OrgRoots, "original(true) root is not correct")
		assert.Equalf(t, 0, stats.NonOrgRoots+stat2.NonOrgRoots, "non-original root is not correct")
		assert.Equalf(t, 0, stats.DeletedOrgRoots+stat2.DeletedOrgRoots, "deleted original root is not correct")
		assert.Equalf(t, 0, stats.Pinned+stat2.Pinned, "pinned original root(including thread) is not correct")
		assert.Equalf(t, 8, stats.PruneOrgRoots+stat2.PruneOrgRoots, "prune original root is not correct")
		assert.Equalf(t, 1, stats.System+stat2.System, "system original root is not correct")
		assert.Equalf(t, 0, stats.Threads+stat2.Threads, "threads is not correct")
		assert.Equalf(t, 0, stats.Db_reactions+stat2.Db_reactions, "deleted from Reactions is not correct")
		assert.Equalf(t, 0, stats.Db_threadMem+stat2.Db_threadMem, "deleted from ThreadMemberships is not correct")
		assert.Equalf(t, 0, stats.Db_threads+stat2.Db_threads, "deleted from Threads is not correct")
		assert.Equalf(t, 4, stats.Db_fileInfo+stat2.Db_fileInfo, "deleted from FileInfo is not correct")
		assert.Equalf(t, 0, stats.Db_preference+stat2.Db_preference, "deleted from Preference is not correct")
		assert.Equalf(t, 8, stats.Db_posts+stat2.Db_posts, "deleted from Posts is not correct")

	})

	t.Run("testing prune thread posts and some file", func(t *testing.T) {

		A := th.CreateTeam()
		th.LinkUserToTeam(U, A)
		A_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
		th.App.AddUserToChannel(U, A_3, false)

		const period = 7

		const POSTS_COUNTS = 7

		U.Password = "Pa$$word11"
		LoginWithClient(U, Client)

		fileinfos_backup := []*model.FileInfo{}

		root, resp := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			Message:   "root message",
		})
		require.Nil(t, resp.Error, "expected no error")

		for i := 0; i < POSTS_COUNTS-1; i++ {
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
				RootId:    root.Id,
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

		stats, err := pr.pruneActions([]string{A_3.Id}, nil, period)
		require.NoError(t, err, "there should be no errors after first pruning.")

		posts, _ := th.App.GetPosts(A_3.Id, 0, 100)
		require.GreaterOrEqualf(t, len(posts.Posts), 7, "there should be no deleted posts after first pruning.")

		for _, post := range posts.Posts {
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
		stat2, err := pr.pruneActions([]string{A_3.Id}, nil, 1)
		require.NoError(t, err, "there should be no errors after second pruning.")

		posts, err = th.App.GetPosts(A_3.Id, 0, 100)
		require.Nil(t, err, "expected no error")
		require.Equalf(t, len(posts.Posts), 0, "there should be no posts after second pruning.")

		for _, fileinfo := range fileinfos_backup {

			b, _ := th.App.FileExists(fileinfo.Path)
			assert.Equalf(t, false, b, "file: %v should not exist after second pruning.", fileinfo.Path)

		}

		dirs, _ := th.App.ListDirectory(".")
		assert.Equalf(t, 0, len(dirs), "there should be no directory after second pruncing. but....%v", dirs)

		// +1 system message
		assert.Equalf(t, 2, stats.OrgRoots+stat2.OrgRoots, "original(true) root is not correct")
		assert.Equalf(t, 0, stats.NonOrgRoots+stat2.NonOrgRoots, "non-original root is not correct")
		assert.Equalf(t, 0, stats.DeletedOrgRoots+stat2.DeletedOrgRoots, "deleted original root is not correct")
		assert.Equalf(t, 0, stats.Pinned+stat2.Pinned, "pinned original root(including thread) is not correct")
		assert.Equalf(t, 2, stats.PruneOrgRoots+stat2.PruneOrgRoots, "prune original root is not correct")
		assert.Equalf(t, 1, stats.System+stat2.System, "system original root is not correct")
		assert.Equalf(t, 6, stats.Threads+stat2.Threads, "threads is not correct")
		assert.Equalf(t, 0, stats.Db_reactions+stat2.Db_reactions, "deleted from Reactions is not correct")
		assert.Equalf(t, 1, stats.Db_threadMem+stat2.Db_threadMem, "deleted from ThreadMemberships is not correct")
		assert.Equalf(t, 1, stats.Db_threads+stat2.Db_threads, "deleted from Threads is not correct")
		assert.Equalf(t, 3, stats.Db_fileInfo+stat2.Db_fileInfo, "deleted from FileInfo is not correct")
		assert.Equalf(t, 0, stats.Db_preference+stat2.Db_preference, "deleted from Preference is not correct")
		assert.Equalf(t, 8, stats.Db_posts+stat2.Db_posts, "deleted from Posts is not correct")
	})

	t.Run("testing prune root pinned posts with files", func(t *testing.T) {

		A := th.CreateTeam()
		th.LinkUserToTeam(U, A)
		A_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
		th.App.AddUserToChannel(U, A_3, false)

		U.Password = "Pa$$word11"
		LoginWithClient(U, Client)

		fileResp, resp := Client.UploadFile([]byte("data"), A_3.Id, "test1")
		require.Nil(t, resp.Error, "expected no error")
		fileId := fileResp.FileInfos[0].Id
		require.NotEmptyf(t, fileId, "upload file should not be empty.", fileResp.ToJson())

		post, resp := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			Message:   "pinned root message",
			FileIds:   model.StringArray{fileId},
		})

		require.Nil(t, resp.Error, "expected no error")

		fileinfo, err := th.App.GetFileInfosForPost((*post).Id, true)
		require.Nil(t, err)

		b, resp := Client.PinPost(post.Id)
		require.Nil(t, resp.Error, "expected no error")
		require.Equalf(t, true, b, "pin should be sucessful.")

		time.Sleep(2 * time.Second)

		pr, err1 := New(th.App)
		require.NoError(t, err1)

		stats, err1 := pr.pruneActions([]string{A_3.Id}, nil, 1)
		require.NoError(t, err1, "there should be no errors after first pruning.")

		posts, _ := th.App.GetPosts(A_3.Id, 0, 100)
		assert.GreaterOrEqualf(t, len(posts.Posts), 1, "pinned post should not be delete.")

		b, err = th.App.FileExists(fileinfo[0].Path)
		require.Nil(t, err, "expected no error")
		assert.Equalf(t, true, b, "file: %v should not be deleted.", fileinfo[0].Path)

		// +1 system message
		assert.Equalf(t, 2, stats.OrgRoots, "original(true) root is not correct")
		assert.Equalf(t, 1, stats.NonOrgRoots, "non-original root is not correct")
		assert.Equalf(t, 0, stats.DeletedOrgRoots, "deleted original root is not correct")
		assert.Equalf(t, 1, stats.Pinned, "pinned original root(including thread) is not correct")
		assert.Equalf(t, 1, stats.PruneOrgRoots, "prune original root is not correct")
		assert.Equalf(t, 1, stats.System, "system original root is not correct")
		assert.Equalf(t, 0, stats.Threads, "threads is not correct")
		assert.Equalf(t, 0, stats.Db_reactions, "deleted from Reactions is not correct")
		assert.Equalf(t, 0, stats.Db_threadMem, "deleted from ThreadMemberships is not correct")
		assert.Equalf(t, 0, stats.Db_threads, "deleted from Threads is not correct")
		assert.Equalf(t, 0, stats.Db_fileInfo, "deleted from FileInfo is not correct")
		assert.Equalf(t, 0, stats.Db_preference, "deleted from Preference is not correct")
		assert.Equalf(t, 1, stats.Db_posts, "deleted from Posts is not correct")
	})

	t.Run("testing prune root unpinned posts", func(t *testing.T) {

		A := th.CreateTeam()
		th.LinkUserToTeam(U, A)
		A_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
		th.App.AddUserToChannel(U, A_3, false)

		U.Password = "Pa$$word11"
		LoginWithClient(U, Client)

		fileResp, resp := Client.UploadFile([]byte("data"), A_3.Id, "test1")
		require.Nil(t, resp.Error, "expected no error")
		fileId := fileResp.FileInfos[0].Id
		require.NotEmptyf(t, fileId, "upload file should not be empty.", fileResp.ToJson())

		post, resp := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			Message:   "pinned root message",
			FileIds:   model.StringArray{fileId},
		})

		require.Nil(t, resp.Error, "expected no error")

		fileinfo, err := th.App.GetFileInfosForPost((*post).Id, true)
		require.Nil(t, err)

		b, resp := Client.PinPost(post.Id)
		require.Nil(t, resp.Error, "expected no error")
		require.Equalf(t, true, b, "pin should be sucessful.")

		b, resp = Client.UnpinPost(post.Id)
		require.Nil(t, resp.Error, "expected no error")
		require.Equalf(t, true, b, "unpin should be sucessful.")

		time.Sleep(2 * time.Second)

		pr, err1 := New(th.App)
		require.NoError(t, err1)

		stats, err1 := pr.pruneActions([]string{A_3.Id}, nil, 1)
		require.NoError(t, err1, "there should be no errors after first pruning.")

		posts, _ := th.App.GetPosts(A_3.Id, 0, 100)
		assert.Equalf(t, len(posts.Posts), 0, "unpinned post should  be delete.")

		b, err = th.App.FileExists(fileinfo[0].Path)
		require.Nil(t, err, "expected no error")
		assert.Equalf(t, false, b, "file: %v should  be deleted.", fileinfo[0].Path)

		// +1 system message
		assert.Equalf(t, 2, stats.OrgRoots, "original(true) root is not correct")
		assert.Equalf(t, 2, stats.NonOrgRoots, "non-original root is not correct")
		assert.Equalf(t, 0, stats.DeletedOrgRoots, "deleted original root is not correct")
		assert.Equalf(t, 0, stats.Pinned, "pinned original root(including thread) is not correct")
		assert.Equalf(t, 2, stats.PruneOrgRoots, "prune original root is not correct")
		assert.Equalf(t, 1, stats.System, "system original root is not correct")
		assert.Equalf(t, 0, stats.Threads, "threads is not correct")
		assert.Equalf(t, 0, stats.Db_reactions, "deleted from Reactions is not correct")
		assert.Equalf(t, 0, stats.Db_threadMem, "deleted from ThreadMemberships is not correct")
		assert.Equalf(t, 0, stats.Db_threads, "deleted from Threads is not correct")
		assert.Equalf(t, 1, stats.Db_fileInfo, "deleted from FileInfo is not correct")
		assert.Equalf(t, 0, stats.Db_preference, "deleted from Preference is not correct")
		assert.Equalf(t, 4, stats.Db_posts, "deleted from Posts is not correct")
	})
	t.Run("testing prune thread pinned posts", func(t *testing.T) {

		A := th.CreateTeam()
		th.LinkUserToTeam(U, A)
		A_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
		th.App.AddUserToChannel(U, A_3, false)

		U.Password = "Pa$$word11"
		LoginWithClient(U, Client)

		root, resp := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			Message:   "root message with threads",
		})

		require.Nil(t, resp.Error, "expected no error")

		_, resp = Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			RootId:    root.Id,
		})
		require.Nil(t, resp.Error, "expected no error")

		thread, resp := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			RootId:    root.Id,
		})
		require.Nil(t, resp.Error, "expected no error")

		b, resp := Client.PinPost(thread.Id)
		require.Nil(t, resp.Error, "expected no error")
		require.Equalf(t, true, b, "pin should be sucessful.")

		time.Sleep(2 * time.Second)

		pr, err := New(th.App)
		require.NoError(t, err)

		stats, err := pr.pruneActions([]string{A_3.Id}, nil, 1)
		require.NoError(t, err, "there should be no errors after first pruning.")

		posts, _ := th.App.GetPosts(A_3.Id, 0, 100)
		assert.GreaterOrEqualf(t, len(posts.Posts), 1, "pinned threads and their family should not be delete.")

		assert.Equalf(t, 2, stats.OrgRoots, "original(true) root is not correct")
		assert.Equalf(t, 0, stats.NonOrgRoots, "non-original root is not correct")
		assert.Equalf(t, 0, stats.DeletedOrgRoots, "deleted original root is not correct")
		assert.Equalf(t, 1, stats.Pinned, "pinned original root(including thread) is not correct")
		assert.Equalf(t, 1, stats.PruneOrgRoots, "prune original root is not correct")
		assert.Equalf(t, 1, stats.System, "system original root is not correct")
		assert.Equalf(t, 0, stats.Threads, "threads is not correct")
		assert.Equalf(t, 0, stats.Db_reactions, "deleted from Reactions is not correct")
		assert.Equalf(t, 0, stats.Db_threadMem, "deleted from ThreadMemberships is not correct")
		assert.Equalf(t, 0, stats.Db_threads, "deleted from Threads is not correct")
		assert.Equalf(t, 0, stats.Db_fileInfo, "deleted from FileInfo is not correct")
		assert.Equalf(t, 0, stats.Db_preference, "deleted from Preference is not correct")
		assert.Equalf(t, 1, stats.Db_posts, "deleted from Posts is not correct")

	})

	t.Run("testing prune thread unpinned posts", func(t *testing.T) {

		A := th.CreateTeam()
		th.LinkUserToTeam(U, A)
		A_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
		th.App.AddUserToChannel(U, A_3, false)

		U.Password = "Pa$$word11"
		LoginWithClient(U, Client)

		root, resp := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			Message:   "root message with threads",
		})

		require.Nil(t, resp.Error, "expected no error")

		_, resp = Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			RootId:    root.Id,
		})
		require.Nil(t, resp.Error, "expected no error")

		thread, resp := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			RootId:    root.Id,
		})
		require.Nil(t, resp.Error, "expected no error")

		b, resp := Client.PinPost(thread.Id)
		require.Nil(t, resp.Error, "expected no error")
		require.Equalf(t, true, b, "pin should be sucessful.")

		b, resp = Client.UnpinPost(thread.Id)
		require.Nil(t, resp.Error, "expected no error")
		require.Equalf(t, true, b, "pin should be sucessful.")

		time.Sleep(2 * time.Second)

		pr, err := New(th.App)
		require.NoError(t, err)

		stats, err := pr.pruneActions([]string{A_3.Id}, nil, 1)
		require.NoError(t, err, "there should be no errors after first pruning.")

		posts, _ := th.App.GetPosts(A_3.Id, 0, 100)
		assert.GreaterOrEqualf(t, len(posts.Posts), 1, "pinned threads and their family should not be delete.")

		assert.Equalf(t, 2, stats.OrgRoots, "original(true) root is not correct")
		assert.Equalf(t, 0, stats.NonOrgRoots, "non-original root is not correct")
		assert.Equalf(t, 0, stats.DeletedOrgRoots, "deleted original root is not correct")
		assert.Equalf(t, 0, stats.Pinned, "pinned original root(including thread) is not correct")
		assert.Equalf(t, 2, stats.PruneOrgRoots, "prune original root is not correct")
		assert.Equalf(t, 1, stats.System, "system original root is not correct")
		assert.Equalf(t, 3, stats.Threads, "threads is not correct")
		assert.Equalf(t, 0, stats.Db_reactions, "deleted from Reactions is not correct")
		assert.Equalf(t, 1, stats.Db_threadMem, "deleted from ThreadMemberships is not correct")
		assert.Equalf(t, 1, stats.Db_threads, "deleted from Threads is not correct")
		assert.Equalf(t, 0, stats.Db_fileInfo, "deleted from FileInfo is not correct")
		assert.Equalf(t, 0, stats.Db_preference, "deleted from Preference is not correct")
		assert.Equalf(t, 1, stats.Db_posts, "deleted from Posts is not correct")

	})
	t.Run("testing prune deleted root with files for permanent case", func(t *testing.T) {

		A := th.CreateTeam()
		th.LinkUserToTeam(U, A)
		A_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
		th.App.AddUserToChannel(U, A_3, false)

		U.Password = "Pa$$word11"
		LoginWithClient(U, Client)

		fileResp, resp := Client.UploadFile([]byte("data"), A_3.Id, "test1")
		root, resp := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			Message:   "root message",
			FileIds:   model.StringArray{fileResp.FileInfos[0].Id},
		})
		require.Nil(t, resp.Error, "expected no error")
		fileinfo, err := th.App.GetFileInfosForPost((*root).Id, true)

		b, resp := Client.DeletePost(root.Id)
		require.Nil(t, resp.Error, "expected no error")
		require.Equalf(t, true, b, "delete should be sucessful.")

		time.Sleep(2 * time.Second)

		pr, err1 := New(th.App)
		require.NoError(t, err1)

		stats, err1 := pr.pruneActions([]string{A_3.Id}, nil, 0)
		require.NoError(t, err1, "there should be no errors after first pruning.")

		posts, _ := th.App.GetPosts(A_3.Id, 0, 100)
		assert.Equalf(t, len(posts.Posts), 1, "deleted root should be pruned.")

		b, err = th.App.FileExists(fileinfo[0].Path)
		require.Nil(t, err, "expected no error")
		assert.Equalf(t, false, b, "file: %v should  be deleted. %v", fileinfo[0].Path)

		// +1 system message
		assert.Equalf(t, 2, stats.OrgRoots, "original(true) root is not correct")
		assert.Equalf(t, 0, stats.NonOrgRoots, "non-original root is not correct")
		assert.Equalf(t, 1, stats.DeletedOrgRoots, "deleted original root is not correct")
		assert.Equalf(t, 0, stats.Pinned, "pinned original root(including thread) is not correct")
		assert.Equalf(t, 1, stats.PruneOrgRoots, "prune original root is not correct")
		assert.Equalf(t, 1, stats.System, "system original root is not correct")
		assert.Equalf(t, 0, stats.Threads, "threads is not correct")
		assert.Equalf(t, 0, stats.Db_reactions, "deleted from Reactions is not correct")
		assert.Equalf(t, 0, stats.Db_threadMem, "deleted from ThreadMemberships is not correct")
		assert.Equalf(t, 0, stats.Db_threads, "deleted from Threads is not correct")
		assert.Equalf(t, 1, stats.Db_fileInfo, "deleted from FileInfo is not correct")
		assert.Equalf(t, 0, stats.Db_preference, "deleted from Preference is not correct")
		assert.Equalf(t, 1, stats.Db_posts, "deleted from Posts is not correct")
	})

	t.Run("testing prune deleted root for specific case", func(t *testing.T) {

		A := th.CreateTeam()
		th.LinkUserToTeam(U, A)
		A_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
		th.App.AddUserToChannel(U, A_3, false)

		U.Password = "Pa$$word11"
		LoginWithClient(U, Client)

		root, resp := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			Message:   "root message",
		})
		require.Nil(t, resp.Error, "expected no error")

		b, resp := Client.DeletePost(root.Id)
		require.Nil(t, resp.Error, "expected no error")
		require.Equalf(t, true, b, "delete should be sucessful.")

		time.Sleep(2 * time.Second)

		pr, err1 := New(th.App)
		require.NoError(t, err1)

		stats, err1 := pr.pruneActions([]string{A_3.Id}, nil, 1)
		require.NoError(t, err1, "there should be no errors after first pruning.")

		// +1 system message
		assert.Equalf(t, 2, stats.OrgRoots, "original(true) root is not correct")
		assert.Equalf(t, 0, stats.NonOrgRoots, "non-original root is not correct")
		assert.Equalf(t, 1, stats.DeletedOrgRoots, "deleted original root is not correct")
		assert.Equalf(t, 0, stats.Pinned, "pinned original root(including thread) is not correct")
		assert.Equalf(t, 2, stats.PruneOrgRoots, "prune original root is not correct")
		assert.Equalf(t, 1, stats.System, "system original root is not correct")
		assert.Equalf(t, 0, stats.Threads, "threads is not correct")
		assert.Equalf(t, 0, stats.Db_reactions, "deleted from Reactions is not correct")
		assert.Equalf(t, 0, stats.Db_threadMem, "deleted from ThreadMemberships is not correct")
		assert.Equalf(t, 0, stats.Db_threads, "deleted from Threads is not correct")
		assert.Equalf(t, 0, stats.Db_fileInfo, "deleted from FileInfo is not correct")
		assert.Equalf(t, 0, stats.Db_preference, "deleted from Preference is not correct")
		assert.Equalf(t, 2, stats.Db_posts, "deleted from Posts is not correct")
	})
	t.Run("testing prune deleted thread for permanent case", func(t *testing.T) {

		A := th.CreateTeam()
		th.LinkUserToTeam(U, A)
		A_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
		th.App.AddUserToChannel(U, A_3, false)

		U.Password = "Pa$$word11"
		LoginWithClient(U, Client)

		root, _ := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			Message:   "root message",
		})

		thread, _ := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			Message:   "thread message",
			RootId:    root.Id,
		})

		Client.DeletePost(thread.Id)

		time.Sleep(2 * time.Second)

		pr, err1 := New(th.App)
		require.NoError(t, err1)

		stats, err1 := pr.pruneActions([]string{A_3.Id}, nil, 0)
		require.NoError(t, err1, "there should be no errors after first pruning.")

		// +1 system message
		assert.Equalf(t, 2, stats.OrgRoots, "original(true) root is not correct")
		assert.Equalf(t, 0, stats.NonOrgRoots, "non-original root is not correct")
		assert.Equalf(t, 0, stats.DeletedOrgRoots, "deleted original root is not correct")
		assert.Equalf(t, 0, stats.Pinned, "pinned original root(including thread) is not correct")
		assert.Equalf(t, 0, stats.PruneOrgRoots, "prune original root is not correct")
		assert.Equalf(t, 1, stats.System, "system original root is not correct")
		assert.Equalf(t, 0, stats.Threads, "threads is not correct")
		assert.Equalf(t, 0, stats.Db_reactions, "deleted from Reactions is not correct")
		assert.Equalf(t, 0, stats.Db_threadMem, "deleted from ThreadMemberships is not correct")
		assert.Equalf(t, 0, stats.Db_threads, "deleted from Threads is not correct")
		assert.Equalf(t, 0, stats.Db_fileInfo, "deleted from FileInfo is not correct")
		assert.Equalf(t, 0, stats.Db_preference, "deleted from Preference is not correct")
		assert.Equalf(t, 0, stats.Db_posts, "deleted from Posts is not correct")
	})
	t.Run("testing prune deleted thread for specific case", func(t *testing.T) {

		A := th.CreateTeam()
		th.LinkUserToTeam(U, A)
		A_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
		th.App.AddUserToChannel(U, A_3, false)

		U.Password = "Pa$$word11"
		LoginWithClient(U, Client)

		root, _ := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			Message:   "root message",
		})

		thread, _ := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			Message:   "thread message",
			RootId:    root.Id,
		})

		Client.DeletePost(thread.Id)

		time.Sleep(2 * time.Second)

		pr, err1 := New(th.App)
		require.NoError(t, err1)

		stats, err1 := pr.pruneActions([]string{A_3.Id}, nil, 1)
		require.NoError(t, err1, "there should be no errors after first pruning.")

		// +1 system message
		assert.Equalf(t, 2, stats.OrgRoots, "original(true) root is not correct")
		assert.Equalf(t, 0, stats.NonOrgRoots, "non-original root is not correct")
		assert.Equalf(t, 0, stats.DeletedOrgRoots, "deleted original root is not correct")
		assert.Equalf(t, 0, stats.Pinned, "pinned original root(including thread) is not correct")
		assert.Equalf(t, 2, stats.PruneOrgRoots, "prune original root is not correct")
		assert.Equalf(t, 1, stats.System, "system original root is not correct")
		assert.Equalf(t, 1, stats.Threads, "threads is not correct")
		assert.Equalf(t, 0, stats.Db_reactions, "deleted from Reactions is not correct")
		assert.Equalf(t, 1, stats.Db_threadMem, "deleted from ThreadMemberships is not correct")
		assert.Equalf(t, 1, stats.Db_threads, "deleted from Threads is not correct")
		assert.Equalf(t, 0, stats.Db_fileInfo, "deleted from FileInfo is not correct")
		assert.Equalf(t, 0, stats.Db_preference, "deleted from Preference is not correct")
		assert.Equalf(t, 3, stats.Db_posts, "deleted from Posts is not correct")
	})
	t.Run("testing not prune edit root", func(t *testing.T) {

		A := th.CreateTeam()
		th.LinkUserToTeam(U, A)
		A_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
		th.App.AddUserToChannel(U, A_3, false)

		U.Password = "Pa$$word11"
		LoginWithClient(U, Client)

		root, _ := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			Message:   "root message",
		})

		msg := "root message edited"
		Client.PatchPost(root.Id, &model.PostPatch{
			Message: &msg,
		})

		time.Sleep(2 * time.Second)

		pr, err1 := New(th.App)
		require.NoError(t, err1)

		stats, err1 := pr.pruneActions([]string{A_3.Id}, nil, 1)
		require.NoError(t, err1, "there should be no errors after first pruning.")

		// +1 system message
		assert.Equalf(t, 2, stats.OrgRoots, "original(true) root is not correct")
		assert.Equalf(t, 1, stats.NonOrgRoots, "non-original root is not correct")
		assert.Equalf(t, 0, stats.DeletedOrgRoots, "deleted original root is not correct")
		assert.Equalf(t, 0, stats.Pinned, "pinned original root(including thread) is not correct")
		assert.Equalf(t, 2, stats.PruneOrgRoots, "prune original root is not correct")
		assert.Equalf(t, 1, stats.System, "system original root is not correct")
		assert.Equalf(t, 0, stats.Threads, "threads is not correct")
		assert.Equalf(t, 0, stats.Db_reactions, "deleted from Reactions is not correct")
		assert.Equalf(t, 0, stats.Db_threadMem, "deleted from ThreadMemberships is not correct")
		assert.Equalf(t, 0, stats.Db_threads, "deleted from Threads is not correct")
		assert.Equalf(t, 0, stats.Db_fileInfo, "deleted from FileInfo is not correct")
		assert.Equalf(t, 0, stats.Db_preference, "deleted from Preference is not correct")
		assert.Equalf(t, 3, stats.Db_posts, "deleted from Posts is not correct")
	})
	t.Run("testing not prune edit thread", func(t *testing.T) {

		A := th.CreateTeam()
		th.LinkUserToTeam(U, A)
		A_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
		th.App.AddUserToChannel(U, A_3, false)

		U.Password = "Pa$$word11"
		LoginWithClient(U, Client)

		root, _ := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			Message:   "root message",
		})
		thread, _ := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			Message:   "root message",
			RootId:    root.Id,
		})

		msg := "thread message edited"
		Client.PatchPost(thread.Id, &model.PostPatch{
			Message: &msg,
		})

		time.Sleep(2 * time.Second)

		pr, err1 := New(th.App)
		require.NoError(t, err1)

		stats, err1 := pr.pruneActions([]string{A_3.Id}, nil, 1)
		require.NoError(t, err1, "there should be no errors after first pruning.")

		// +1 system message
		assert.Equalf(t, 2, stats.OrgRoots, "original(true) root is not correct")
		assert.Equalf(t, 0, stats.NonOrgRoots, "non-original root is not correct")
		assert.Equalf(t, 0, stats.DeletedOrgRoots, "deleted original root is not correct")
		assert.Equalf(t, 0, stats.Pinned, "pinned original root(including thread) is not correct")
		assert.Equalf(t, 2, stats.PruneOrgRoots, "prune original root is not correct")
		assert.Equalf(t, 1, stats.System, "system original root is not correct")
		assert.Equalf(t, 2, stats.Threads, "threads is not correct")
		assert.Equalf(t, 0, stats.Db_reactions, "deleted from Reactions is not correct")
		assert.Equalf(t, 1, stats.Db_threadMem, "deleted from ThreadMemberships is not correct")
		assert.Equalf(t, 1, stats.Db_threads, "deleted from Threads is not correct")
		assert.Equalf(t, 0, stats.Db_fileInfo, "deleted from FileInfo is not correct")
		assert.Equalf(t, 0, stats.Db_preference, "deleted from Preference is not correct")
		assert.Equalf(t, 4, stats.Db_posts, "deleted from Posts is not correct")
	})
	t.Run("testing flag/reaction", func(t *testing.T) {

		A := th.CreateTeam()
		th.LinkUserToTeam(U, A)
		A_3 := th.CreateChannelWithClientAndTeam(Client, model.CHANNEL_PRIVATE, A.Id)
		th.App.AddUserToChannel(U, A_3, false)

		U.Password = "Pa$$word11"
		LoginWithClient(U, Client)

		root, _ := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			Message:   "root message",
		})

		Client.SaveReaction(&model.Reaction{
			PostId:    root.Id,
			EmojiName: "innocent",
		})

                Client.DeleteReaction{}
		Client.UpdatePreferences(U.Id, &model.Preferences{

			model.Preference{
				Category: "flagged_post",
				Name:     root.Id},
		})

		thread, _ := Client.CreatePost(&model.Post{
			ChannelId: A_3.Id,
			Message:   "root message",
			RootId:    root.Id,
		})

		time.Sleep(2 * time.Second)

		pr, err1 := New(th.App)
		require.NoError(t, err1)

		stats, err1 := pr.pruneActions([]string{A_3.Id}, nil, 1)
		require.NoError(t, err1, "there should be no errors after first pruning.")

		// +1 system message
		assert.Equalf(t, 2, stats.OrgRoots, "original(true) root is not correct")
		assert.Equalf(t, 0, stats.NonOrgRoots, "non-original root is not correct")
		assert.Equalf(t, 0, stats.DeletedOrgRoots, "deleted original root is not correct")
		assert.Equalf(t, 0, stats.Pinned, "pinned original root(including thread) is not correct")
		assert.Equalf(t, 2, stats.PruneOrgRoots, "prune original root is not correct")
		assert.Equalf(t, 1, stats.System, "system original root is not correct")
		assert.Equalf(t, 2, stats.Threads, "threads is not correct")
		assert.Equalf(t, 0, stats.Db_reactions, "deleted from Reactions is not correct")
		assert.Equalf(t, 1, stats.Db_threadMem, "deleted from ThreadMemberships is not correct")
		assert.Equalf(t, 1, stats.Db_threads, "deleted from Threads is not correct")
		assert.Equalf(t, 0, stats.Db_fileInfo, "deleted from FileInfo is not correct")
		assert.Equalf(t, 0, stats.Db_preference, "deleted from Preference is not correct")
		assert.Equalf(t, 4, stats.Db_posts, "deleted from Posts is not correct")
	})
	t.Run("testing exception", func(t *testing.T) {
        }
}
