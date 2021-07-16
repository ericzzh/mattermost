package retention

import (
	// "encoding/json"
	"fmt"
	"os"
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

type Pruner struct {
	app      *app.App
	srv      *app.Server
	sqlstore *sqlstore.SqlStore
}

func New(a *app.App) *Pruner {

	s := a.Srv()
	// st, _ := json.Marshal(s.Config().SqlSettings)
	// mlog.Debug(string(st))

	ss := sqlstore.New(s.Config().SqlSettings, nil)

	return &Pruner{
		app:      a,
		srv:      s,
		sqlstore: ss,
	}

}

func (pr *Pruner) PruneGeneral() error {
	plies, err := pr.getTeamAsChannels()
	if err != nil {
		return err
	}
	ex := pr.fetchAllChannelIds(plies)

	if err := pr.pruneActions(nil, ex, policy.period); err != nil {
		return errors.Wrapf(err, "failed to execute PruneGeneral()")
	}

	return nil
}

func (pr *Pruner) merge{}

func (pr *Pruner) getUsersAsChannels(id string) (plies SimpleSpecificPolicy, err error) {
	return nil, nil
}
func (pr *Pruner) getTeamAsChannels(id string) (plies SimpleSpecificPolicy, err error) {

	plies = make(SimpleSpecificPolicy)

	chs, err := pr.srv.Store.Channel().GetTeamChannels(id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to call GetTeamChannels()")
	}
	for _, ch := range []*model.Channel(*chs) {
		plies[ch.Id] = policy.team[id]

	}

	return plies, nil
}

func (pr *Pruner) fetchAllChannelIds(plies SimpleSpecificPolicy) []string {

	return nil
}

func (pr *Pruner) pruneActions(ch []string, ex []string, period time.Duration) error {

	ss := pr.sqlstore

	now := time.Now()
	endTime := model.GetMillisForTime(now.Add(-time.Second * period))

	mlog.Debug(fmt.Sprintf("Prune: endTime: %v", endTime))

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

	// OriginalId is not a key, so fetching the sub-root using this OriginalId is not perfomant, we left them out
	sql = sql.Where("UpdateAt < ?", endTime).Where("RootId = ?", "").Where(
		"IsPinned <> ?", 1) //.Where("OriginalId = ?", "")

	if ex != nil {
		sql = sql.Where(sq.NotEq{"ChannelId": ex})
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

	mlog.Debug(fmt.Sprintf("Prune: cadidate root counts :%v", len(roots)))

	//----------------------------------------
	//   Thread posts fetching
	//----------------------------------------

	// Fetch all sub-Root(i.e. edit, pin to root) and Threads from root
	// using ROOTID and ORIGINIALID
	// We should also filter all the DeletedAt field to fetch the active posts
	// We can't use the Deleted record to justify.
	rootsid := make([]string, 0)
	for _, root := range roots {
		//  Thread is not link to sub-root
		if root.OriginalId == "" {
			rootsid = append(rootsid, root.Id)
		}
	}

	builder = getQueryBuilder(ss)
	sql = builder.Select("*").From("Posts").Where(sq.Or{sq.Eq{"RootId": rootsid}, sq.Eq{"OriginalId": rootsid}})
	sqlstr, args, err = sql.ToSql()
	if err != nil {
		return errors.Wrapf(err, "failed to build candidate threads fetching sql string.")
	}
	var threadsCand []*model.Post
	_, err = ss.GetMaster().Select(&threadsCand, sqlstr, args...)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch candidate threads.")
	}

	mlog.Debug(fmt.Sprintf("Prune: cadidate threads counts :%v", len(threadsCand)))

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

	mlog.Debug(fmt.Sprintf("Prune: found %d FileInfo Ids", len(fileIdMap)))
	mlog.Debug(fmt.Sprintf("Prune: found %d Reaction post Ids", len(reactionMap)))

	// Save all file information
	// to delay to last to delete. because if fail to delete posts, there is a chance to roll back.
	var fileids []string

	for key := range fileIdMap {
		fileids = append(fileids, key)
	}

	var fileInfos []*model.FileInfo
	sql = sq.Select("*").From("FileInfo").Where(sq.Eq{"Id": fileids})
	sqlstr, args, err = sql.ToSql()
	if err != nil {
		return errors.Wrapf(err, "failed to build fileid fetching sql string.")
	}

	_, err = ss.GetMaster().Select(&fileInfos, sqlstr, args...)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch file ids.")
	}

	mlog.Debug(fmt.Sprintf("Prune: file id counts :%v", len(fileInfos)))

	//----------------------------------------
	//  Delete process
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
	query, args, err := sq.Delete("Reactions").Where(sq.Eq{"Id": reactionIds}).ToSql()

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
	query, args, err = sq.Delete("ThreadMemberShips").Where(sq.Eq{"Id": rootsid}).ToSql()

	if err != nil {
		return errors.Wrapf(err, "failed to build delete from ThreadMemberShips query string.")
	}

	if _, err = transaction.Exec(query, args...); err != nil {
		return errors.Wrapf(err, "failed to execute delete from ThreadMemberShips query")
	}

	mlog.Info("Prune: ThreadMemberShips table was cleaned.")

	//****************************************
	// Threads deletion
	//****************************************
	query, args, err = sq.Delete("Threads").Where(sq.Eq{"Id": rootsid}).ToSql()

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
	query, args, err = sq.Delete("FileInfo").Where(sq.Eq{"Id": fileids}).ToSql()

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
	query, args, err = sq.Delete("Preferences").Where(sq.And{sq.Eq{"Name": delIds}, sq.Eq{"Category": "flagged_post"}}).ToSql()

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
	query, args, err = sq.Delete("Posts").Where(sq.Eq{"Id": delIds}).ToSql()

	if err != nil {
		return errors.Wrapf(err, "failed to build delete from Posts query string.")
	}

	if _, err = transaction.Exec(query, args...); err != nil {
		return errors.Wrapf(err, "failed to execute delete from Posts query")
	}

	mlog.Info("Prune: Posts table was cleaned.")

	if err = transaction.Commit(); err != nil {
		return errors.Wrap(err, "commit_transaction")
	}

	//****************************************
	//  Start deleting files
	//****************************************

	// Start deleting files
	for _, fileInfo := range fileInfos {
		// the Dir of the path is the file id ,every file should have a individual id
		path := filepath.Dir(fileInfo.Path)
		if err = os.RemoveAll(path); err != nil {
			mlog.Error(fmt.Sprintf("Prune: failed to delete file %s", path), mlog.Err(err))
		}
	}

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
