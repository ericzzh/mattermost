package retention

import (
	"testing"
        // "fmt"

	"github.com/mattermost/mattermost-server/v5/shared/mlog"
	// "github.com/mattermost/mattermost-server/v5/store/sqlstore"
	"github.com/mattermost/mattermost-server/v5/testlib"
	// "github.com/mattermost/mattermost-server/v5/imports/retention"
)

var mainHelper *testlib.MainHelper

func TestMain(m *testing.M) {
	mlog.DisableZap()
	mainHelper = testlib.NewMainHelperWithOptions(nil)
	defer mainHelper.Close()

        initStores()
        // fmt.Println("initStores OK.")

	mainHelper.Main(m)
        tearDownStores()
}
