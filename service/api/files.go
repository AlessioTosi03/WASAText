package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) serveFile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	photoPath := ps.ByName("photo")
	http.ServeFile(w, r, "/tmp/"+photoPath)
}
