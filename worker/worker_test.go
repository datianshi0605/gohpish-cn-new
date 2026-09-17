package worker

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/mailer"
	"github.com/gophish/gophish/models"
)

type logMailer struct {
	queue chan []mailer.Mail
}

func (m *logMailer) Start(ctx context.Context) {}

func (m *logMailer) Queue(ms []mailer.Mail) {
	m.queue <- ms
}

// testContext is context to cover API related functions
type testContext struct {
	config *config.Config
}

func setupTest(t *testing.T) *testContext {
	conf := &config.Config{
		DBName:         "sqlite3",
		DBPath:         ":memory:",
		MigrationsPath: "../db/db_sqlite3/migrations/",
	}
	err := models.Setup(conf)
	if err != nil {
		t.Fatalf("Failed creating database: %v", err)
	}
	ctx := &testContext{}
	ctx.config = conf
	createTestData(t, ctx)
	return ctx
}

func createTestData(t *testing.T, ctx *testContext) {
	ctx.config.TestFlag = true
	// Add a group
	group := models.Group{Name: "Test Group"}
	for i := 0; i < 10; i++ {
		group.Targets = append(group.Targets, models.Target{
			BaseRecipient: models.BaseRecipient{
				Email:     fmt.Sprintf("test%d@example.com", i),
				FirstName: "First",
				LastName:  "Example"}})
	}
	group.UserId = 1
	models.PostGroup(&group)

	// Add a template
	template := models.Template{Name: "Test Template"}
	template.Subject = "Test subject"
	template.Text = "Text text"
	template.HTML = "<html>Test</html>"
	template.UserId = 1
	models.PostTemplate(&template)

	// Add a landing page
	p := models.Page{Name: "Test Page"}
	p.HTML = "<html>Test</html>"
	p.UserId = 1
	models.PostPage(&p)

	// Add a sending profile
	smtp := models.SMTP{Name: "Test Page"}
	smtp.UserId = 1
	smtp.Host = "example.com"
	smtp.FromAddress = "test@test.com"
	models.PostSMTP(&smtp)
}

func setupCampaign(id int) (*models.Campaign, error) {
	// Setup and "launch" our campaign
	// Set the status such that no emails are attempted
	c := models.Campaign{Name: fmt.Sprintf("Test campaign - %d", id)}
	c.UserId = 1
	template, err := models.GetTemplate(1, 1)
	if err != nil {
		return nil, err
	}
	c.Template = template

	page, err := models.GetPage(1, 1)
	if err != nil {
		return nil, err
	}
	c.Page = page

	smtp, err := models.GetSMTP(1, 1)
	if err != nil {
		return nil, err
	}
	c.SMTP = smtp

	group, err := models.GetGroup(1, 1)
	if err != nil {
		return nil, err
	}
	c.Groups = []models.Group{group}
	err = models.PostCampaign(&c, c.UserId)
	if err != nil {
		return nil, err
	}
	err = c.UpdateStatus(models.CampaignEmailsSent)
	return &c, err
}

func TestMailLogGrouping(t *testing.T) {
	setupTest(t)

	// Create the campaigns and unlock the maillogs so that they're picked up
	// by the worker
	for i := 0; i < 10; i++ {
		campaign, err := setupCampaign(i)
		if err != nil {
			t.Fatalf("error creating campaign: %v", err)
		}
		ms, err := models.GetMailLogsByCampaign(campaign.Id)
		if err != nil {
			t.Fatalf("error getting maillogs for campaign: %v", err)
		}
		for _, m := range ms {
			m.Unlock()
		}
	}

	lm := &logMailer{queue: make(chan []mailer.Mail)}
	worker := &DefaultWorker{}
	worker.mailer = lm

	// Trigger the worker, generating the maillogs and sending them to the
	// mailer
	worker.processCampaigns(time.Now())

	// Verify that each slice of maillogs received belong to the same campaign
	for i := 0; i < 10; i++ {
		ms := <-lm.queue
		maillog, ok := ms[0].(*models.MailLog)
		if !ok {
			t.Fatalf("unable to cast mail to models.MailLog")
		}
		expected := maillog.CampaignId
		for _, m := range ms {
			maillog, ok = m.(*models.MailLog)
			if !ok {
				t.Fatalf("unable to cast mail to models.MailLog")
			}
			got := maillog.CampaignId
			if got != expected {
				t.Fatalf("unexpected campaign ID received for maillog: got %d expected %d", got, expected)
			}
		}
	}
}

