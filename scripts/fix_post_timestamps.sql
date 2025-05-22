-- Fix zero time values in posts table
UPDATE posts 
SET 
    created_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE created_at = '0001-01-01 00:00:00'::timestamp 
   OR updated_at = '0001-01-01 00:00:00'::timestamp; 