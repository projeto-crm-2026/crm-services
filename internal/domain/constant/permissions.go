package constant

const (
	PermMembersInvite     = "members.invite"
	PermMembersList       = "members.list"
	PermMembersManageRole = "members.manage_roles"
	PermMembersRemove     = "members.remove"
	PermMembersDeactivate = "members.deactivate"

	PermContactsCreate = "contacts.create"
	PermContactsRead   = "contacts.read"
	PermContactsUpdate = "contacts.update"
	PermContactsDelete = "contacts.delete"

	PermChatsRead = "chats.read"

	PermWebhooksCreate       = "webhooks.create"
	PermWebhooksRead         = "webhooks.read"
	PermWebhooksUpdate       = "webhooks.update"
	PermWebhooksDelete       = "webhooks.delete"
	PermWebhooksTokensCreate = "webhooks.tokens.create"
	PermWebhooksTokensRead   = "webhooks.tokens.read"
	PermWebhooksTokensDelete = "webhooks.tokens.delete"

	PermApiKeysCreate = "api_keys.create"
	PermApiKeysRead   = "api_keys.read"
	PermApiKeysDelete = "api_keys.delete"

	PermOrganizationsRead   = "organizations.read"
	PermOrganizationsUpdate = "organizations.update"
	PermOrganizationsDelete = "organizations.delete"
	PermOrganizationsManage = "organizations.manage"

	PermPlansRead = "plans.read"

	PermSubscriptionsManage = "subscriptions.manage"
	PermPaymentsRead        = "payments.read"

	PermRolesRead   = "roles.read"
	PermRolesManage = "roles.manage"
)

type PermissionDef struct {
	Name        string
	Description string
	Category    string
}

var PermissionCatalog = []PermissionDef{
	{PermMembersInvite, "Invite new members to the organization", "members"},
	{PermMembersList, "List organization members", "members"},
	{PermMembersManageRole, "Change a member's role", "members"},
	{PermMembersRemove, "Remove a member from the organization", "members"},
	{PermMembersDeactivate, "Deactivate or reactivate a member", "members"},

	{PermContactsCreate, "Create contacts", "contacts"},
	{PermContactsRead, "View contacts", "contacts"},
	{PermContactsUpdate, "Update contacts", "contacts"},
	{PermContactsDelete, "Delete contacts", "contacts"},

	{PermChatsRead, "View and access chats", "chats"},

	{PermWebhooksCreate, "Create outgoing webhooks", "webhooks"},
	{PermWebhooksRead, "View outgoing webhooks", "webhooks"},
	{PermWebhooksUpdate, "Update outgoing webhooks", "webhooks"},
	{PermWebhooksDelete, "Delete outgoing webhooks", "webhooks"},
	{PermWebhooksTokensCreate, "Create incoming webhook tokens", "webhooks"},
	{PermWebhooksTokensRead, "List incoming webhook tokens", "webhooks"},
	{PermWebhooksTokensDelete, "Delete incoming webhook tokens", "webhooks"},

	{PermApiKeysCreate, "Create API keys", "api_keys"},
	{PermApiKeysRead, "List API keys", "api_keys"},
	{PermApiKeysDelete, "Delete API keys", "api_keys"},

	{PermOrganizationsRead, "View organization details", "organizations"},
	{PermOrganizationsUpdate, "Update organization settings", "organizations"},
	{PermOrganizationsDelete, "Permanently delete the organization", "organizations"},
	{PermOrganizationsManage, "Restore and manage organization lifecycle", "organizations"},

	{PermPlansRead, "View plans and usage", "plans"},

	{PermSubscriptionsManage, "Subscribe, upgrade, and cancel subscriptions", "subscriptions"},
	{PermPaymentsRead, "View payment history", "payments"},

	{PermRolesRead, "View roles and permissions", "roles"},
	{PermRolesManage, "Create, edit, and delete custom roles", "roles"},
}

func AllPermissionNames() []string {
	names := make([]string, len(PermissionCatalog))
	for i, p := range PermissionCatalog {
		names[i] = p.Name
	}
	return names
}

var SystemRolePermissions = map[string][]string{
	"owner": AllPermissionNames(),
	"admin": {
		PermMembersInvite, PermMembersList, PermMembersManageRole, PermMembersRemove, PermMembersDeactivate,
		PermContactsCreate, PermContactsRead, PermContactsUpdate, PermContactsDelete,
		PermChatsRead,
		PermWebhooksCreate, PermWebhooksRead, PermWebhooksUpdate, PermWebhooksDelete,
		PermWebhooksTokensCreate, PermWebhooksTokensRead, PermWebhooksTokensDelete,
		PermApiKeysCreate, PermApiKeysRead, PermApiKeysDelete,
		PermOrganizationsRead, PermOrganizationsUpdate,
		PermPlansRead,
		PermSubscriptionsManage, PermPaymentsRead,
		PermRolesRead, PermRolesManage,
	},
	"member": {
		PermMembersList,
		PermContactsCreate, PermContactsRead, PermContactsUpdate,
		PermChatsRead,
		PermOrganizationsRead,
		PermPlansRead,
		PermPaymentsRead,
	},
}
