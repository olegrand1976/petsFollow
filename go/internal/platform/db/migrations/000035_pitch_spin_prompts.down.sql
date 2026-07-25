-- Revert current prompts to v1 seed ids if present.

UPDATE sales.agent_prompt_current
SET version_id = 'a0000000-0000-4000-8000-000000000010'
WHERE agent_kind = 'vet_live'
  AND EXISTS (SELECT 1 FROM sales.agent_prompt_versions WHERE id = 'a0000000-0000-4000-8000-000000000010');

UPDATE sales.agent_prompt_current
SET version_id = 'a0000000-0000-4000-8000-000000000011'
WHERE agent_kind = 'coach'
  AND EXISTS (SELECT 1 FROM sales.agent_prompt_versions WHERE id = 'a0000000-0000-4000-8000-000000000011');

DELETE FROM sales.agent_prompt_versions
WHERE id IN (
    'a0000000-0000-4000-8000-000000000012',
    'a0000000-0000-4000-8000-000000000013'
);
