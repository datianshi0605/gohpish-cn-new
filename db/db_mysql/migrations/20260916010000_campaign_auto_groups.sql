-- +goose Up
CREATE TABLE campaign_auto_groups (
 campaign_id bigint NOT NULL,
 group_id bigint NOT NULL,
 PRIMARY KEY (campaign_id, group_id)
);

-- +goose Down
DROP TABLE campaign_auto_groups;
