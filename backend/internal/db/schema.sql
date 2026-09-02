-- Database schema, applied at startup and on every `import` run.
--
-- Statements must be idempotent (CREATE TABLE IF NOT EXISTS, INSERT OR
-- IGNORE): this file runs every time the process starts, against a database
-- that may already have been migrated.
--
-- Add your tables here.

-- ---------------------------------------------------------------------------
-- Throwaway: seeds the numbers behind GET /api/demo. Delete this table and
-- internal/demo once you are oriented; it exists to prove the wiring works.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS demo_values (
    position INTEGER PRIMARY KEY,
    value    REAL NOT NULL
);

INSERT OR IGNORE INTO demo_values (position, value) VALUES
    (1, 1), (2, 2), (3, 3), (4, 4), (5, 5);
