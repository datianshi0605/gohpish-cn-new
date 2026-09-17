package models

import (
	"fmt"
	check "gopkg.in/check.v1"
	"sync"
	"time"
)

func (s *ModelsSuite) TestConcurrentThousandRecipientAppendAndClaims(c *check.C) {
	campaign := s.createCampaign(c)
	c.Assert(SetCampaignLongTerm(campaign.Id, 1, true), check.IsNil)
	targets := make([]BaseRecipient, 1200)
	for i := range targets {
		targets[i] = BaseRecipient{Email: fmt.Sprintf("load-%04d@example.com", i)}
	}
	start := time.Now()
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	added := make(chan int, 8)
	gate := make(chan struct{})
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-gate
			out, err := AppendCampaignRecipients(campaign.Id, 1, AppendRecipientsRequest{Targets: targets})
			errs <- err
			added <- out.Added
		}()
	}
	close(gate)
	wg.Wait()
	close(errs)
	close(added)
	for err := range errs {
		c.Assert(err, check.IsNil)
	}
	total := 0
	for n := range added {
		total += n
	}
	c.Assert(total, check.Equals, 1200)
	current, err := GetCampaign(campaign.Id, 1)
	c.Assert(err, check.IsNil)
	c.Assert(len(current.Results), check.Equals, 1204)
	mails, err := GetMailLogsByCampaign(campaign.Id)
	c.Assert(err, check.IsNil)
	c.Assert(LockMailLogs(mails, false), check.IsNil)
	claims := make(chan []*MailLog, 16)
	claimErrors := make(chan error, 16)
	gate = make(chan struct{})
	for i := 0; i < 16; i++ {
		candidates := make([]*MailLog, len(mails))
		for j, m := range mails {
			v := *m
			candidates[j] = &v
		}
		wg.Add(1)
		go func(ms []*MailLog) {
			defer wg.Done()
			<-gate
			got, e := ClaimMailLogs(ms)
			claims <- got
			claimErrors <- e
		}(candidates)
	}
	close(gate)
	wg.Wait()
	close(claims)
	close(claimErrors)
	for err := range claimErrors {
		c.Assert(err, check.IsNil)
	}
	seen := map[int64]bool{}
	for batch := range claims {
		for _, m := range batch {
			c.Assert(seen[m.Id], check.Equals, false)
			seen[m.Id] = true
		}
	}
	c.Assert(len(seen), check.Equals, 1204)
	c.Logf("SQLite in-memory, one DB connection: 8 concurrent appends of 1200 addresses + 16 claimers for 1204 records: %s", time.Since(start))
}

func (s *ModelsSuite) TestCompletionCannotResurrectClaimedMail(c *check.C) {
	campaign := s.createCampaign(c)
	c.Assert(SetCampaignLongTerm(campaign.Id, 1, true), check.IsNil)
	mails, err := GetMailLogsByCampaign(campaign.Id)
	c.Assert(err, check.IsNil)
	var wg sync.WaitGroup
	gate := make(chan struct{})
	done := make(chan error, 1)
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-gate
		_, _ = AppendCampaignRecipients(campaign.Id, 1, AppendRecipientsRequest{Targets: []BaseRecipient{{Email: "race@example.com"}}})
	}()
	go func() { defer wg.Done(); <-gate; done <- CompleteCampaign(campaign.Id, 1) }()
	close(gate)
	wg.Wait()
	c.Assert(<-done, check.IsNil)
	c.Assert(LockMailLogs(mails, false), check.IsNil)
	c.Assert(mails[0].Unlock(), check.IsNil)
	c.Assert(mails[0].Backoff(fmt.Errorf("temporary")), check.IsNil)
	pending, err := GetMailLogsByCampaign(campaign.Id)
	c.Assert(err, check.IsNil)
	c.Assert(len(pending), check.Equals, 0)
	allowed, err := mails[0].ShouldSend()
	c.Assert(err, check.IsNil)
	c.Assert(allowed, check.Equals, false)
	c.Assert(campaign.UpdateStatus(CampaignInProgress), check.IsNil)
	current, err := GetCampaign(campaign.Id, 1)
	c.Assert(err, check.IsNil)
	c.Assert(current.Status, check.Equals, CampaignComplete)
	_, err = AppendCampaignRecipients(campaign.Id, 1, AppendRecipientsRequest{Targets: []BaseRecipient{{Email: "too-late@example.com"}}})
	c.Assert(err, check.NotNil)
}

