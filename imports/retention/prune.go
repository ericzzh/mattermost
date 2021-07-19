package retention

import (
	// "encoding/json"
	"fmt"
	"path/filepath"

	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"github.com/mattermost/gorp"
	"github.com/mattermost/mattermost-server/v5/app"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/shared/mlog"
	"github.com/pkg/errors"

	"time"

	"github.com/mattermost/mattermost-server/v5/store/sqlstore"
)

type Prune struct {
	app      *app.App
	srv      *app.Server
	sqlstore *sqlstore.SqlStore
	merged   SimpleSpecificPolicy
}

func New(a *app.App) (*Prune, error) {

	s := a.Srv()
	// st, _ := json.Marshal(s.Config().SqlSettings)
	// mlog.Debug(string(st))

	ss := sqlstore.New(s.Config().SqlSettings, nil)
	merged, err := mergeToChannels(s)
	if err != nil {
		return nil, errors.Wrapf(err, "Prune: call mergeToChannels wrong.")
	}
	return &Prune{
		app:      a,
		srv:      s,
		sqlstore: ss,
		merged:   merged,
	}, nil

}
func mergeToChannels(srv *app.Server) (mergedChMap SimpleSpecificPolicy, err error) {
	//from specific case the general case

	mergedChMap = SimpleSpecificPolicy{}
	// channels specific rules
	for k, p := range policy.channel {
		mlog.Debug(fmt.Sprintf("Prune: merging channel %s, period %d", k, p))
		mergedChMap[k] = p
	}

	// user directed channel
	for u, p := range policy.user {
		usr, err := srv.Store.User().GetByUsername(u)
		if err != nil {
			return nil, errors.Wrapf(err, "Prune: get user(%s) id wrong.", u)
		}
		chs, err := srv.Store.Channel().GetChannels("", usr.Id, true, 0)
		if err != nil {
			return nil, errors.Wrapf(err, "Prune: get user(%s) direct channel wrong.", u)
		}

		for _, ch := range *chs {
			mlog.Debug(fmt.Sprintf("Prune: merging user %s channel %s, period %d", u, ch.Id, p))
			mergedChMap[ch.Id] = p

		}
	}

	// teams channel
	// mlog.Debug(fmt.Sprintf("Prune: teams %v", policy.team))
	for k, p := range policy.team {

		chs, err := srv.Store.Channel().GetTeamChannels(k)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to call GetTeamChannels()")
		}

		// mlog.Debug(fmt.Sprintf("Prune: print all team %v, channels. %v", k, *chs))
		for _, ch := range *chs {
			// not overwrite the specific channel
			if _, ok := mergedChMap[ch.Id]; !ok {
				mlog.Debug(fmt.Sprintf("Prune: merging team %s channel %s, period %d", k, ch.Id, p))
				mergedChMap[ch.Id] = p
			}

		}
	}

	return mergedChMap, nil
}
func (pr *Prune) Prune() error {

	mlog.Debug("Prune: Staring prune channel posts.")
	if err := pr.pruneGeneral(); err != nil {
		return errors.Wrapf(err, "Prune: call Prune wrong.")
	}

	mlog.Debug("Prune: general case completed.")

	for chid, p := range pr.merged {

		if err := pr.pruneActions([]string{chid}, nil, p); err != nil {
			return errors.Wrapf(err, "failed to call pruneActions().")
		}
		mlog.Debug(fmt.Sprintf("Prune: specific case, channel: %s, completed.", chid))

	}
	return nil

}
func (pr *Prune) pruneGeneral() error {
	ex := pr.fetchAllChannelIds(pr.merged)

	if err := pr.pruneActions(nil, ex, policy.period); err != nil {
		return errors.Wrapf(err, "failed to call pruneActions().")
	}

	return nil
}

func (pr *Prune) fetchAllChannelIds(chsMap SimpleSpecificPolicy) (chIds []string) {
	for id := range chsMap {
		chIds = append(chIds, id)

	}
	return chIds
}

