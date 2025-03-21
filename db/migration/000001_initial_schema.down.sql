-- Drop triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_proxies_updated_at ON proxies;
DROP TRIGGER IF EXISTS update_certificates_updated_at ON certificates;
DROP TRIGGER IF EXISTS update_traffic_stats_updated_at ON traffic_stats;
DROP TRIGGER IF EXISTS update_daily_stats_updated_at ON daily_stats;
DROP TRIGGER IF EXISTS update_events_updated_at ON events;
DROP TRIGGER IF EXISTS update_backups_updated_at ON backups;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_proxies_user_id;
DROP INDEX IF EXISTS idx_proxies_port;
DROP INDEX IF EXISTS idx_certificates_domain;
DROP INDEX IF EXISTS idx_traffic_stats_user_id;
DROP INDEX IF EXISTS idx_daily_stats_user_id;
DROP INDEX IF EXISTS idx_daily_stats_date;
DROP INDEX IF EXISTS idx_events_user_id;
DROP INDEX IF EXISTS idx_events_created_at;
DROP INDEX IF EXISTS idx_backups_status;
DROP INDEX IF EXISTS idx_backups_timestamp;

-- Drop tables
DROP TABLE IF EXISTS backups;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS daily_stats;
DROP TABLE IF EXISTS traffic_stats;
DROP TABLE IF EXISTS certificates;
DROP TABLE IF EXISTS proxies;
DROP TABLE IF EXISTS users; 