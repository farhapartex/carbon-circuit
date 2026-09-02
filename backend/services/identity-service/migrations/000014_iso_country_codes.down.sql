UPDATE identity.business_registry_records
SET country_code = 'UK',
    registration_number = 'UK-' || substring(registration_number from 4)
WHERE country_code = 'GB';
