-- Restore previous progressive ladder: 5% → 8% → 10% → 12%.
DELETE FROM billing.commission_tiers;

INSERT INTO billing.commission_tiers (id, min_clients, max_clients, rate_bps) VALUES
  ('b0000000-0000-4000-8000-000000000019', 1, 4, 500),
  ('b0000000-0000-4000-8000-000000000020', 5, 14, 800),
  ('b0000000-0000-4000-8000-000000000021', 15, 39, 1000),
  ('b0000000-0000-4000-8000-000000000022', 40, NULL, 1200);
