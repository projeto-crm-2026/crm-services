package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
	"github.com/projeto-crm-2026/crm-services/pkg/sqlutils"
	"github.com/projeto-crm-2026/crm-services/pkg/utils"
)

var ErrContactNotFound = errors.New("contact not found")

type ContactFilters struct {
	OrganizationID               uuid.UUID
	Status, Type, Source         any
	AssignedToID, CreatedByID    *uint
	City, State, Country, Search *string
	Tags                         []string
	CreatedAfter, CreatedBefore  *time.Time
}

type ContactRepo interface {
	Create(ctx context.Context, contact *entity.Contact) (*entity.Contact, error)
	GetByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*entity.Contact, error)
	GetByEmail(ctx context.Context, email string, organizationID uuid.UUID) (*entity.Contact, error)
	Update(ctx context.Context, contact *entity.Contact) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	List(ctx context.Context, filters ContactFilters) ([]*entity.Contact, error)
	ListPaginated(ctx context.Context, filters ContactFilters, page, pageSize int) (*sqlutils.PaginatedResult[entity.Contact], error)

	Search(ctx context.Context, query string, filters ContactFilters) ([]*entity.Contact, error)
}

type contactRepo struct {
	pool *pgxpool.Pool
}

func NewContactRepo(pool *pgxpool.Pool) ContactRepo {
	return &contactRepo{pool: pool}
}

const (
	fullFields = `id,uuid,organization_id,type,first_name,last_name,full_name,email,phone,mobile_phone,alternate_email,
		company_name,job_title,department,street,number,complement,district,city,state,zip_code,
		country,status,source,tags,notes,assigned_to_id,created_by_id,updated_by_id,created_at,updated_at`

	listFields = `id,uuid,organization_id,type,first_name,last_name,full_name,email,phone,mobile_phone,company_name,
		job_title,status,source,city,state,country,assigned_to_id,created_at,updated_at`
)

