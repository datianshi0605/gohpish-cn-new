package api

import (
	"encoding/json"
	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (as *Server) CampaignLongTerm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&req); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "请求格式错误"}, 400)
		return
	}
	if err := models.SetCampaignLongTerm(id, ctx.Get(r, "user_id").(int64), req.Enabled); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, 400)
		return
	}
	JSONResponse(w, models.Response{Success: true, Message: "长期演练设置已更新"}, 200)
}
func (as *Server) CampaignAppendRecipients(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	var req models.AppendRecipientsRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20)).Decode(&req); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "请求格式错误"}, 400)
		return
	}
	result, err := models.AppendCampaignRecipients(id, ctx.Get(r, "user_id").(int64), req)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, 400)
		return
	}
	JSONResponse(w, result, http.StatusCreated)
}

func (as *Server) CampaignAutoGroups(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	uid := ctx.Get(r, "user_id").(int64)
	if r.Method == http.MethodGet {
		groups, err := models.GetCampaignAutoGroups(id, uid)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "活动不存在或无权访问"}, http.StatusNotFound)
			return
		}
		JSONResponse(w, groups, http.StatusOK)
		return
	}
	var req struct {
		GroupIDs []int64 `json:"group_ids"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(&req); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "请求格式错误"}, 400)
		return
	}
	if err := models.SetCampaignAutoGroups(id, uid, req.GroupIDs); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, 400)
		return
	}
	JSONResponse(w, models.Response{Success: true, Message: "关联用户组已保存，新人员将自动加入演练"}, 200)
}
