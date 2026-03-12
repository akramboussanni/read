-- Add assignment_type to support path-progress assignments
-- 'quiz'           : existing type - student must complete a specific quiz node
-- 'path_progress'  : new type     - student must reach/complete a specific node in the course parcours

ALTER TABLE classroom_assignments ADD COLUMN assignment_type TEXT NOT NULL DEFAULT 'quiz';
