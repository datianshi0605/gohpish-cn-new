package api

import (
	"fmt"
	"github.com/gophish/gophish/models"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLongTermCampaignAPI(t *testing.T) {
	ctx := setupTest(t)
	createTestData(t)
	campaigns, err := models.GetCampaigns(1)
	if err != nil || len(campaigns) != 1 {
		t.Fatal("missing fixture", err)
	}
	base := fmt.Sprintf("/api/campaigns/%d", campaigns[0].Id)
	call := func(path, body, key string) int {
		req := httptest.NewRequest("POST", base+path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if key != "" {
			req.Header.Set("Authorization", "Bearer "+key)
		}
		res := httptest.NewRecorder()
		ctx.apiServer.ServeHTTP(res, req)
		return res.Code
	}
	if code := call("/long-term", `{"enabled":true}`, ""); code < 400 {
		t.Fatal("unauthenticated change allowed")
	}
	if code := call("/long-term", `{"enabled":true}`, ctx.apiKey); code != 200 {
		t.Fatal("enable failed", code)
	}
	groups, err := models.GetGroups(1)
	if err != nil || len(groups) != 1 {
		t.Fatal("missing group", err)
	}
	payload := fmt.Sprintf(`{"group_ids":[%d]}`, groups[0].Id)
	if code := call("/auto-groups", payload, ""); code < 400 {
		t.Fatal("unauthenticated subscription allowed")
	}
	if code := call("/auto-groups", payload, ctx.apiKey); code != 200 {
		t.Fatal("subscription failed", code)
	}
	if code := call("/auto-groups", `{"group_ids":[999999]}`, ctx.apiKey); code != 400 {
		t.Fatal("missing group accepted", code)
	}
	linked, err := models.GetCampaignAutoGroups(campaigns[0].Id, 1)
	if err != nil || len(linked) != 1 {
		t.Fatal("subscription missing", err)
	}
	if code := call("/recipients", `{"targets":[{"email":"new@example.com"}]}`, ctx.apiKey); code != 201 {
		t.Fatal("append failed", code)
	}
	if code := call("/recipients", `{"targets":[{"email":"invalid"}]}`, ctx.apiKey); code != 400 {
		t.Fatal("invalid email allowed", code)
	}
	if code := call("/recipients", `{`, ctx.apiKey); code != 400 {
		t.Fatal("invalid JSON allowed", code)
	}
	if err := models.CompleteCampaign(campaigns[0].Id, 1); err != nil {
		t.Fatal(err)
	}
	if code := call("/recipients", `{"targets":[{"email":"other@example.com"}]}`, ctx.apiKey); code != 400 {
		t.Fatal("completed campaign accepted append", code)
	}
}
