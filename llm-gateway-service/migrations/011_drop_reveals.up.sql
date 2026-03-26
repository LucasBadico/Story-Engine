-- Reveals table is orphaned: main service replaced reveals with plots system (task-1 refactor #21).
-- Plot developments are ingested as events, not reveals. Nothing writes to this table.
DROP TABLE IF EXISTS reveals;
