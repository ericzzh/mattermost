package retention

import (
	"encoding/json"
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

type Stats struct {
	OrgRoots        int
	NonOrgRoots     int
	DeletedOrgRoots int
	Pinned          int
	PruneOrgRoots   int
	System          int
	Threads         int
	Db_reactions    int
	Db_threadMem    int
	Db_threads      int
	Db_fileInfo     int
	Db_preference   int
	Db_posts        int
}
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

		if _, err := pr.pruneActions([]string{chid}, nil, p); err != nil {
			return errors.Wrapf(err, "failed to call pruneActions().")
		}
		mlog.Debug(fmt.Sprintf("Prune: specific case, channel: %s, completed.", chid))

	}
	return nil

}
func (pr *Prune) pruneGeneral() error {
	ex := pr.fetchAllChannelIds(pr.merged)

	if _, err := pr.pruneActions(nil, ex, policy.period); err != nil {
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
func (pr *Prune) pruneActions(ch []string, ex []string, period time.Duration) (*Stats, error) {

	var st Stats

	_ = json.MarshalIndent // for debug
	ss := pr.sqlstore

	now := time.Now()
	endTime := model.GetMillisForTime(now.Add(-time.Second * period))

	mlog.Info(fmt.Sprintf("************************* Prune start  End time is: %v *************************", endTime))

	//----------------------------------------
	//   Root post fetching
	//----------------------------------------
	var roots []*model.Post

	// Fetch all the root messages
	// we use UpdateAt as a key, because all thread updated (add, pin) will update this field.

	builder := getQueryBuilder(ss)
	sql := builder.Select("*").From("Posts")

	if ch != nil {
		sql = sql.Where(sq.Eq{"ChannelId": ch})
	}

	//Select all the true roots candidate, we will filter them out afterwards
	//Because originalid is not a key, we don't select them separately, and select them here together
	sql = sql.Where("UpdateAt < ? And RootId = ''", endTime)

	//except the specific channels
	if ex != nil {
		sql = sql.Where(sq.NotEq{"ChannelId": ex})
	}

	sqlstr, args, err := sql.ToSql()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build true roots candidate sql string.")
	}

	_, err = ss.GetMaster().Select(&roots, sqlstr, args...)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to select true roots candidate.")
	}

	mlog.Info(fmt.Sprintf("Prune: %v root posts were selected from Posts.", len(roots)))

	// tmp, _ := json.MarshalIndent(roots, "", "\t")
	// fmt.Printf("*** Debug root sql result %v", string(tmp))

	//A tree to express the structure of a post family
	type rootTree struct {
		post        *model.Post
		subroots    map[string]*model.Post
		threads     map[string]*model.Post
		childPinned bool
	}

	//A dictionary struct for classifiy the roots
	rootDict := struct {
		trueroots map[string]*rootTree
		deleted   map[string]*rootTree
		pinned    map[string]*rootTree
		prune     map[string]*rootTree
		system    map[string]*rootTree
	}{
		map[string]*rootTree{},
		map[string]*rootTree{},
		map[string]*rootTree{},
		map[string]*rootTree{},
		map[string]*rootTree{},
	}

	// true root
	for _, root := range roots {

		// only consider true root here
		if root.OriginalId != "" {
			continue
		}

		// VERY IMPORTANT to be aware, we map all the root to a single node
		rt := &rootTree{
			post:     root,
			subroots: map[string]*model.Post{},
			threads:  map[string]*model.Post{},
		}

		rootDict.trueroots[root.Id] = rt

		if root.Type != "" {
			rootDict.system[root.Id] = rt
		}

		// true root deleted. not originial
		if root.DeleteAt != 0 {

			rootDict.deleted[root.Id] = rt
			// if the true root is marked as deleted, it should be a prune root
                        // no matter permanent or not
			rootDict.prune[root.Id] = rt

		} else if root.IsPinned {

			// Pinned root is not pruned

			rootDict.pinned[root.Id] = rt

		} else {

			// period == 0 means permanent policy, we only prune when period != 0
			if period != 0 {
				// not deleted and not pinned true root should be pruned
				rootDict.prune[root.Id] = rt
			}
		}
	}

	// sub root
	var cntSubroots int
	for _, root := range roots {
		// only consider sub root here
		if root.OriginalId == "" {
			continue
		}
		// this also modify the same node in the rootDict.prune maps.
		rootDict.trueroots[root.OriginalId].subroots[root.Id] = root
		cntSubroots++
	}

	//----------------------------------------
	//   Thread post fetching
	//----------------------------------------
	trueRootsId := make([]string, 0)
	for _, root := range rootDict.prune {

		trueRootsId = append(trueRootsId, root.post.Id)

	}

	// we don't consider much thread information ,such as deleted, originalid
	// because if the root is to be pruned, all the thread should be pruned
	sql = builder.Select("*").From("Posts").Where(sq.Eq{"RootId": trueRootsId})
	sqlstr, args, err = sql.ToSql()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build candidate threads fetching sql string.")
	}

	var threadsCand []*model.Post
	_, err = ss.GetMaster().Select(&threadsCand, sqlstr, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch candidate threads.")
	}

	mlog.Info(fmt.Sprintf("Prune: %v threads were selected from Posts.", len(threadsCand)))

	//check and filter the threads
	for _, thread := range threadsCand {

		root := rootDict.trueroots[thread.RootId]

		// save the thread to its root
		root.threads[thread.Id] = thread

		// if thread is pinned, the whole root will be preserverd
		if thread.DeleteAt == 0 && thread.IsPinned {
			root.childPinned = true
			rootDict.pinned[root.post.Id] = root
			delete(rootDict.prune, root.post.Id)
		}
	}

	st.OrgRoots = len(rootDict.trueroots)
	mlog.Info(fmt.Sprintf("Prune: %v true root posts.", st.OrgRoots))

	st.DeletedOrgRoots = len(rootDict.deleted)
	mlog.Info(fmt.Sprintf("Prune: %v true deleted root posts.", st.DeletedOrgRoots))

	st.NonOrgRoots = cntSubroots
	mlog.Info(fmt.Sprintf("Prune: %v non-originial root posts.", st.NonOrgRoots))

	st.System = len(rootDict.system)
	mlog.Info(fmt.Sprintf("Prune: %v system root posts.", st.System))

	st.Pinned = len(rootDict.pinned)
	mlog.Info(fmt.Sprintf("Prune: %v pinned root(including pinned threads) posts.", st.Pinned))

	st.PruneOrgRoots = len(rootDict.prune)
	mlog.Info(fmt.Sprintf("Prune: %v prune roots", st.PruneOrgRoots))

	var cntThread int
	for _, r := range rootDict.prune {
		cntThread = cntThread + len(r.threads)
	}

	st.Threads = cntThread
	mlog.Info(fmt.Sprintf("Prune: %v prune threads", st.Threads))

	// we need to compute trueRootsId again, because the data meybe modified
	chMap := map[string]bool{}
	trueRootsId = make([]string, 0)
	for _, root := range rootDict.prune {

		trueRootsId = append(trueRootsId, root.post.Id)
		chMap[root.post.ChannelId] = true

	}
	//----------------------------------------
	//   Get all post id and other key information
	//----------------------------------------
	delIds := []string{}
	fileIdMap := map[string]bool{}
	reactionMap := map[string]bool{}

	for _, root := range rootDict.prune {
		delIds = append(delIds, root.post.Id)
		// we don't mind if the file is attached to a deleted post or not.
		// because all the root tree will be deleted
		if len(root.post.FileIds) != 0 {
			for _, fileid := range []string(root.post.FileIds) {
				fileIdMap[fileid] = true
			}
		}

		if root.post.HasReactions {
			reactionMap[root.post.Id] = true
		}

		for _, sub := range root.subroots {

			delIds = append(delIds, sub.Id)
			// we don't mind if the file is attached to a deleted post or not.
			// because all the root tree will be deleted
			if len(sub.FileIds) != 0 {
				for _, fileid := range []string(sub.FileIds) {
					fileIdMap[fileid] = true
				}
			}
		}

		for _, thread := range root.threads {

			delIds = append(delIds, thread.Id)
			// we don't mind if the file is attached to a deleted post or not.
			// because all the root tree will be deleted
			if len(thread.FileIds) != 0 {
				for _, fileid := range thread.FileIds {
					fileIdMap[fileid] = true
				}
			}

			if thread.HasReactions {
				reactionMap[thread.Id] = true
			}
		}

	}

	mlog.Info(fmt.Sprintf("Prune: %d Reaction post Ids", len(reactionMap)))

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
		return nil, errors.Wrapf(err, "failed to build fileid fetching sql string.")
	}

	_, err = ss.GetMaster().Select(&fileInfos, sqlstr, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch file ids.")
	}

	mlog.Info(fmt.Sprintf("Prune: %d FileInfo records.", len(fileInfos)))

	//----------------------------------------
	//  Deleting process
	//----------------------------------------

	transaction, err := ss.GetMaster().Begin()

	if err != nil {
		return nil, errors.Wrapf(err, "failed to start transaction.")
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
		return nil, errors.Wrapf(err, "failed to build delete from Reaction query string.")
	}

	sqlres, err := transaction.Exec(query, args...)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute delete from Reaction query")
	}

	rc, _ := sqlres.RowsAffected()
	st.Db_reactions = int(rc)
	mlog.Info(fmt.Sprintf("Prune: Reaction table was pruned. Effected rows: %v", rc))

	//****************************************
	// ThreadMembership deletion
	//****************************************
	query, args, err = builder.Delete("ThreadMemberships").Where(sq.Eq{"PostId": trueRootsId}).ToSql()

	if err != nil {
		return nil, errors.Wrapf(err, "failed to build delete from ThreadMemberships query string.")
	}

	sqlres, err = transaction.Exec(query, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute delete from ThreadMemberships query")
	}

	rc, _ = sqlres.RowsAffected()
	st.Db_threadMem = int(rc)
	mlog.Info(fmt.Sprintf("Prune: ThreadMemberships table was pruned. Effected rows: %v", rc))

	//****************************************
	// Threads deletion
	//****************************************
	query, args, err = builder.Delete("Threads").Where(sq.Eq{"PostId": trueRootsId}).ToSql()

	if err != nil {
		return nil, errors.Wrapf(err, "failed to build delete from Threads query string.")
	}
	sqlres, err = transaction.Exec(query, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute delete from Threads query")
	}

	rc, _ = sqlres.RowsAffected()
	st.Db_threads = int(rc)
	mlog.Info(fmt.Sprintf("Prune: Threads table was pruned. Effected rows: %v", rc))

	//****************************************
	// FileInfo deletion
	//****************************************
	query, args, err = builder.Delete("FileInfo").Where(sq.Eq{"Id": fileids}).ToSql()

	if err != nil {
		return nil, errors.Wrapf(err, "failed to build delete from FileInfo query string.")
	}
	sqlres, err = transaction.Exec(query, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute delete from FileInfo query")
	}

	rc, _ = sqlres.RowsAffected()
	st.Db_fileInfo = int(rc)
	mlog.Info(fmt.Sprintf("Prune: FileInfo table was pruned. Effected rows: %v", rc))

	//****************************************
	// Preferences deletion
	//****************************************
	query, args, err = builder.Delete("Preferences").Where(sq.And{sq.Eq{"Name": delIds}, sq.Eq{"Category": "flagged_post"}}).ToSql()

	if err != nil {
		return nil, errors.Wrapf(err, "failed to build delete from Preferences query string.")
	}
	sqlres, err = transaction.Exec(query, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute delete from Preferences query")
	}

	rc, _ = sqlres.RowsAffected()
	st.Db_preference = int(rc)
	mlog.Info(fmt.Sprintf("Prune: Preferences table was pruned. Effected rows: %v", rc))

	//****************************************
	// Posts deletion
	//****************************************
	query, args, err = builder.Delete("Posts").Where(sq.Eq{"Id": delIds}).ToSql()

	if err != nil {
		return nil, errors.Wrapf(err, "failed to build delete from Posts query string.")
	}
	sqlres, err = transaction.Exec(query, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute delete from Posts query")
	}

	rc, _ = sqlres.RowsAffected()
	st.Db_posts = int(rc)
	mlog.Info(fmt.Sprintf("Prune: Posts table were pruned. Effected rows: %v", rc))

	if err := transaction.Commit(); err != nil {
		return nil, errors.Wrap(err, "commit_transaction")
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

	mlog.Info(fmt.Sprintf("Prune: %v files were pruned.", len(fileInfos)))

	for chId := range chMap {
		pr.srv.Store.Channel().InvalidatePinnedPostCount(chId)
		pr.srv.Store.Post().InvalidateLastPostTimeCache(chId)
	}

	mlog.Info("Prune: invalidudate cache completed.")

	mlog.Info(fmt.Sprintf("************************* Prune end    End time is: %v *************************", endTime))
	return &st, nil

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
