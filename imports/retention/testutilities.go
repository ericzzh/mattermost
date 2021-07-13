package retention

import (
	// "fmt"
	"os"
	// "regexp"
	"sync"
	"testing"

	// "time"

	// "github.com/go-sql-driver/mysql"
	// "github.com/lib/pq"
	// "github.com/mattermost/gorp"
	// "github.com/pkg/errors"
	// "github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/require"

	// "github.com/mattermost/mattermost-server/v5/einterfaces/mocks"
	"github.com/mattermost/mattermost-server/v5/api4"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
	"github.com/mattermost/mattermost-server/v5/store/sqlstore"

	// "github.com/mattermost/mattermost-server/v5/store/searchtest"
	"github.com/mattermost/mattermost-server/v5/store/storetest"
	"github.com/mattermost/mattermost-server/v5/utils"
)

type storeType struct {
	Name        string
	SqlSettings *model.SqlSettings
	SqlStore    *sqlstore.SqlStore
	Store       store.Store
}

var storeTypes []*storeType

func newStoreType(name, driver string) *storeType {
	return &storeType{
		Name:        name,
		SqlSettings: storetest.MakeSqlSettings(driver, false),
	}
}
func initStores() {
	if testing.Short() {
		return
	}
	// In CI, we already run the entire test suite for both mysql and postgres in parallel.
	// So we just run the tests for the current database set.
	if os.Getenv("IS_CI") == "true" {
		switch os.Getenv("MM_SQLSETTINGS_DRIVERNAME") {
		case "mysql":
			storeTypes = append(storeTypes, newStoreType("MySQL", model.DATABASE_DRIVER_MYSQL))
		case "postgres":
			storeTypes = append(storeTypes, newStoreType("PostgreSQL", model.DATABASE_DRIVER_POSTGRES))
		}
	} else {
		storeTypes = append(storeTypes, newStoreType("MySQL", model.DATABASE_DRIVER_MYSQL),
			newStoreType("PostgreSQL", model.DATABASE_DRIVER_POSTGRES))
	}

	defer func() {
		if err := recover(); err != nil {
			tearDownStores()
			panic(err)
		}
	}()
	var wg sync.WaitGroup
	for _, st := range storeTypes {
		st := st
		wg.Add(1)
		go func() {
			defer wg.Done()
			st.SqlStore = sqlstore.New(*st.SqlSettings, nil)
			st.Store = st.SqlStore
			st.Store.DropAllTables()
			st.Store.MarkSystemRanUnitTests()
		}()
	}
	wg.Wait()
}

var tearDownStoresOnce sync.Once

func tearDownStores() {
	if testing.Short() {
		return
	}
	tearDownStoresOnce.Do(func() {
		var wg sync.WaitGroup
		wg.Add(len(storeTypes))
		for _, st := range storeTypes {
			st := st
			go func() {
				if st.Store != nil {
					st.Store.Close()
				}
				if st.SqlSettings != nil {
					storetest.CleanupSqlSettings(st.SqlSettings)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	})
}

func StoreTestWithSqlStore(t *testing.T, f func(*testing.T, store.Store, storetest.SqlStore)) {
	defer func() {
		if err := recover(); err != nil {
			tearDownStores()
			panic(err)
		}
	}()
	for _, st := range storeTypes {
		st := st
		t.Run(st.Name, func(t *testing.T) {
			if testing.Short() {
				t.SkipNow()
			}
			f(t, st.Store, st.SqlStore)
		})
	}
}

func NewTestId() string {
	newId := []byte(model.NewId())

	for i := 1; i < len(newId); i = i + 2 {
		newId[i] = 48 + newId[i-1]%10
	}

	return string(newId)
}

func MakeEmail() string {
	return "success_" + model.NewId() + "@simulator.amazonses.com"
}

func  LoginWithClient(user *model.User, client *model.Client4) {
	utils.DisableDebugLogForTest()
	_, resp := client.Login(user.Email, user.Password)
	if resp.Error != nil {
		panic(resp.Error)
	}
	utils.EnableDebugLogForTest()
}

func  CreateDmChannel(th *api4.TestHelper, user *model.User, other *model.User) *model.Channel {
	utils.DisableDebugLogForTest()
	var err *model.AppError
	var channel *model.Channel
	if channel, err = th.App.GetOrCreateDirectChannel(user.Id, other.Id); err != nil {
		panic(err)
	}
	utils.EnableDebugLogForTest()
	return channel
}
