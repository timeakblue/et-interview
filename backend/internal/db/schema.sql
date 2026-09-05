-- Database schema, applied at startup and on every `import` run.
--
-- Statements must be idempotent (CREATE TABLE IF NOT EXISTS, INSERT OR
-- IGNORE): this file runs every time the process starts, against a database
-- that may already have been migrated.
--
-- Add your tables here.
-- backend/internal/db/schema.sql

-- Drop the old demo table if it exists
DROP TABLE IF EXISTS demo_values;

-- Main table for DCIP survey readings
CREATE TABLE IF NOT EXISTS readings (
    -- Identity
    id TEXT PRIMARY KEY,
    line_id TEXT NOT NULL,
    
    -- Transmitting Pair (tx1, tx2)
    tx1_id TEXT NOT NULL,
    tx1_lat REAL NOT NULL,
    tx1_lon REAL NOT NULL,
    tx1_alt REAL NOT NULL,
    tx2_id TEXT NOT NULL,
    tx2_lat REAL NOT NULL,
    tx2_lon REAL NOT NULL,
    tx2_alt REAL NOT NULL,

    -- Receiving Pair (rx1, rx2)
    rx1_id TEXT NOT NULL,
    rx1_lat REAL NOT NULL,
    rx1_lon REAL NOT NULL,
    rx1_alt REAL NOT NULL,
    rx2_id TEXT NOT NULL,
    rx2_lat REAL NOT NULL,
    rx2_lon REAL NOT NULL,
    rx2_alt REAL NOT NULL,

    -- Acquisition Metadata
    timestamp TEXT NOT NULL,
    array_type TEXT NOT NULL,
    symmetry INTEGER NOT NULL,
    stacks INTEGER NOT NULL,
    input_current REAL NOT NULL,
    contact_resistance REAL NOT NULL,

    -- Measurements
    apparent_resistivity REAL NOT NULL,
    apparent_resistivity_err REAL NOT NULL,
    chargeability REAL NOT NULL,
    chargeability_err REAL NOT NULL,
    decay_curve TEXT NOT NULL, -- Stored as semicolon-separated string or JSON array

    -- QC Review Status ('pass', 'flag', 'reject', or NULL)
    qc_review TEXT DEFAULT '',

    -- Generated Geometry Key for fast 4-electrode repeat grouping
    geometry_key TEXT GENERATED ALWAYS AS (
        tx1_id || '_' || tx2_id || '_' || rx1_id || '_' || rx2_id
    ) STORED
);

-- Index for fast queries by Line
CREATE INDEX IF NOT EXISTS idx_readings_line_id ON readings(line_id);

-- Index for fast grouping and variance calculations across repeat geometries
CREATE INDEX IF NOT EXISTS idx_readings_geometry_key ON readings(geometry_key);
