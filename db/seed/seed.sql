-- Seed: plan catalog (per D-01)
INSERT INTO plans (uuid, name, display_name, price_cents, currency, max_contacts, max_members, max_chat_responders, mp_preapproval_plan_id, is_active, created_at, updated_at)
VALUES
    (gen_random_uuid(), 'free', 'Free', 0, 'BRL', 100, 3, 1, NULL, true, NOW(), NOW()),
    (gen_random_uuid(), 'pro', 'Pro', 4990, 'BRL', 1000, 10, 5, NULL, true, NOW(), NOW()),
    (gen_random_uuid(), 'business', 'Business', 14990, 'BRL', 10000, 50, 25, NULL, true, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Seed: permission catalog
INSERT INTO permissions (name, description, category, created_at)
VALUES
    ('members.invite',          'Invite new members to the organization',       'members',        NOW()),
    ('members.list',            'List organization members',                    'members',        NOW()),
    ('members.manage_roles',    'Change a member''s role',                      'members',        NOW()),
    ('members.remove',          'Remove a member from the organization',        'members',        NOW()),
    ('members.deactivate',      'Deactivate or reactivate a member',            'members',        NOW()),
    ('contacts.create',         'Create contacts',                              'contacts',       NOW()),
    ('contacts.read',           'View contacts',                                'contacts',       NOW()),
    ('contacts.update',         'Update contacts',                              'contacts',       NOW()),
    ('contacts.delete',         'Delete contacts',                              'contacts',       NOW()),
    ('chats.read',              'View and access chats',                        'chats',          NOW()),
    ('webhooks.create',         'Create outgoing webhooks',                     'webhooks',       NOW()),
    ('webhooks.read',           'View outgoing webhooks',                       'webhooks',       NOW()),
    ('webhooks.update',         'Update outgoing webhooks',                     'webhooks',       NOW()),
    ('webhooks.delete',         'Delete outgoing webhooks',                     'webhooks',       NOW()),
    ('webhooks.tokens.create',  'Create incoming webhook tokens',               'webhooks',       NOW()),
    ('webhooks.tokens.read',    'List incoming webhook tokens',                 'webhooks',       NOW()),
    ('webhooks.tokens.delete',  'Delete incoming webhook tokens',               'webhooks',       NOW()),
    ('api_keys.create',         'Create API keys',                              'api_keys',       NOW()),
    ('api_keys.read',           'List API keys',                                'api_keys',       NOW()),
    ('api_keys.delete',         'Delete API keys',                              'api_keys',       NOW()),
    ('organizations.read',      'View organization details',                    'organizations',  NOW()),
    ('organizations.update',    'Update organization settings',                 'organizations',  NOW()),
    ('organizations.delete',    'Permanently delete the organization',          'organizations',  NOW()),
    ('organizations.manage',    'Restore and manage organization lifecycle',    'organizations',  NOW()),
    ('plans.read',              'View plans and usage',                          'plans',          NOW()),
    ('subscriptions.manage',    'Subscribe, upgrade, and cancel subscriptions', 'subscriptions',  NOW()),
    ('payments.read',           'View payment history',                          'payments',       NOW()),
    ('roles.read',              'View roles and permissions',                    'roles',          NOW()),
    ('roles.manage',            'Create, edit, and delete custom roles',        'roles',          NOW())
ON CONFLICT (name) DO NOTHING;

-- Seed: test organization and admin user
-- Password: password123 (bcrypt hash)

DO $$
DECLARE
    v_org_uuid UUID;
    v_owner_role_id INT;
    v_admin_role_id INT;
    v_member_role_id INT;
BEGIN
    IF EXISTS (SELECT 1 FROM "user" WHERE email = 'admin@test.com' AND deleted_at IS NULL) THEN
        RAISE NOTICE 'Seed data already exists, skipping.';
        RETURN;
    END IF;

    -- create org
    INSERT INTO organizations (uuid, name, slug, email, phone, website, industry, plan, plan_id, is_active, created_at, updated_at)
    VALUES (
        gen_random_uuid(),
        'Test Organization',
        'test-organization',
        'contact@testorg.com',
        '+5511999999999',
        'https://testorg.com',
        'Technology',
        'free',
        (SELECT id FROM plans WHERE name = 'free'),
        true,
        NOW(),
        NOW()
    )
    RETURNING uuid INTO v_org_uuid;

    -- create system roles for test org
    INSERT INTO roles (uuid, organization_id, name, description, is_system, created_at, updated_at)
    VALUES (gen_random_uuid(), v_org_uuid, 'owner', 'Full access to all organization resources', true, NOW(), NOW())
    RETURNING id INTO v_owner_role_id;

    INSERT INTO roles (uuid, organization_id, name, description, is_system, created_at, updated_at)
    VALUES (gen_random_uuid(), v_org_uuid, 'admin', 'Administrative access (cannot delete organization)', true, NOW(), NOW())
    RETURNING id INTO v_admin_role_id;

    INSERT INTO roles (uuid, organization_id, name, description, is_system, created_at, updated_at)
    VALUES (gen_random_uuid(), v_org_uuid, 'member', 'Basic member access', true, NOW(), NOW())
    RETURNING id INTO v_member_role_id;

    -- owner gets ALL permissions
    INSERT INTO role_permissions (role_id, permission_id)
    SELECT v_owner_role_id, id FROM permissions;

    -- admin gets all except organizations.delete and organizations.manage
    INSERT INTO role_permissions (role_id, permission_id)
    SELECT v_admin_role_id, id FROM permissions
    WHERE name NOT IN ('organizations.delete', 'organizations.manage');

    -- member gets basic read + contacts create/update + chats
    INSERT INTO role_permissions (role_id, permission_id)
    SELECT v_member_role_id, id FROM permissions
    WHERE name IN (
        'members.list',
        'contacts.create', 'contacts.read', 'contacts.update',
        'chats.read',
        'organizations.read',
        'plans.read',
        'payments.read'
    );

    -- create free subscription for test org
    INSERT INTO subscriptions (uuid, organization_id, plan_id, status, cancel_at_period_end, created_at, updated_at)
    VALUES (
        gen_random_uuid(),
        v_org_uuid,
        (SELECT id FROM plans WHERE name = 'free'),
        'active',
        false,
        NOW(),
        NOW()
    );

    -- create test user with owner role
    INSERT INTO "user" (uuid, organization_id, name, email, password_hash, role, role_id, status, created_at, updated_at)
    VALUES (
        gen_random_uuid(),
        v_org_uuid,
        'Admin User',
        'admin@test.com',
        '$2a$14$JPJpuGo99NDZ4S9i92v3h.D88MgWtn9jQX1soI/mBwGKdPbXgZHH.',
        'admin',
        v_owner_role_id,
        'active',
        NOW(),
        NOW()
    );

    RAISE NOTICE 'Seed data created successfully.';
END $$;