func (r *contactRepo) Create(ctx context.Context, c *entity.Contact) (*entity.Contact, error) {
	normalizeContact(c)
	contact := &entity.Contact{}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO contacts (type,organization_id,first_name,last_name,full_name,email,phone,mobile_phone,
			alternate_email,company_name,job_title,department,street,number,complement,district,
			city,state,zip_code,country,status,source,tags,notes,assigned_to_id,created_by_id,
			created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,NOW(),NOW())
		RETURNING `+fullFields,
		c.Type, c.OrganizationID, c.FirstName, c.LastName, c.FullName,
		utils.NullString(c.Email), utils.NullString(c.Phone), utils.NullString(c.MobilePhone), utils.NullString(c.AlternateEmail),
		utils.NullString(c.CompanyName), utils.NullString(c.JobTitle), utils.NullString(c.Department),
		utils.NullString(c.Street), utils.NullString(c.Number), utils.NullString(c.Complement), utils.NullString(c.District),
		utils.NullString(c.City), utils.NullString(c.State), utils.NullString(c.ZipCode), utils.NullString(c.Country),
		c.Status, c.Source, c.Tags, utils.NullString(c.Notes), c.AssignedToID, c.CreatedByID,
	).Scan(scanContactFull(contact)...)

	return contact, err
}

func (r *contactRepo) GetByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*entity.Contact, error) {
	contact := &entity.Contact{}
	err := r.pool.QueryRow(ctx,
		fmt.Sprintf("SELECT %s FROM contacts WHERE uuid=$1 AND organization_id=$2 AND deleted_at IS NULL", fullFields),
		id, organizationID,
	).Scan(scanContactFull(contact)...)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrContactNotFound
	}
	return contact, err
}

func (r *contactRepo) GetByEmail(ctx context.Context, email string, organizationID uuid.UUID) (*entity.Contact, error) {
	contact := &entity.Contact{}
	err := r.pool.QueryRow(ctx,
		fmt.Sprintf("SELECT %s FROM contacts WHERE email=$1 AND organization_id=$2 AND deleted_at IS NULL", fullFields),
		email, organizationID,
	).Scan(scanContactFull(contact)...)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrContactNotFound
	}
	return contact, err
}

func (r *contactRepo) Update(ctx context.Context, c *entity.Contact) error {
	normalizeContact(c)
	res, err := r.pool.Exec(ctx, `
		UPDATE contacts SET type=$1,first_name=$2,last_name=$3,full_name=$4,email=$5,phone=$6,
			mobile_phone=$7,alternate_email=$8,company_name=$9,job_title=$10,department=$11,
			street=$12,number=$13,complement=$14,district=$15,city=$16,state=$17,zip_code=$18,
			country=$19,status=$20,source=$21,tags=$22,notes=$23,assigned_to_id=$24,
			updated_by_id=$25,updated_at=NOW()
		WHERE uuid=$26 AND deleted_at IS NULL`,
		c.Type, c.FirstName, c.LastName, c.FullName,
		utils.NullString(c.Email), utils.NullString(c.Phone), utils.NullString(c.MobilePhone), utils.NullString(c.AlternateEmail),
		utils.NullString(c.CompanyName), utils.NullString(c.JobTitle), utils.NullString(c.Department),
		utils.NullString(c.Street), utils.NullString(c.Number), utils.NullString(c.Complement), utils.NullString(c.District),
		utils.NullString(c.City), utils.NullString(c.State), utils.NullString(c.ZipCode), utils.NullString(c.Country),
		c.Status, c.Source, c.Tags, utils.NullString(c.Notes), c.AssignedToID, c.UpdatedByID, c.UUID,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrContactNotFound
	}
	return nil
}

func (r *contactRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.pool.Exec(ctx, "DELETE FROM contacts WHERE uuid=$1", id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrContactNotFound
	}
	return nil
}

func (r *contactRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	res, err := r.pool.Exec(ctx, "UPDATE contacts SET deleted_at=NOW() WHERE uuid=$1 AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrContactNotFound
	}
	return nil
}

func (r *contactRepo) List(ctx context.Context, f ContactFilters) ([]*entity.Contact, error) {
	q := f.buildQuery(listFields, "", 1000, 0)
	rows, err := r.pool.Query(ctx, q.Build(), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanContactRows(rows)
}

func (r *contactRepo) ListPaginated(ctx context.Context, f ContactFilters, page, size int) (*sqlutils.PaginatedResult[entity.Contact], error) {
	countQ := f.buildQuery("", "", 0, 0)
	var total int64
	if err := r.pool.QueryRow(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM contacts WHERE %s", countQ.Where()),
		countQ.Args()...,
	).Scan(&total); err != nil {
		return nil, err
	}

	q := f.buildQuery(listFields, "created_at DESC", size, (page-1)*size)
	rows, err := r.pool.Query(ctx, q.Build(), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	contacts, err := scanContactRows(rows)
	if err != nil {
		return nil, err
	}

	return &sqlutils.PaginatedResult[entity.Contact]{
		Data:       contacts,
		Total:      total,
		Page:       page,
		PageSize:   size,
		TotalPages: int(total+int64(size)-1) / size,
	}, nil
}

func (r *contactRepo) Search(ctx context.Context, term string, f ContactFilters) ([]*entity.Contact, error) {
	q := f.buildQuery(listFields, "", 100, 0)
	pos := q.AddSearch([]string{"full_name", "email", "phone", "mobile_phone", "company_name"}, term)
	q.WithOrder(fmt.Sprintf(
		"CASE WHEN full_name ILIKE $%d THEN 1 WHEN email ILIKE $%d THEN 2 WHEN company_name ILIKE $%d THEN 3 ELSE 4 END,full_name",
		pos, pos, pos,
	))

	rows, err := r.pool.Query(ctx, q.Build(), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanContactRows(rows)
}

func (f ContactFilters) buildQuery(fields, order string, limit, offset int) *sqlutils.QueryBuilder {
	q := sqlutils.NewQueryBuilder("contacts").
		WithFields(fields).
		WithOrder(order).
		WithLimit(limit).
		WithOffset(offset)
	if f.OrganizationID != (uuid.UUID{}) {
		q.Add("organization_id=$%d", f.OrganizationID)
	}
	q.Add("status=$%d", f.Status)
	q.Add("type=$%d", f.Type)
	q.Add("source=$%d", f.Source)
	q.Add("assigned_to_id=$%d", f.AssignedToID)
	q.Add("created_by_id=$%d", f.CreatedByID)
	q.Add("city ILIKE $%d", sqlutils.WrapLike(f.City))
	q.Add("state ILIKE $%d", sqlutils.WrapLike(f.State))
	q.Add("country=$%d", f.Country)
	if len(f.Tags) > 0 {
		q.Add("tags && $%d", f.Tags)
	}
	q.Add("created_at>=$%d", f.CreatedAfter)
	q.Add("created_at<=$%d", f.CreatedBefore)
	if f.Search != nil && *f.Search != "" {
		q.AddSearch([]string{"full_name", "email", "phone", "mobile_phone", "company_name"}, *f.Search)
	}
	return q
}

func normalizeContact(c *entity.Contact) {
	if c.FullName == "" {
		if c.Type == entity.ContactTypePerson {
			c.FullName = strings.TrimSpace(c.FirstName + " " + c.LastName)
		} else {
			c.FullName = c.CompanyName.String
		}
	}
}

func scanContactFull(c *entity.Contact) []any {
	return []any{
		&c.ID, &c.UUID, &c.OrganizationID, &c.Type, &c.FirstName, &c.LastName, &c.FullName,
		&c.Email, &c.Phone, &c.MobilePhone, &c.AlternateEmail,
		&c.CompanyName, &c.JobTitle, &c.Department,
		&c.Street, &c.Number, &c.Complement, &c.District,
		&c.City, &c.State, &c.ZipCode, &c.Country,
		&c.Status, &c.Source, &c.Tags, &c.Notes,
		&c.AssignedToID, &c.CreatedByID, &c.UpdatedByID,
		&c.CreatedAt, &c.UpdatedAt,
	}
}

func scanContactListDest(c *entity.Contact) []any {
	return []any{
		&c.ID, &c.UUID, &c.OrganizationID, &c.Type, &c.FirstName, &c.LastName, &c.FullName,
		&c.Email, &c.Phone, &c.MobilePhone, &c.CompanyName, &c.JobTitle,
		&c.Status, &c.Source, &c.City, &c.State, &c.Country,
		&c.AssignedToID, &c.CreatedAt, &c.UpdatedAt,
	}
}

func scanContactRows(rows pgx.Rows) ([]*entity.Contact, error) {
	var cs []*entity.Contact
	for rows.Next() {
		contact := &entity.Contact{}
		if err := rows.Scan(scanContactListDest(contact)...); err != nil {
			return nil, err
		}
		cs = append(cs, contact)
	}
	return cs, rows.Err()
}
