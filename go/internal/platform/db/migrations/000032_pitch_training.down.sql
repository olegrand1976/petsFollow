DROP TABLE IF EXISTS sales.pitch_sim_feedback;
DROP TABLE IF EXISTS sales.pitch_simulations;
ALTER TABLE sales.agent_prompt_versions DROP CONSTRAINT IF EXISTS agent_prompt_versions_analyzer_run_fk;
DROP TABLE IF EXISTS sales.pitch_analyzer_runs;
DROP TABLE IF EXISTS sales.agent_prompt_current;
DROP TABLE IF EXISTS sales.agent_prompt_versions;
DROP TABLE IF EXISTS sales.pitch_scripts;
