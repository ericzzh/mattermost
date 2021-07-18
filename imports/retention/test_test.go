package retention

import (
	"github.com/mattermost/mattermost-server/v5/api4"
	"testing"
)

func __TestMyTest(t *testing.T) {

	th := api4.Setup(t)
	defer th.TearDown()
        th.Server.Config().SqlSettings.DataSource = mainHelper.Settings.DataSource
	// Prune(th.App)
}
