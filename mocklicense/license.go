package mocklicense

import (
	"github.com/mattermost/mattermost-server/v5/app"
	"github.com/mattermost/mattermost-server/v5/einterfaces"
	"github.com/mattermost/mattermost-server/v5/model"
)

type Mocklicense struct {
	srv *app.Server
}

func init() {
	app.RegisterLicenseInterface(NewMocklicense)
}

func NewMocklicense(s *app.Server) einterfaces.LicenseInterface {
	return &Mocklicense{
		srv: s,
	}
}

func (m *Mocklicense) CanStartTrial() (bool, error) {
	return false, nil
}

func (m *Mocklicense) GetPrevTrial()  (*model.License, error){
       return nil, nil 
}
