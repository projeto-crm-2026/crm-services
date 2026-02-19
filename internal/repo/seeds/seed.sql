-- Seed: test organization and admin user
-- Password: password123 (bcrypt hash)

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM "user" WHERE email = 'admin@test.com' AND deleted_at IS NULL) THEN
        RAISE NOTICE 'Seed data already exists, skipping.';
        RETURN;
    END IF;

    INSERT INTO organizations (uuid, name, slug, email, phone, website, industry, plan, is_active, created_at, updated_at)
    VALUES (
        gen_random_uuid(),
        'Test Organization',
        'test-organization',
        'contact@testorg.com',
        '+5511999999999',
        'https://testorg.com',
        'Technology',
        'free',
        true,
        NOW(),
        NOW()
    );

    INSERT INTO "user" (uuid, organization_id, name, email, password_hash, role, status, created_at, updated_at)
    VALUES (
        gen_random_uuid(),
        (SELECT uuid FROM organizations WHERE slug = 'test-organization' AND deleted_at IS NULL LIMIT 1),
        'Admin User',
        'admin@test.com',
        '$2a$14$JPJpuGo99NDZ4S9i92v3h.D88MgWtn9jQX1soI/mBwGKdPbXgZHH.',
        'admin',
        'active',
        NOW(),
        NOW()
    );

    RAISE NOTICE 'Seed data created successfully.';
END $$;
