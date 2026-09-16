package models

import (
	"errors"
	"fmt"

	log "github.com/gophish/gophish/logger"
	"github.com/jinzhu/gorm"
)

// CampaignAutoGroup stores stable IDs so renaming a group keeps its subscription.
type CampaignAutoGroup struct {
	CampaignId int64 `gorm:"primary_key;AUTO_INCREMENT:false"`
	GroupId    int64 `gorm:"primary_key;AUTO_INCREMENT:false"`
}

func GetCampaignAutoGroups(id, uid int64) ([]Group, error) {
	var campaign Campaign
	if err := db.Where("id=? AND user_id=?", id, uid).First(&campaign).Error; err != nil {
		return nil, err
	}
	groups := []Group{}
	err := db.Table("groups").Select("`groups`.*").Joins("JOIN campaign_auto_groups ON campaign_auto_groups.group_id = `groups`.id").Where("campaign_auto_groups.campaign_id=? AND `groups`.user_id=?", id, uid).Order("`groups`.id").Scan(&groups).Error
	return groups, err
}

func SetCampaignAutoGroups(id, uid int64, ids []int64) error {
	if len(ids) > 1000 {
		return errors.New("最多关联 1000 个用户组")
	}
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()
	if err := tx.Model(&Campaign{}).Where("id=? AND user_id=?", id, uid).UpdateColumn("id", id).Error; err != nil {
		return err
	}
	var campaign Campaign
	if err := tx.Where("id=? AND user_id=?", id, uid).First(&campaign).Error; err != nil {
		return err
	}
	if !campaign.LongTerm || campaign.Status == CampaignComplete {
		return errors.New("仅未完成的长期演练可以关联用户组")
	}
	unique := map[int64]bool{}
	for _, gid := range ids {
		var group Group
		if err := tx.Where("id=? AND user_id=?", gid, uid).First(&group).Error; err != nil {
			return errors.New("用户组不存在或无权访问")
		}
		unique[gid] = true
	}
	if err := tx.Where("campaign_id=?", id).Delete(&CampaignAutoGroup{}).Error; err != nil {
		return err
	}
	for gid := range unique {
		if err := tx.Create(&CampaignAutoGroup{CampaignId: id, GroupId: gid}).Error; err != nil {
			return err
		}
	}
	return tx.Commit().Error
}

func campaignAutoTargets(tx *gorm.DB, id, uid int64) ([]BaseRecipient, error) {
	targets := []BaseRecipient{}
	// Join current group membership inside the campaign transaction. Removed or
	// deleted subscriptions are never followed using an earlier cached snapshot.
	err := tx.Table("targets").Select("DISTINCT targets.email, targets.first_name, targets.last_name, targets.position").
		Joins("JOIN group_targets ON group_targets.target_id = targets.id").
		Joins("JOIN `groups` ON `groups`.id = group_targets.group_id").
		Joins("JOIN campaign_auto_groups ON campaign_auto_groups.group_id = `groups`.id").
		Where("campaign_auto_groups.campaign_id=? AND `groups`.user_id=?", id, uid).Scan(&targets).Error
	return targets, err
}

// SyncLongTermCampaigns is called by the existing worker every minute. Failure in
// one campaign does not prevent other campaigns or queued deliveries running.
func SyncLongTermCampaigns() error {
	var campaigns []Campaign
	if err := db.Where("long_term=? AND status<>?", true, CampaignComplete).Find(&campaigns).Error; err != nil {
		return err
	}
	var firstError error
	for _, c := range campaigns {
		if _, err := AppendCampaignRecipients(c.Id, c.UserId, AppendRecipientsRequest{autoSync: true}); err != nil {
			log.Errorf("Automatic recipient sync failed for campaign %d: %v", c.Id, err)
			if firstError == nil {
				firstError = fmt.Errorf("campaign %d: %v", c.Id, err)
			}
		}
	}
	return firstError
}
