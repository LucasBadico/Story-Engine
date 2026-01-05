ALTER TABLE worlds
ADD COLUMN time_config JSONB DEFAULT '{
  "base_unit": "year",
  "hours_per_day": 24,
  "days_per_week": 7,
  "days_per_year": 365,
  "months_per_year": 12
}'::jsonb;

