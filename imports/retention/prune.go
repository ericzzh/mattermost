package retention

import (
	"encoding/json"
	"fmt"

	"github.com/mattermost/mattermost-server/v5/app"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/shared/mlog"

	"github.com/mattermost/mattermost-server/v5/store/sqlstore"
	"time"
)

func Prune(a *app.App) error {

	s := a.Srv()
	st, _ := json.Marshal(s.Config().SqlSettings)
	mlog.Debug(string(st))

	ss := sqlstore.New(s.Config().SqlSettings, nil)

	// var posts []*model.Post
	// var fetchQuery string
	// if skipFetchThreads {
	// 	fetchQuery = "SELECT p.*, (SELECT COUNT(Posts.Id) FROM Posts WHERE Posts.RootId = (CASE WHEN p.RootId = '' THEN p.Id ELSE p.RootId END) AND Posts.DeleteAt = 0) as ReplyCount FROM Posts p WHERE ChannelId = :ChannelId AND DeleteAt = 0 ORDER BY CreateAt DESC LIMIT :Limit OFFSET :Offset"
	// } else {
	// 	fetchQuery = "SELECT * FROM Posts WHERE ChannelId = :ChannelId AND DeleteAt = 0 ORDER BY CreateAt DESC LIMIT :Limit OFFSET :Offset"
	// }
	// _, err := s.GetReplica().Select(&posts, fetchQuery, map[string]interface{}{"ChannelId": channelId, "Offset": offset, "Limit": limit})
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failed to find Posts")
	// }
	// return posts, nil

	// 	allTeams, err := a.GetAllTeams()
	//
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	if len(allTeams) == 0 {
	// 		mlog.Info("Prune: No teams found")
	// 		return nil
	// 	}
	//
	// 	for _, team := range allTeams {
	// 		mlog.Debug(fmt.Sprintf("Prune: Processing Team: %v", team.Name))
	//
	//                 err := pruneGeneral(team)
	//                 if err != nil {
	//                      return err
	//                 }
	//
	// 	}

	// 	now := time.Now()
	//         s := srv.sqlStore
	//
	//
	//         // general case
	// 	endat := time.Add(-time.Second * policy.period)
	return nil

}

func pruneGeneral(ss *sqlstore.SqlStore) error {

	now := time.Now()
	endTime := now.Add(-time.Second * policy.period)
	var posts []*model.Post

	ss.GetMaster().Select(&posts, `Select * From Posts Where UpdateAt < :EndTime`,
		map[string]interface{}{"EndTime": endTime})

	return nil

}
