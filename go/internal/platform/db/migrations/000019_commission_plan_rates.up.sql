-- Progressive vet tiers: 7% → 9% → 11% → 12% (clients 1–10 / 11–30 / 31–60 / 61+).
DELETE FROM billing.commission_tiers;

INSERT INTO billing.commission_tiers (id, min_clients, max_clients, rate_bps) VALUES
  ('a0000000-0000-4000-8000-000000000019', 1, 10, 700),
  ('a0000000-0000-4000-8000-000000000020', 11, 30, 900),
  ('a0000000-0000-4000-8000-000000000021', 31, 60, 1100),
  ('a0000000-0000-4000-8000-000000000022', 61, NULL, 1200);