// TO DO: Only select necessary fields
func (pr *Prune) pruneActions(ch []string, ex []string, period time.Duration) error {

	ss := pr.sqlstore

	now := time.Now()
	endTime := model.GetMillisForTime(now.Add(-time.Second * period))

	mlog.Info(fmt.Sprintf("Prune: endTime: %v", endTime))

	//----------------------------------------
	//   Root post fetching
	//----------------------------------------
	var roots []*model.Post

	// Fetch all the root messages
	// we use UpdateAt as a key, because all thread updated (add, pin) will update this field.

	// *** If we use limit, we must fetch all true root post we need, any roots to be further check&filter, will
	//     cause repeat processing ***

	builder := getQueryBuilder(ss)
	sql := builder.Select("*").From("Posts")

	if ch != nil {
		sql = sql.Where(sq.Eq{"ChannelId": ch})
	}

	sql = sql.Where("UpdateAt < ?", endTime)

	if period == 0 {
		// Only clean the Deleted post in  pemanent channls
		// To follow the root -> thread logic, we only fetch the deleted root
		// and threads fetch follows the same logic
		// This logic works because if a root is deleted ,all the subsequent thread will also be deleted.
		// No pinned filtered out, all should be cleaned
		sql = sql.Where("RootId = '' AND DeleteAt <> 0")
	} else {
		// OriginalId is not a key, so fetching the sub-root using this OriginalId is not perfomant, we left them out
		// We should aslo fetch the deleted root id
		sql = sql.Where("(RootId = '' AND IsPinned = 0 AND DeleteAt = 0 ) OR (RootId ='' AND DeleteAt <> 0)") //.Where("OriginalId = ?", "")

		if ex != nil {
			sql = sql.Where(sq.NotEq{"ChannelId": ex})
		}
	}

	sqlstr, args, err := sql.ToSql()
	if err != nil {
		return errors.Wrapf(err, "failed to build root candidate sql string.")
	}

	// mlog.Debug(fmt.Sprintf("Prune: roots fetching sql string: %v, args: %v", sqlstr, args))

	_, err = ss.GetMaster().Select(&roots, sqlstr, args...)

	if err != nil {
		return errors.Wrapf(err, "failed to fetch all the candidate roots.")
	}

	mlog.Info(fmt.Sprintf("Prune: cadidate root counts :%v", len(roots)))

	//----------------------------------------
	//   Thread posts fetching
	//----------------------------------------

	// Fetch all sub-Root(i.e. edit, pin to root) and Threads from root
	// using ROOTID and ORIGINIALID
	// We should also filter all the DeletedAt field to fetch the active posts
	// We can't use the Deleted record to justify.
	chMap := map[string]bool{}
	rootsid := make([]string, 0)
	for _, root := range roots {
		//  Thread is not link to sub-root
		if root.OriginalId == "" {
			rootsid = append(rootsid, root.Id)
			// save the channel
			// Don't neend to save thread's channel, because they must be the same
			chMap[root.ChannelId] = true
		}
	}

	//
	// sql = builder.Select("*").From("Posts").Where(sq.Or{sq.Eq{"RootId": rootsid}, sq.Eq{"OriginalId": rootsid}})
	sql = builder.Select("*").From("Posts").Where(sq.Eq{"RootId": rootsid})
	sqlstr, args, err = sql.ToSql()
	if err != nil {
		return errors.Wrapf(err, "failed to build candidate threads fetching sql string.")
	}
	var threadsCand []*model.Post
	_, err = ss.GetMaster().Select(&threadsCand, sqlstr, args...)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch candidate threads.")
	}

	mlog.Info(fmt.Sprintf("Prune: cadidate threads counts :%v", len(threadsCand)))

	// *** May the following logic is never executed, because the Pinned Thread will update root UpdateAt too
	//     but we leave them as comment out for some possible future change to be easy ***
	// 	rootsMap := map[string][]*model.Post{}
	// 	for _, post := range threadsCand {
	// 		r := post.RootId
	//
	// 		if _, ok := rootsMap[r]; !ok {
	// 			rootsMap[r] = []*model.Post{}
	// 		}
	// 		rootsMap[r] = append(rootsMap[r], post)
	//
	//                 //If there is Pinned thread of a root, all the root will be out
	// 		if post.IsPinned && post.DeleteAt != 0 {
	// 			delete(rootsMap, r)
	// 		}
	// 	}

	threads := threadsCand

	//----------------------------------------
	//   Get all post id and other key information
	//----------------------------------------
	delIds := []string{}
	fileIdMap := map[string]bool{}
	reactionMap := map[string]bool{}
	for _, root := range roots {
		delIds = append(delIds, root.Id)
		if len(root.FileIds) != 0 {
			for _, fileid := range []string(root.FileIds) {
				fileIdMap[fileid] = true
			}
		}

		if root.HasReactions {
			reactionMap[root.Id] = true
		}
	}

	for _, thread := range threads {
		delIds = append(delIds, thread.Id)
		if len(thread.FileIds) != 0 {
			for _, fileid := range thread.FileIds {
				fileIdMap[fileid] = true
			}
		}

		if thread.HasReactions {
			reactionMap[thread.Id] = true
		}

	}

	mlog.Info(fmt.Sprintf("Prune: found %d FileInfo Ids", len(fileIdMap)))
	mlog.Info(fmt.Sprintf("Prune: found %d Reaction post Ids", len(reactionMap)))

	// Save all file information
	// to delay to last to delete. because if fail to delete posts, there is a chance to roll back.
	var fileids []string

	for key := range fileIdMap {
		fileids = append(fileids, key)
	}

	var fileInfos []*model.FileInfo
	sql = builder.Select("*").From("FileInfo").Where(sq.Eq{"Id": fileids})
	sqlstr, args, err = sql.ToSql()
	if err != nil {
		return errors.Wrapf(err, "failed to build fileid fetching sql string.")
	}

	_, err = ss.GetMaster().Select(&fileInfos, sqlstr, args...)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch file ids.")
	}

	mlog.Info(fmt.Sprintf("Prune: file id counts :%v", len(fileInfos)))

	//----------------------------------------
	//  Deleting process
	//----------------------------------------

	transaction, err := ss.GetMaster().Begin()

	if err != nil {
		return errors.Wrapf(err, "failed to start transaction.")
	}

	defer finalizeTransaction(transaction)

	//****************************************
	// Reaction post deletion
	//****************************************
	var reactionIds []string
	for id := range reactionMap {
		reactionIds = append(reactionIds, id)
	}
	query, args, err := builder.Delete("Reactions").Where(sq.Eq{"Id": reactionIds}).ToSql()

	if err != nil {
		return errors.Wrapf(err, "failed to build delete from Reaction query string.")
	}

	if _, err = transaction.Exec(query, args...); err != nil {
		return errors.Wrapf(err, "failed to execute delete from Reaction query")
	}

	mlog.Info("Prune: Reaction table was cleaned.")

	//****************************************
	// ThreadMembership deletion
	//****************************************
	query, args, err = builder.Delete("ThreadMemberships").Where(sq.Eq{"PostId": rootsid}).ToSql()

	if err != nil {
		return errors.Wrapf(err, "failed to build delete from ThreadMemberships query string.")
	}

	if _, err = transaction.Exec(query, args...); err != nil {
		return errors.Wrapf(err, "failed to execute delete from ThreadMemberships query")
	}

	mlog.Info("Prune: ThreadMemberships table was cleaned.")

	//****************************************
	// Threads deletion
	//****************************************
	query, args, err = builder.Delete("Threads").Where(sq.Eq{"PostId": rootsid}).ToSql()

	if err != nil {
		return errors.Wrapf(err, "failed to build delete from Threads query string.")
	}

	if _, err = transaction.Exec(query, args...); err != nil {
		return errors.Wrapf(err, "failed to execute delete from Threads query")
	}

	mlog.Info("Prune: Threads table was cleaned.")

	//****************************************
	// FileInfo deletion
	//****************************************
	query, args, err = builder.Delete("FileInfo").Where(sq.Eq{"Id": fileids}).ToSql()

	if err != nil {
		return errors.Wrapf(err, "failed to build delete from FileInfo query string.")
	}

	if _, err = transaction.Exec(query, args...); err != nil {
		return errors.Wrapf(err, "failed to execute delete from FileInfo query")
	}

	mlog.Info("Prune: FileInfo table was cleaned.")

	//****************************************
	// Preferences deletion
	//****************************************
	query, args, err = builder.Delete("Preferences").Where(sq.And{sq.Eq{"Name": delIds}, sq.Eq{"Category": "flagged_post"}}).ToSql()

	if err != nil {
		return errors.Wrapf(err, "failed to build delete from Preferences query string.")
	}

	if _, err = transaction.Exec(query, args...); err != nil {
		return errors.Wrapf(err, "failed to execute delete from Preferences query")
	}

	mlog.Info("Prune: Preferences table was cleaned.")

	//****************************************
	// Posts deletion
	//****************************************
	query, args, err = builder.Delete("Posts").Where(sq.Eq{"Id": delIds}).ToSql()

	if err != nil {
		return errors.Wrapf(err, "failed to build delete from Posts query string.")
	}

	if _, err = transaction.Exec(query, args...); err != nil {
		return errors.Wrapf(err, "failed to execute delete from Posts query")
	}

	mlog.Info(fmt.Sprintf("Prune: %v posts were deleted from table Posts.", len(delIds)))

	if err := transaction.Commit(); err != nil {
		return errors.Wrap(err, "commit_transaction")
	}

	//****************************************
	//  Start deleting files
	//****************************************

	// Start deleting files
	for _, fileInfo := range fileInfos {
		// the Dir of the path is the file id ,every file should have a individual id
		path := filepath.Dir(fileInfo.Path)
		if err := pr.app.RemoveDirectory(path); err != nil {
			mlog.Error(fmt.Sprintf("Prune: failed to delete file %s", path), mlog.Err(err))
		}

		for {

			path = filepath.Dir(path)

			if path == "." {
				break
			}

			if fs, err := pr.app.ListDirectory(path); err != nil {
				mlog.Error(fmt.Sprintf("Prune: failed to list directory %s", path), mlog.Err(err))
                                break
			} else {

				if len(fs) == 0 {
					if err := pr.app.RemoveDirectory(path); err != nil {
						mlog.Error(fmt.Sprintf("Prune: failed to delete file %s", path), mlog.Err(err))
					}
				} else {
					break
				}
			}

		}
	}

	mlog.Info("Prune: files was cleaned.")

	for chId := range chMap {
		pr.srv.Store.Channel().InvalidatePinnedPostCount(chId)
		pr.srv.Store.Post().InvalidateLastPostTimeCache(chId)
	}

	mlog.Info("Prune: invalidudate cache completed.")


	return nil

}

func getQueryBuilder(ss *sqlstore.SqlStore) sq.StatementBuilderType {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	if ss.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		builder = builder.PlaceholderFormat(sq.Dollar)
	}
	return builder
}

func finalizeTransaction(transaction *gorp.Transaction) {
	// Rollback returns sql.ErrTxDone if the transaction was already closed.
	if err := transaction.Rollback(); err != nil && err != sql.ErrTxDone {
		mlog.Error("Failed to rollback transaction", mlog.Err(err))
	}
}
