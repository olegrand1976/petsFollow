-- Clear legacy temporary journey skips so steps can retry after pref/eligibility changes.
DELETE FROM discovery.email_sends
WHERE status = 'skipped'
  AND COALESCE(meta->>'reason', '') IN ('pref_off', 'not_eligible', '');
