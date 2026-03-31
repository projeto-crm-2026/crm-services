package migration

import (
	"fmt"

	"gorm.io/gorm"
)

func RunCustomMigrations(db *gorm.DB) error {
	if err := db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_webhook_events_gin
        ON webhook USING GIN (events);
    `).Error; err != nil {
		return fmt.Errorf("failed to create webhook events index: %w", err)
	}

	if err := db.Exec(`
		UPDATE organizations
		SET plan_id = (SELECT id FROM plans WHERE name = organizations.plan)
		WHERE plan_id IS NULL AND plan IS NOT NULL
		AND EXISTS (SELECT 1 FROM plans WHERE name = organizations.plan);
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill organization plan_id: %w", err)
	}

	if err := backfillSystemRoles(db); err != nil {
		return err
	}

	if err := backfillRolePermissions(db); err != nil {
		return err
	}

	if err := backfillFreeSubscriptions(db); err != nil {
		return err
	}

	return nil
}

func backfillSystemRoles(db *gorm.DB) error {
	if err := db.Exec(`
		DO $$
		DECLARE
			org RECORD;
			v_owner_id INT;
			v_admin_id INT;
			v_member_id INT;
		BEGIN
			FOR org IN
				SELECT uuid FROM organizations
				WHERE deleted_at IS NULL
				AND uuid NOT IN (SELECT DISTINCT organization_id FROM roles WHERE deleted_at IS NULL)
			LOOP
				INSERT INTO roles (uuid, organization_id, name, description, is_system, created_at, updated_at)
				VALUES (gen_random_uuid(), org.uuid, 'owner', 'Full access to all organization resources', true, NOW(), NOW())
				RETURNING id INTO v_owner_id;

				INSERT INTO roles (uuid, organization_id, name, description, is_system, created_at, updated_at)
				VALUES (gen_random_uuid(), org.uuid, 'admin', 'Administrative access (cannot delete organization)', true, NOW(), NOW())
				RETURNING id INTO v_admin_id;

				INSERT INTO roles (uuid, organization_id, name, description, is_system, created_at, updated_at)
				VALUES (gen_random_uuid(), org.uuid, 'member', 'Basic member access', true, NOW(), NOW())
				RETURNING id INTO v_member_id;

				INSERT INTO role_permissions (role_id, permission_id) SELECT v_owner_id, id FROM permissions;
				INSERT INTO role_permissions (role_id, permission_id) SELECT v_admin_id, id FROM permissions WHERE name NOT IN ('organizations.delete', 'organizations.manage');
				INSERT INTO role_permissions (role_id, permission_id) SELECT v_member_id, id FROM permissions WHERE name IN ('members.list','contacts.create','contacts.read','contacts.update','chats.read','organizations.read','plans.read','payments.read');

				UPDATE "user" SET role_id = v_owner_id
				WHERE organization_id = org.uuid AND role = 'admin' AND role_id IS NULL AND deleted_at IS NULL
				AND id = (SELECT id FROM "user" WHERE organization_id = org.uuid AND role = 'admin' AND deleted_at IS NULL ORDER BY id LIMIT 1);

				UPDATE "user" SET role_id = v_admin_id
				WHERE organization_id = org.uuid AND role = 'admin' AND role_id IS NULL AND deleted_at IS NULL;

				UPDATE "user" SET role_id = v_member_id
				WHERE organization_id = org.uuid AND role = 'member' AND role_id IS NULL AND deleted_at IS NULL;
			END LOOP;
		END $$;
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill system roles: %w", err)
	}
	return nil
}

func backfillRolePermissions(db *gorm.DB) error {
	if err := db.Exec(`
		DO $$
		DECLARE
			r RECORD;
		BEGIN
			FOR r IN
				SELECT id, name FROM roles
				WHERE is_system = true AND deleted_at IS NULL
				AND id NOT IN (SELECT DISTINCT role_id FROM role_permissions)
			LOOP
				IF r.name = 'owner' THEN
					INSERT INTO role_permissions (role_id, permission_id)
					SELECT r.id, id FROM permissions
					ON CONFLICT DO NOTHING;
				ELSIF r.name = 'admin' THEN
					INSERT INTO role_permissions (role_id, permission_id)
					SELECT r.id, id FROM permissions
					WHERE name NOT IN ('organizations.delete', 'organizations.manage')
					ON CONFLICT DO NOTHING;
				ELSIF r.name = 'member' THEN
					INSERT INTO role_permissions (role_id, permission_id)
					SELECT r.id, id FROM permissions
					WHERE name IN ('members.list','contacts.create','contacts.read','contacts.update','chats.read','organizations.read','plans.read','payments.read')
					ON CONFLICT DO NOTHING;
				END IF;
			END LOOP;
		END $$;
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill role permissions: %w", err)
	}
	return nil
}

func backfillFreeSubscriptions(db *gorm.DB) error {
	if err := db.Exec(`
		INSERT INTO subscriptions (uuid, organization_id, plan_id, status, cancel_at_period_end, created_at, updated_at)
		SELECT gen_random_uuid(), o.uuid, p.id, 'active', false, NOW(), NOW()
		FROM organizations o
		CROSS JOIN plans p
		WHERE p.name = 'free'
		AND o.deleted_at IS NULL
		AND o.uuid NOT IN (SELECT organization_id FROM subscriptions WHERE deleted_at IS NULL);
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill free subscriptions: %w", err)
	}
	return nil
}
