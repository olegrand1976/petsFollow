-- Staging / env sans re-seed : la colonne contact_phone ajoutée en 000079 reste
-- vide pour les comptes démo déjà présents → gate /complete-contact-phone casse les e2e.
UPDATE identity.users SET contact_phone = '0470 12 34 56'
WHERE email = 'commercial.demo@petsfollow.test' AND TRIM(contact_phone) = '';

UPDATE identity.users SET contact_phone = '0471 98 76 54'
WHERE email = 'commercial.demo2@petsfollow.test' AND TRIM(contact_phone) = '';

UPDATE identity.users SET contact_phone = '0472 11 22 33'
WHERE email = 'commercial.manager@petsfollow.test' AND TRIM(contact_phone) = '';
