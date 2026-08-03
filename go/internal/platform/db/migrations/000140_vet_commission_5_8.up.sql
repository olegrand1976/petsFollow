-- Vet progressive tiers: base 7.50% → 8% (clients 1–10 / 11–30 / 31–60 / 61+).
-- Effective after plan factor (×1 triennial / ×0.67 monthly·annual) stays in [5%, 8%].
DELETE FROM billing.commission_tiers;

INSERT INTO billing.commission_tiers (id, min_clients, max_clients, rate_bps) VALUES
  ('a0000000-0000-4000-8000-000000000140', 1, 10, 750),
  ('a0000000-0000-4000-8000-000000000141', 11, 30, 767),
  ('a0000000-0000-4000-8000-000000000142', 31, 60, 783),
  ('a0000000-0000-4000-8000-000000000143', 61, NULL, 800);