func (s *ModelsSuite) TestCachedSmtpFromAndModeMetadata(c *check.C) {
	m := MailLog{cachedCampaign: &Campaign{SMTP: SMTP{FromAddress: "broken address"}}}
	_, err := m.GetSmtpFrom()
	c.Assert(err, check.NotNil)
	m.cachedCampaign.SMTP.FromAddress = "Sender <sender@example.com>"
	from, err := m.GetSmtpFrom()
	c.Assert(err, check.IsNil)
	c.Assert(from, check.Equals, "sender@example.com")
	campaign := s.createCampaign(c)
	c.Assert(SetCampaignLongTerm(campaign.Id, 1, true), check.IsNil)
	results, err := GetCampaignResults(campaign.Id, 1)
	c.Assert(err, check.IsNil)
	c.Assert(results.LongTerm, check.Equals, true)
	summaries, err := GetCampaignSummaries(1)
	c.Assert(err, check.IsNil)
	c.Assert(summaries.Campaigns[0].LongTerm, check.Equals, true)
	summary, err := GetCampaignSummary(campaign.Id, 1)
	c.Assert(err, check.IsNil)
	c.Assert(summary.LongTerm, check.Equals, true)
}

func (s *ModelsSuite) TestResultIDStopsOnDatabaseError(c *check.C) {
	tx := db.Begin()
	c.Assert(tx.Error, check.IsNil)
	c.Assert(tx.Rollback().Error, check.IsNil)
	r := Result{}
	c.Assert(r.GenerateId(tx), check.NotNil)
}

func (s *ModelsSuite) TestStaleClaimRespectsBackoffAndDeletion(c *check.C) {
	campaign := s.createCampaign(c)
	mails, err := GetMailLogsByCampaign(campaign.Id)
	c.Assert(err, check.IsNil)
	c.Assert(LockMailLogs(mails, false), check.IsNil)
	stale := *mails[0]
	c.Assert(db.Model(&MailLog{}).Where("id=?", stale.Id).UpdateColumn("send_date", time.Now().Add(time.Hour)).Error, check.IsNil)
	claimed, err := ClaimMailLogs([]*MailLog{&stale})
	c.Assert(err, check.IsNil)
	c.Assert(len(claimed), check.Equals, 0)
	c.Assert(DeleteCampaign(campaign.Id), check.IsNil)
	claimed, err = ClaimMailLogs(mails)
	c.Assert(err, check.IsNil)
	c.Assert(len(claimed), check.Equals, 0)
	c.Assert(LockMailLogs(mails, false), check.IsNil)
	current, err := GetMailLogsByCampaign(campaign.Id)
	c.Assert(err, check.IsNil)
	c.Assert(len(current), check.Equals, 0)
}

func (s *ModelsSuite) TestAggregatedCampaignStatsAndOwnership(c *check.C) {
	campaign := s.createCampaign(c)
	statuses := []string{EventDataSubmit, EventClicked, EventOpened, EventSent}
	for i, r := range campaign.Results {
		c.Assert(db.Model(&Result{}).Where("r_id=?", r.RId).Updates(map[string]interface{}{"status": statuses[i], "reported": i == 0}).Error, check.IsNil)
	}
	expected := CampaignStats{Total: 4, EmailsSent: 4, OpenedEmail: 3, ClickedLink: 2, SubmittedData: 1, EmailReported: 1}
	summary, err := GetCampaignSummary(campaign.Id, 1)
	c.Assert(err, check.IsNil)
	c.Assert(summary.Stats, check.Equals, expected)
	other := s.createCampaign(c)
	c.Assert(db.Model(&Campaign{}).Where("id=?", other.Id).UpdateColumn("user_id", 99).Error, check.IsNil)
	overview, err := GetCampaignSummaries(1)
	c.Assert(err, check.IsNil)
	c.Assert(len(overview.Campaigns), check.Equals, 1)
	c.Assert(overview.Campaigns[0].Stats, check.Equals, expected)
	empty, err := GetCampaignSummaries(12345)
	c.Assert(err, check.IsNil)
	c.Assert(empty.Total, check.Equals, int64(0))
}
