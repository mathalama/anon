ALTER TABLE reports ADD CONSTRAINT unique_reporter_reported UNIQUE (reporter_user_id, reported_user_id);
