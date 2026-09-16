package models

import (
	check "gopkg.in/check.v1"
	"time"
)

func (s *ModelsSuite) TestAutomaticMultipleGroups(c *check.C) {
	campaign := s.createCampaignDependencies(c)
	campaign.LongTerm = true
	c.Assert(PostCampaign(&campaign, 1), check.IsNil)
	tracked, err := GetCampaignAutoGroups(campaign.Id, 1)
	c.Assert(err, check.IsNil)
	c.Assert(len(tracked), check.Equals, 1)
	first, err := GetGroup(tracked[0].Id, 1)
	c.Assert(err, check.IsNil)
	second := Group{Name: "另一批同事", UserId: 1, Targets: []Target{{BaseRecipient: BaseRecipient{Email: "second@example.com"}}, first.Targets[0]}}
	c.Assert(PostGroup(&second), check.IsNil)
	c.Assert(SetCampaignAutoGroups(campaign.Id, 1, []int64{first.Id, second.Id, second.Id}), check.IsNil)
	c.Assert(SyncLongTermCampaigns(), check.IsNil)
	updated, err := GetCampaign(campaign.Id, 1)
	c.Assert(err, check.IsNil)
	c.Assert(len(updated.Results), check.Equals, 5)
	// Later additions to either group, plus overlapping membership, only enroll once.
	first.Name = "重命名后的入职组"
	first.Targets = append(first.Targets, Target{BaseRecipient: BaseRecipient{Email: "overlap@example.com"}})
	c.Assert(PutGroup(&first), check.IsNil)
	second.Targets = append(second.Targets, Target{BaseRecipient: BaseRecipient{Email: "overlap@example.com"}}, Target{BaseRecipient: BaseRecipient{Email: "third@example.com"}})
	c.Assert(PutGroup(&second), check.IsNil)
	c.Assert(SyncLongTermCampaigns(), check.IsNil)
	updated, err = GetCampaign(campaign.Id, 1)
	c.Assert(err, check.IsNil)
	c.Assert(len(updated.Results), check.Equals, 7)
	c.Assert(SyncLongTermCampaigns(), check.IsNil)
	logs, err := GetMailLogsByCampaign(campaign.Id)
	c.Assert(err, check.IsNil)
	c.Assert(len(logs), check.Equals, 7)
	tracked, err = GetCampaignAutoGroups(campaign.Id, 1)
	c.Assert(err, check.IsNil)
	c.Assert(tracked[0].Name, check.Equals, first.Name)
	// Unsubscribing does not remove results and prevents following later changes.
	c.Assert(SetCampaignAutoGroups(campaign.Id, 1, []int64{second.Id}), check.IsNil)
	first.Targets = append(first.Targets, Target{BaseRecipient: BaseRecipient{Email: "untracked@example.com"}})
	c.Assert(PutGroup(&first), check.IsNil)
	c.Assert(SyncLongTermCampaigns(), check.IsNil)
	updated, _ = GetCampaign(campaign.Id, 1)
	c.Assert(len(updated.Results), check.Equals, 7)
	// Deleting a followed group leaves history intact and doesn't block future sync.
	c.Assert(DeleteGroup(&second), check.IsNil)
	c.Assert(SyncLongTermCampaigns(), check.IsNil)
	updated, _ = GetCampaign(campaign.Id, 1)
	c.Assert(len(updated.Results), check.Equals, 7)
	c.Assert(DeleteCampaign(campaign.Id), check.IsNil)
	var links []CampaignAutoGroup
	c.Assert(db.Where("campaign_id=?", campaign.Id).Find(&links).Error, check.IsNil)
	c.Assert(len(links), check.Equals, 0)
}

func (s *ModelsSuite) TestAutomaticGroupsOwnershipAndDisable(c *check.C) {
	campaign := s.createCampaignDependencies(c)
	campaign.LongTerm = true
	campaign.LaunchDate = time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	c.Assert(PostCampaign(&campaign, 1), check.IsNil)
	first := campaign.Groups[0]
	foreign := Group{Name: "Other user", UserId: 2, Targets: []Target{{BaseRecipient: BaseRecipient{Email: "foreign@example.com"}}}}
	c.Assert(PostGroup(&foreign), check.IsNil)
	c.Assert(SetCampaignAutoGroups(campaign.Id, 1, []int64{foreign.Id}), check.NotNil)
	c.Assert(SetCampaignAutoGroups(campaign.Id, 2, []int64{foreign.Id}), check.NotNil)
	linked, err := GetCampaignAutoGroups(campaign.Id, 1)
	c.Assert(err, check.IsNil)
	c.Assert(len(linked), check.Equals, 1) // Invalid changes cannot clear existing links.
	first.Targets = append(first.Targets, Target{BaseRecipient: BaseRecipient{Email: "later@example.com"}})
	c.Assert(PutGroup(&first), check.IsNil)
	c.Assert(SetCampaignLongTerm(campaign.Id, 1, false), check.IsNil)
	c.Assert(SyncLongTermCampaigns(), check.IsNil)
	current, _ := GetCampaign(campaign.Id, 1)
	c.Assert(len(current.Results), check.Equals, 4)
	c.Assert(SetCampaignLongTerm(campaign.Id, 1, true), check.IsNil)
	c.Assert(SyncLongTermCampaigns(), check.IsNil)
	current, _ = GetCampaign(campaign.Id, 1)
	c.Assert(len(current.Results), check.Equals, 5)
	for _, r := range current.Results {
		c.Assert(r.SendDate.Before(campaign.LaunchDate), check.Equals, false)
	}
	c.Assert(CompleteCampaign(campaign.Id, 1), check.IsNil)
	first.Targets = append(first.Targets, Target{BaseRecipient: BaseRecipient{Email: "after-complete@example.com"}})
	c.Assert(PutGroup(&first), check.IsNil)
	c.Assert(SyncLongTermCampaigns(), check.IsNil)
	current, _ = GetCampaign(campaign.Id, 1)
	c.Assert(len(current.Results), check.Equals, 5)
}
