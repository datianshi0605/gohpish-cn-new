package models

import (
	"errors"
	"net/mail"
	"strings"
	"time"
)

// AppendCampaignRecipients adds only new recipients. The existing worker picks
// up committed mail logs; never relaunch a campaign (which can resend old mail).
type AppendRecipientsRequest struct {
	autoSync bool // internal worker request; never accepted from JSON

	Targets  []BaseRecipient `json:"targets"`
	Groups   []Group         `json:"groups"`
	SendDate time.Time       `json:"send_date"`
}
type AppendRecipientsResponse struct {
	Added   int `json:"added"`
	Skipped int `json:"skipped"`
}

func SetCampaignLongTerm(id, uid int64, enabled bool) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()
	// An UPDATE obtains the campaign row lock on both supported databases.
	if err := tx.Model(&Campaign{}).Where("id=? AND user_id=?", id, uid).UpdateColumn("id", id).Error; err != nil {
		return err
	}
	var c Campaign
	if err := tx.Where("id=? AND user_id=?", id, uid).First(&c).Error; err != nil {
		return err
	}
	if c.Status == CampaignComplete {
		return errors.New("已完成的活动不能更改长期演练设置")
	}
	if err := tx.Model(&c).UpdateColumn("long_term", enabled).Error; err != nil {
		return err
	}
	return tx.Commit().Error
}

func AppendCampaignRecipients(id, uid int64, req AppendRecipientsRequest) (AppendRecipientsResponse, error) {
	out := AppendRecipientsResponse{}
	targets := append([]BaseRecipient{}, req.Targets...)
	for _, ref := range req.Groups {
		g, err := GetGroupByName(ref.Name, uid)
		if err != nil {
			return out, err
		}
		for _, t := range g.Targets {
			targets = append(targets, t.BaseRecipient)
		}
	}
	tx := db.Begin()
	if tx.Error != nil {
		return out, tx.Error
	}
	defer tx.Rollback()
	if err := tx.Model(&Campaign{}).Where("id=? AND user_id=?", id, uid).UpdateColumn("id", id).Error; err != nil {
		return out, err
	}
	var c Campaign
	if err := tx.Where("id=? AND user_id=?", id, uid).First(&c).Error; err != nil {
		return out, err
	}
	if !c.LongTerm || c.Status == CampaignComplete {
		return out, errors.New("仅未完成的长期演练支持追加人员")
	}
	if req.autoSync {
		var err error
		targets, err = campaignAutoTargets(tx, id, uid)
		if err != nil {
			return out, err
		}
		if len(targets) == 0 {
			return out, nil
		}
	}
	if len(targets) == 0 {
		return out, errors.New("请添加人员或选择用户组")
	}
	if !req.autoSync && len(targets) > 10000 {
		return out, errors.New("每次最多追加 10000 人")
	}
	for i := range targets {
		targets[i].Email = strings.TrimSpace(targets[i].Email)
		addr, err := mail.ParseAddress(targets[i].Email)
		if err != nil || addr.Address != targets[i].Email {
			return out, errors.New("名单包含无效邮箱，请检查后重试")
		}
	}
	now := time.Now().UTC()
	sendDate := req.SendDate.UTC()
	if sendDate.IsZero() || sendDate.Before(now) {
		sendDate = now
	}
	if sendDate.Before(c.LaunchDate) {
		sendDate = c.LaunchDate
	}
	var existing []Result
	if err := tx.Where("campaign_id=?", id).Find(&existing).Error; err != nil {
		return out, err
	}
	seen := make(map[string]bool)
	for _, r := range existing {
		seen[strings.ToLower(strings.TrimSpace(r.Email))] = true
	}
	for _, t := range targets {
		key := strings.ToLower(t.Email)
		if seen[key] {
			out.Skipped++
			continue
		}
		seen[key] = true
		r := Result{BaseRecipient: t, CampaignId: id, UserId: uid, Status: StatusScheduled, SendDate: sendDate, ModifiedDate: now}
		if err := r.GenerateId(tx); err != nil {
			return AppendRecipientsResponse{}, err
		}
		if err := tx.Save(&r).Error; err != nil {
			return AppendRecipientsResponse{}, err
		}
		m := MailLog{CampaignId: id, UserId: uid, RId: r.RId, SendDate: sendDate, Processing: false}
		if err := tx.Save(&m).Error; err != nil {
			return AppendRecipientsResponse{}, err
		}
		out.Added++
	}
	if err := tx.Commit().Error; err != nil {
		return AppendRecipientsResponse{}, err
	}
	return out, nil
}
