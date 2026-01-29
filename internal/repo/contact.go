package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projeto-crm-2026/crm-services/internal/domain/entity"
)

type ContactFilters struct {
	Status, Type, Source         interface{}
	AssignedToID, CreatedByID    *uint
	City, State, Country, Search *string
	Tags                         []string
	CreatedAfter, CreatedBefore  *time.Time
}

type PaginatedResult[T any] struct {
	Data       []*T  `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

type ContactRepo interface {
	Create(ctx context.Context, contact *entity.Contact) (*entity.Contact, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Contact, error)
	GetByEmail(ctx context.Context, email string) (*entity.Contact, error)
	Update(ctx context.Context, contact *entity.Contact) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	List(ctx context.Context, filters ContactFilters) ([]*entity.Contact, error)
	ListPaginated(ctx context.Context, filters ContactFilters, page, pageSize int) (*PaginatedResult[entity.Contact], error)

	Search(ctx context.Context, query string, filters ContactFilters) ([]*entity.Contact, error)
}

type contactRepo struct {
	pool *pgxpool.Pool
}

func NewContactRepo(pool *pgxpool.Pool) ContactRepo {
	return &contactRepo{pool: pool}
}

const (
	fullFields = `id,type,first_name,last_name,full_name,email,phone,mobile_phone,alternate_email,
		company_name,job_title,department,street,number,complement,district,city,state,zip_code,
		country,status,source,tags,notes,assigned_to_id,created_by_id,updated_by_id,created_at,updated_at`

	listFields = `id,type,first_name,last_name,full_name,email,phone,mobile_phone,company_name,
		job_title,status,source,city,state,country,assigned_to_id,created_at,updated_at`
)

func (r *contactRepo) Create(ctx context.Context, c *entity.Contact) (*entity.Contact, error) {
	r.normalize(c)
	contact := &entity.Contact{}

	err := r.pool.QueryRow(ctx, `
		INSERT INTO contacts (type,first_name,last_name,full_name,email,phone,mobile_phone,
			alternate_email,company_name,job_title,department,street,number,complement,district,
			city,state,zip_code,country,status,source,tags,notes,assigned_to_id,created_by_id,
			created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,NOW(),NOW())
		RETURNING id,created_at,updated_at`,
		c.Type, c.FirstName, c.LastName, c.FullName,
		validate(c.Email), validate(c.Phone), validate(c.MobilePhone), validate(c.AlternateEmail),
		validate(c.CompanyName), validate(c.JobTitle), validate(c.Department),
		validate(c.Street), validate(c.Number), validate(c.Complement), validate(c.District),
		validate(c.City), validate(c.State), validate(c.ZipCode), validate(c.Country),
		c.Status, c.Source, c.Tags, validate(c.Notes), c.AssignedToID, c.CreatedByID,
	).Scan(r.scanFull(contact)...)

	return contact, err
}

func (r *contactRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Contact, error) {
	contact := &entity.Contact{}
	err := r.pool.QueryRow(ctx,
		fmt.Sprintf("SELECT %s FROM contacts WHERE id=$1 AND deleted_at IS NULL", fullFields), id,
	).Scan(r.scanFull(contact)...)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrContactNotFound
	}
	return contact, err
}

func (r *contactRepo) GetByEmail(ctx context.Context, email string) (*entity.Contact, error) {
	contact := &entity.Contact{}
	err := r.pool.QueryRow(ctx,
		fmt.Sprintf("SELECT %s FROM contacts WHERE email=$1 AND deleted_at IS NULL", fullFields), email,
	).Scan(r.scanFull(contact)...)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrContactNotFound
	}
	return contact, err
}

func (r *contactRepo) Update(ctx context.Context, c *entity.Contact) error {
	r.normalize(c)
	res, err := r.pool.Exec(ctx, `
		UPDATE contacts SET type=$1,first_name=$2,last_name=$3,full_name=$4,email=$5,phone=$6,
			mobile_phone=$7,alternate_email=$8,company_name=$9,job_title=$10,department=$11,
			street=$12,number=$13,complement=$14,district=$15,city=$16,state=$17,zip_code=$18,
			country=$19,status=$20,source=$21,tags=$22,notes=$23,assigned_to_id=$24,
			updated_by_id=$25,updated_at=NOW()
		WHERE id=$26 AND deleted_at IS NULL`,
		c.Type, c.FirstName, c.LastName, c.FullName,
		validate(c.Email), validate(c.Phone), validate(c.MobilePhone), validate(c.AlternateEmail),
		validate(c.CompanyName), validate(c.JobTitle), validate(c.Department),
		validate(c.Street), validate(c.Number), validate(c.Complement), validate(c.District),
		validate(c.City), validate(c.State), validate(c.ZipCode), validate(c.Country),
		c.Status, c.Source, c.Tags, validate(c.Notes), c.AssignedToID, c.UpdatedByID, c.ID,
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
	return r.exec(ctx, "DELETE FROM contacts WHERE id=$1", id)
}

func (r *contactRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.exec(ctx, "UPDATE contacts SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL", id)
}

func (r *contactRepo) List(ctx context.Context, f ContactFilters) ([]*entity.Contact, error) {
	q := r.buildQuery(listFields, f, "", 1000, 0)
	rows, err := r.pool.Query(ctx, q.sql, q.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanList(rows)
}

func (r *contactRepo) ListPaginated(ctx context.Context, f ContactFilters, page, size int) (*PaginatedResult[entity.Contact], error) {
	q := r.buildQuery("", f, "", 0, 0)

	var total int64
	if err := r.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM contacts WHERE %s", q.where), q.args...).Scan(&total); err != nil {
		return nil, err
	}

	q = r.buildQuery(listFields, f, "created_at DESC", size, (page-1)*size)
	rows, err := r.pool.Query(ctx, q.sql, q.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	contacts, err := r.scanList(rows)
	if err != nil {
		return nil, err
	}

	return &PaginatedResult[entity.Contact]{
		Data: contacts, Total: total, Page: page, PageSize: size,
		TotalPages: int(total+int64(size)-1) / size,
	}, nil
}

func (r *contactRepo) Search(ctx context.Context, term string, f ContactFilters) ([]*entity.Contact, error) {
	q := r.buildQuery(listFields, f, "", 100, 0)
	q.addSearch(term)
	q.order = "CASE WHEN full_name ILIKE $1 THEN 1 WHEN email ILIKE $1 THEN 2 WHEN company_name ILIKE $1 THEN 3 ELSE 4 END,full_name"

	rows, err := r.pool.Query(ctx, q.build(), q.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanList(rows)
}

type query struct {
	sql, fields, where, order string
	args                      []interface{}
	limit, offset             int
}

func (r *contactRepo) buildQuery(fields string, f ContactFilters, order string, limit, offset int) *query {
	q := &query{fields: fields, where: "deleted_at IS NULL", order: order, limit: limit, offset: offset}

	q.add("status=$%d", f.Status)
	q.add("type=$%d", f.Type)
	q.add("source=$%d", f.Source)
	q.add("assigned_to_id=$%d", f.AssignedToID)
	q.add("created_by_id=$%d", f.CreatedByID)
	q.add("city ILIKE $%d", wrap(f.City))
	q.add("state ILIKE $%d", wrap(f.State))
	q.add("country=$%d", f.Country)
	if len(f.Tags) > 0 {
		q.add("tags && $%d", f.Tags)
	}
	q.add("created_at>=$%d", f.CreatedAfter)
	q.add("created_at<=$%d", f.CreatedBefore)
	if f.Search != nil && *f.Search != "" {
		q.addSearch(*f.Search)
	}

	if fields != "" {
		q.sql = fmt.Sprintf("SELECT %s FROM contacts WHERE %s", fields, q.where)
	}
	return q
}

func (q *query) add(cond string, val interface{}) {
	q.where += fmt.Sprintf(" AND "+cond, len(q.args)+1)
	q.args = append(q.args, val)
}

func (q *query) addSearch(term string) {
	p := "%" + term + "%"
	pos := len(q.args) + 1
	q.where += fmt.Sprintf(" AND (full_name ILIKE $%d OR email ILIKE $%d OR phone ILIKE $%d OR mobile_phone ILIKE $%d OR company_name ILIKE $%d)",
		pos, pos, pos, pos, pos)
	q.args = append(q.args, p)
}

func (q *query) build() string {
	sql := q.sql
	if q.order != "" {
		sql += " ORDER BY " + q.order
	}
	if q.limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", q.limit)
	}
	if q.offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", q.offset)
	}
	return sql
}

func (r *contactRepo) exec(ctx context.Context, query string, args ...interface{}) error {
	res, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrContactNotFound
	}
	return nil
}

func (r *contactRepo) normalize(c *entity.Contact) {
	if c.FullName == "" {
		if c.Type == entity.ContactTypePerson {
			c.FullName = strings.TrimSpace(c.FirstName + " " + c.LastName)
		} else {
			c.FullName = c.CompanyName.String
		}
	}
}

func (r *contactRepo) scanFull(c *entity.Contact) []interface{} {
	return []interface{}{
		&c.ID, &c.Type, &c.FirstName, &c.LastName, &c.FullName,
		&c.Email, &c.Phone, &c.MobilePhone, &c.AlternateEmail,
		&c.CompanyName, &c.JobTitle, &c.Department,
		&c.Street, &c.Number, &c.Complement, &c.District,
		&c.City, &c.State, &c.ZipCode, &c.Country,
		&c.Status, &c.Source, &c.Tags, &c.Notes,
		&c.AssignedToID, &c.CreatedByID, &c.UpdatedByID,
		&c.CreatedAt, &c.UpdatedAt,
	}
}

func (r *contactRepo) scanListDest(c *entity.Contact) []interface{} {
	return []interface{}{
		&c.ID, &c.Type, &c.FirstName, &c.LastName, &c.FullName,
		&c.Email, &c.Phone, &c.MobilePhone, &c.CompanyName, &c.JobTitle,
		&c.Status, &c.Source, &c.City, &c.State, &c.Country,
		&c.AssignedToID, &c.CreatedAt, &c.UpdatedAt,
	}
}

func (r *contactRepo) scanList(rows pgx.Rows) ([]*entity.Contact, error) {
	var cs []*entity.Contact
	for rows.Next() {
		contact := &entity.Contact{}
		if err := rows.Scan(r.scanListDest(contact)...); err != nil {
			return nil, err
		}
		cs = append(cs, contact)
	}
	return cs, rows.Err()
}

func validate(ns sql.NullString) interface{} {
	if ns.Valid {
		return ns.String
	}
	return nil
}

func wrap(s *string) interface{} {
	if s == nil {
		return nil
	}
	return "%" + *s + "%"
}
