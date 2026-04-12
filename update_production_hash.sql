-- Update password hashes in production database
UPDATE users 
SET password_hash = '$2a$10$bMPWlUqSeT.VdRH6dC1Ga.pUh.J8gM0UYTIAJm01BWM7pAya.IOAq'
WHERE email IN ('admin@phosboard.cl', 'tenant.admin@example.com');

-- Verify update
SELECT email, LEFT(password_hash, 30) as hash_prefix, LENGTH(password_hash) as hash_len
FROM users 
WHERE email LIKE '%admin%';
