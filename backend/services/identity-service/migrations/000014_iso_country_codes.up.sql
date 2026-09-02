UPDATE identity.business_registry_records
SET country_code = 'GB',
    registration_number = 'GB-' || substring(registration_number from 4)
WHERE country_code = 'UK';
