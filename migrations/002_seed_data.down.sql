-- Down Migration

DELETE FROM galleries WHERE title = 'Main';
DELETE FROM users WHERE username = 'igor';