func TestAutomaticGroupRecipientsReachWorker(t *testing.T) {
	setupTest(t)
	campaign, err := setupCampaign(1)
	if err != nil {
		t.Fatal(err)
	}
	if err = models.SetCampaignLongTerm(campaign.Id, 1, true); err != nil {
		t.Fatal(err)
	}
	if err = models.SetCampaignAutoGroups(campaign.Id, 1, []int64{1}); err != nil {
		t.Fatal(err)
	}
	group, err := models.GetGroup(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	group.Targets = append(group.Targets, models.Target{BaseRecipient: models.BaseRecipient{Email: "new.joiner@example.com"}})
	if err = models.PutGroup(&group); err != nil {
		t.Fatal(err)
	}
	if err = models.SyncLongTermCampaigns(); err != nil {
		t.Fatal(err)
	}
	lm := &logMailer{queue: make(chan []mailer.Mail, 2)}
	w := &DefaultWorker{mailer: lm}
	if err = w.processCampaigns(time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	select {
	case mails := <-lm.queue:
		if len(mails) != 1 {
			t.Fatalf("expected only new recipient, got %d", len(mails))
		}
		ml := mails[0].(*models.MailLog)
		result, err := models.GetResult(ml.RId)
		if err != nil || result.Email != "new.joiner@example.com" {
			t.Fatal("wrong recipient", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("new recipient never reached worker")
	}
	if err = models.SyncLongTermCampaigns(); err != nil {
		t.Fatal(err)
	}
	logs, err := models.GetMailLogsByCampaign(campaign.Id)
	if err != nil || len(logs) != 11 {
		t.Fatal("repeat synchronization duplicated mail logs", len(logs), err)
	}
}

func TestConcurrentWorkerClaimsDoNotDuplicate(t *testing.T) {
	setupTest(t)
	c, err := setupCampaign(1)
	if err != nil {
		t.Fatal(err)
	}
	ms, err := models.GetMailLogsByCampaign(c.Id)
	if err != nil {
		t.Fatal(err)
	}
	if err = models.LockMailLogs(ms, false); err != nil {
		t.Fatal(err)
	}
	lm := &logMailer{queue: make(chan []mailer.Mail, 16)}
	var wg sync.WaitGroup
	errors := make(chan error, 8)
	gate := make(chan struct{})
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-gate
			w := &DefaultWorker{mailer: lm}
			errors <- w.processCampaigns(time.Now())
		}()
	}
	close(gate)
	wg.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	seen := map[string]bool{}
	deadline := time.After(2 * time.Second)
	for len(seen) < 10 {
		select {
		case batch := <-lm.queue:
			for _, mail := range batch {
				id := mail.(*models.MailLog).RId
				if seen[id] {
					t.Fatal("duplicate mail queued")
				}
				seen[id] = true
			}
		case <-deadline:
			t.Fatal("missing mail")
		}
	}
	select {
	case <-lm.queue:
		t.Fatal("unexpected duplicate batch")
	case <-time.After(20 * time.Millisecond):
	}
}

func TestImmediateLaunchDoesNotTakeAutoAppendedMail(t *testing.T) {
	setupTest(t)
	c, err := setupCampaign(1)
	if err != nil {
		t.Fatal(err)
	}
	if err = models.SetCampaignLongTerm(c.Id, 1, true); err != nil {
		t.Fatal(err)
	}
	_, err = models.AppendCampaignRecipients(c.Id, 1, models.AppendRecipientsRequest{Targets: []models.BaseRecipient{{Email: "later-added@example.com"}}})
	if err != nil {
		t.Fatal(err)
	}
	lm := &logMailer{queue: make(chan []mailer.Mail, 4)}
	w := &DefaultWorker{mailer: lm}
	w.LaunchCampaign(*c)
	batch := <-lm.queue
	if len(batch) != 10 {
		t.Fatalf("launch took appended mail: %d", len(batch))
	}
	if err = w.processCampaigns(time.Now()); err != nil {
		t.Fatal(err)
	}
	select {
	case batch = <-lm.queue:
		if len(batch) != 1 {
			t.Fatal("append was duplicated or lost")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("append not queued")
	}
}

func TestBrokenCampaignDoesNotLockOtherCampaigns(t *testing.T) {
	setupTest(t)
	bad, err := setupCampaign(1)
	if err != nil {
		t.Fatal(err)
	}
	template := models.Template{Name: "Healthy Template", UserId: 1, Subject: "test", Text: "test"}
	if err = models.PostTemplate(&template); err != nil {
		t.Fatal(err)
	}
	good := *bad
	good.Id = 0
	good.Name = "Healthy campaign"
	good.Template = template
	good.Results = nil
	if err = models.PostCampaign(&good, 1); err != nil {
		t.Fatal(err)
	}
	for _, id := range []int64{bad.Id, good.Id} {
		ms, e := models.GetMailLogsByCampaign(id)
		if e != nil {
			t.Fatal(e)
		}
		if e = models.LockMailLogs(ms, false); e != nil {
			t.Fatal(e)
		}
	}
	if err = models.DeleteTemplate(bad.Template.Id, 1); err != nil {
		t.Fatal(err)
	}
	lm := &logMailer{queue: make(chan []mailer.Mail, 2)}
	w := &DefaultWorker{mailer: lm}
	if err = w.processCampaigns(time.Now()); err == nil {
		t.Fatal("missing template must return an error")
	}
	select {
	case batch := <-lm.queue:
		if len(batch) != 10 || batch[0].(*models.MailLog).CampaignId != good.Id {
			t.Fatal("healthy campaign was not queued")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("healthy campaign blocked")
	}
	ms, err := models.GetMailLogsByCampaign(bad.Id)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range ms {
		if m.Processing {
			t.Fatal("broken campaign remains permanently locked")
		}
	}
}
