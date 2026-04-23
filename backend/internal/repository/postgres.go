package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"hublead-attio-integration/backend/internal/domain"
	"hublead-attio-integration/backend/internal/service"
)

type PostgresStore struct {
	db *pgxpool.Pool
}

func NewPostgresStore(db *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *PostgresStore) LookupCompany(ctx context.Context, normalizedLinkedInURL string, normalizedDomain string) (domain.CompanySummary, bool, error) {
	const query = `
		select company_name, normalized_linkedin_url, coalesce(normalized_domain, ''),
		       attio_record_id, attio_web_url, first_synced_at, last_synced_at
		from companies
		where normalized_linkedin_url = $1
		   or ($2 <> '' and normalized_domain = $2)
		order by case when normalized_linkedin_url = $1 then 0 else 1 end
		limit 1`

	var company domain.CompanySummary
	err := s.db.QueryRow(ctx, query, normalizedLinkedInURL, normalizedDomain).Scan(
		&company.Name,
		&company.LinkedInURL,
		&company.Domain,
		&company.AttioRecordID,
		&company.AttioWebURL,
		&company.FirstSyncedAt,
		&company.LastSyncedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.CompanySummary{}, false, nil
	}
	if err != nil {
		return domain.CompanySummary{}, false, err
	}
	return company, true, nil
}

func (s *PostgresStore) UpsertCompany(ctx context.Context, input service.UpsertCompanyInput) (domain.CompanySummary, error) {
	const query = `
		with updated as (
			update companies set
				normalized_linkedin_url = $1,
				normalized_domain = nullif($2, ''),
				company_name = $3,
				attio_record_id = $4,
				attio_web_url = $5,
				raw_extracted_json = $6,
				last_synced_at = now(),
				last_sync_status = 'synced',
				last_sync_error = null
			where normalized_linkedin_url = $1
			   or ($2 <> '' and normalized_domain = $2)
			returning company_name, normalized_linkedin_url, coalesce(normalized_domain, '') as normalized_domain,
			          attio_record_id, attio_web_url, first_synced_at, last_synced_at
		), inserted as (
			insert into companies (
				normalized_linkedin_url,
				normalized_domain,
				company_name,
				attio_record_id,
				attio_web_url,
				raw_extracted_json,
				first_synced_at,
				last_synced_at,
				last_sync_status,
				last_sync_error
			)
			select $1, nullif($2, ''), $3, $4, $5, $6, now(), now(), 'synced', null
			where not exists (select 1 from updated)
			on conflict (normalized_linkedin_url) do update set
			normalized_domain = excluded.normalized_domain,
			company_name = excluded.company_name,
			attio_record_id = excluded.attio_record_id,
			attio_web_url = excluded.attio_web_url,
			raw_extracted_json = excluded.raw_extracted_json,
			last_synced_at = now(),
			last_sync_status = 'synced',
			last_sync_error = null
			returning company_name, normalized_linkedin_url, coalesce(normalized_domain, '') as normalized_domain,
			          attio_record_id, attio_web_url, first_synced_at, last_synced_at
		)
		select * from updated
		union all
		select * from inserted
		limit 1`

	var company domain.CompanySummary
	err := s.db.QueryRow(ctx, query,
		input.NormalizedLinkedInURL,
		input.NormalizedDomain,
		input.CompanyName,
		input.AttioRecordID,
		input.AttioWebURL,
		input.RawExtractedJSON,
	).Scan(
		&company.Name,
		&company.LinkedInURL,
		&company.Domain,
		&company.AttioRecordID,
		&company.AttioWebURL,
		&company.FirstSyncedAt,
		&company.LastSyncedAt,
	)
	return company, err
}
