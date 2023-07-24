package {{.Vars.AppName}}

import (
	"net/http"
)

type API struct {}

func (self *API) ServeHTTP(w http.ReponseWriter, r *http.Request)  {
	
}