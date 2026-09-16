
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE smtp ADD COLUMN dingtalk_app_key varchar(255);
ALTER TABLE smtp ADD COLUMN dingtalk_app_secret varchar(255);
ALTER TABLE smtp ADD COLUMN dingtalk_agent_id varchar(255);
ALTER TABLE smtp ADD COLUMN dingtalk_recipient_mode varchar(64);
ALTER TABLE smtp ADD COLUMN dingtalk_user_mappings text;
ALTER TABLE smtp ADD COLUMN dingtalk_message_template text;
ALTER TABLE smtp ADD COLUMN dingtalk_include_qr_code boolean DEFAULT false;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

