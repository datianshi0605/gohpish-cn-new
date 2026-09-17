package worker

import (
	"context"
	"time"

	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/mailer"
	"github.com/gophish/gophish/models"
	"github.com/sirupsen/logrus"
)

// Worker is an interface that defines the operations needed for a background worker
type Worker interface {
	Start()
	LaunchCampaign(c models.Campaign)
	SendTestEmail(s *models.EmailRequest) error
}

// DefaultWorker is the background worker that handles watching for new campaigns and sending emails appropriately.
type DefaultWorker struct {
	mailer mailer.Mailer
}

// New creates a new worker object to handle the creation of campaigns
func New(options ...func(Worker) error) (Worker, error) {
	defaultMailer := mailer.NewMailWorker()
	w := &DefaultWorker{
		mailer: defaultMailer,
	}
	for _, opt := range options {
		if err := opt(w); err != nil {
			return nil, err
		}
	}
	return w, nil
}

// WithMailer sets the mailer for a given worker.
// By default, workers use a standard, default mailworker.
func WithMailer(m mailer.Mailer) func(*DefaultWorker) error {
	return func(w *DefaultWorker) error {
		w.mailer = m
		return nil
	}
}

// processCampaigns loads maillogs scheduled to be sent before the provided
// time and sends them to the mailer.
func (w *DefaultWorker) processCampaigns(t time.Time) error {
	candidates, err := models.GetQueuedMailLogs(t.UTC())
	if err != nil {
		return err
	}
	ms, err := models.ClaimMailLogs(candidates)
	if err != nil {
		return err
	}
	groups := make(map[int64][]*models.MailLog)
	for _, m := range ms {
		groups[m.CampaignId] = append(groups[m.CampaignId], m)
	}
	var firstError error
	for cid, group := range groups {
		c, err := models.GetCampaignMailContext(cid, group[0].UserId)
		if err == nil && c.Status == models.CampaignQueued {
			err = c.UpdateStatus(models.CampaignInProgress)
		}
		if err != nil || c.Status == models.CampaignComplete {
			if err != nil {
				log.Error(err)
				if firstError == nil {
					firstError = err
				}
			}
			if unlockErr := models.LockMailLogs(group, false); unlockErr != nil {
				log.Error(unlockErr)
				if firstError == nil {
					firstError = unlockErr
				}
			}
			continue
		}
		mails := make([]mailer.Mail, 0, len(group))
		for _, m := range group {
			m.CacheCampaign(&c)
			mails = append(mails, m)
		}
		go func(ms []mailer.Mail) {
			log.WithFields(logrus.Fields{"num_emails": len(ms)}).Info("Sending emails to mailer for processing")
			w.mailer.Queue(ms)
		}(mails)
	}
	return firstError
}

// Start launches the worker to poll the database every minute for any pending maillogs
// that need to be processed.
func (w *DefaultWorker) Start() {
	log.Info("Background Worker Started Successfully - Waiting for Campaigns")
	go w.mailer.Start(context.Background())
	for range time.Tick(1 * time.Minute) {
		if err := models.SyncLongTermCampaigns(); err != nil {
			log.Error(err)
		}
		// Include recipients queued during this tick rather than waiting another minute.
		err := w.processCampaigns(time.Now())
		if err != nil {
			log.Error(err)
			continue
		}
	}
}

// LaunchCampaign starts a campaign
func (w *DefaultWorker) LaunchCampaign(c models.Campaign) {
	ms, err := models.GetMailLogsByCampaign(c.Id)
	if err != nil {
		log.Error(err)
		return
	}
	// Only the original immediately scheduled records are owned by this launch.
	// Future records and newly appended recipients belong to the polling worker.
	initial := make(map[string]bool)
	for _, r := range c.Results {
		initial[r.RId] = true
	}
	owned := []*models.MailLog{}
	now := time.Now().UTC()
	for _, m := range ms {
		if initial[m.RId] && m.Processing && !m.SendDate.After(now) {
			owned = append(owned, m)
		}
	}
	if len(owned) == 0 {
		return
	}
	campaign, err := models.GetCampaignMailContext(c.Id, c.UserId)
	if err != nil || campaign.Status == models.CampaignComplete {
		if err != nil {
			log.Error(err)
		}
		if err := models.LockMailLogs(owned, false); err != nil {
			log.Error(err)
		}
		return
	}
	mails := make([]mailer.Mail, 0, len(owned))
	for _, m := range owned {
		m.CacheCampaign(&campaign)
		mails = append(mails, m)
	}
	w.mailer.Queue(mails)
}

// SendTestEmail sends a test email
func (w *DefaultWorker) SendTestEmail(s *models.EmailRequest) error {
	go func() {
		ms := []mailer.Mail{s}
		w.mailer.Queue(ms)
	}()
	return <-s.ErrorChan
}
