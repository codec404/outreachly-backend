-- Reusable trigger function that auto-updates updated_at on any row change.
-- Applied via per-table triggers in subsequent migrations.
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
