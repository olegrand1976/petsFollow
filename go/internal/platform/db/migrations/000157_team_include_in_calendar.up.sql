-- Toggle: include team member as a calendar column / assignee option.
ALTER TABLE practice.team_members
    ADD COLUMN IF NOT EXISTS include_in_calendar BOOLEAN NOT NULL DEFAULT true;
