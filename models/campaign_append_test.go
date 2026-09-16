package models

import (
	check "gopkg.in/check.v1"
	"strings"
	"time"
)

func (s *ModelsSuite) TestLongTermAppendPreservesCampaign(c *check.C) {
	campaign := s.createCampaign(c)
	original := campaign.Results[0]
	req := AppendRecipientsRequest{Targets: []BaseRecipient{{Email: "new@example.com"}, {Email: strings.ToUpper(original.Email)}, {Email: " NEW@example.com "}}}
	_, err := AppendCampaignRecipients(campaign.Id, campaign.UserId, req)
	c.Assert(err, check.NotNil)
	c.Assert(SetCampaignLongTerm(campaign.Id, campaign.UserId, true), check.IsNil)
	_, err = AppendCampaignRecipients(campaign.Id, 999, req)
	c.Assert(err, check.NotNil)
	out, err := AppendCampaignRecipients(campaign.Id, campaign.UserId, req)
	c.Assert(err, check.IsNil)
	c.Assert(out.Added, check.Equals, 1)
	c.Assert(out.Skipped, check.Equals, 2)
	updated, err := GetCampaign(campaign.Id, campaign.UserId)
	c.Assert(err, check.IsNil)
	c.Assert(updated.TemplateId, check.Equals, campaign.TemplateId)
	c.Assert(updated.PageId, check.Equals, campaign.PageId)
	c.Assert(updated.SMTPId, check.Equals, campaign.SMTPId)
	c.Assert(updated.URL, check.Equals, campaign.URL)
	c.Assert(len(updated.Results), check.Equals, len(campaign.Results)+1)
	same, err := GetResult(original.RId)
	c.Assert(err, check.IsNil)
	c.Assert(same.Status, check.Equals, original.Status)
	c.Assert(same.Email, check.Equals, original.Email)
	logs, err := GetMailLogsByCampaign(campaign.Id)
	c.Assert(err, check.IsNil)
	c.Assert(len(logs), check.Equals, len(campaign.Results)+1)
	for _, m := range logs {
		if m.RId == updated.Results[len(updated.Results)-1].RId {
			c.Assert(m.Processing, check.Equals, false)
		}
	}
	out, err = AppendCampaignRecipients(campaign.Id, campaign.UserId, req)
	c.Assert(err, check.IsNil)
	c.Assert(out.Added, check.Equals, 0)
	c.Assert(CompleteCampaign(campaign.Id, campaign.UserId), check.IsNil)
	_, err = AppendCampaignRecipients(campaign.Id, campaign.UserId, req)
	c.Assert(err, check.NotNil)
	c.Assert(SetCampaignLongTerm(campaign.Id, campaign.UserId, true), check.NotNil)
	logs, err = GetMailLogsByCampaign(campaign.Id)
	c.Assert(err, check.IsNil)
	c.Assert(len(logs), check.Equals, 0)
}

func (s *ModelsSuite) TestLongTermGroupScheduleAndValidation(c *check.C) {
	campaign := s.createCampaign(c)
	c.Assert(SetCampaignLongTerm(campaign.Id, campaign.UserId, true), check.IsNil)
	_, err := AppendCampaignRecipients(campaign.Id, campaign.UserId, AppendRecipientsRequest{Targets: []BaseRecipient{{Email: "ok@example.com"}, {Email: "invalid"}}})
	c.Assert(err, check.NotNil)
	current, _ := GetCampaign(campaign.Id, campaign.UserId)
	c.Assert(len(current.Results), check.Equals, len(campaign.Results))
	g := Group{Name: "New staff", UserId: campaign.UserId, Targets: []Target{{BaseRecipient: BaseRecipient{Email: "later@example.com"}}}}
	c.Assert(PostGroup(&g), check.IsNil)
	when := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	out, err := AppendCampaignRecipients(campaign.Id, campaign.UserId, AppendRecipientsRequest{Groups: []Group{{Name: g.Name}}, SendDate: when})
	c.Assert(err, check.IsNil)
	c.Assert(out.Added, check.Equals, 1)
	current, _ = GetCampaign(campaign.Id, campaign.UserId)
	for _, r := range current.Results {
		if r.Email == "later@example.com" {
			c.Assert(r.SendDate.Equal(when), check.Equals, true)
			c.Assert(r.Status, check.Equals, StatusScheduled)
		}
	}
	c.Assert(SetCampaignLongTerm(campaign.Id, campaign.UserId, false), check.IsNil)
	_, err = AppendCampaignRecipients(campaign.Id, campaign.UserId, AppendRecipientsRequest{Groups: []Group{{Name: g.Name}}})
	c.Assert(err, check.NotNil)
}
