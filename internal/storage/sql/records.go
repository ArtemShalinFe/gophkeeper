package sql

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/ArtemShalinFe/gophkeeper/internal/models"
)

func tmpErrBeginTxErr() string {
	return "an occured error while trying start tx"
}

func tmpErrRollbackTxErr() string {
	return "an occured error while trying rollback tx"
}

func tmpErrCommitTxErr() string {
	return "an occured error while trying commit tx"
}

// ListRecords - used to retrieving user records.
func (db *DB) ListRecords(ctx context.Context, userID string, offset int, limit int) ([]*models.Record, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf(tmpErrBeginTxErr(), err)
	}

	defer func(tx pgx.Tx) {
		if err := tx.Rollback(ctx); err != nil {
			if !errors.Is(err, pgx.ErrTxClosed) {
				db.log.Error(tmpErrRollbackTxErr(), zap.Error(err))
			}
		}
	}(tx)

	sql := `SELECT r.id, r.userid, r.description, r.dtype, r.created, r.modified, r.hashsum, r.version, dr.data
	FROM records as r
		LEFT JOIN datarecords as dr
		ON r.id = dr.recordid
	WHERE userid = $1
	LIMIT $2 
	OFFSET $3`

	rows, err := tx.Query(ctx, sql, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("an occured error while geting records, err: %w", err)
	}
	defer rows.Close()

	var rs []*models.Record
	var rids []string
	for rows.Next() {
		var r models.Record
		err := rows.Scan(&r.ID, &r.Owner, &r.Description, &r.Type, &r.Created, &r.Modified,
			&r.Hashsum, &r.Version, &r.Data)
		if err != nil {
			return nil, fmt.Errorf("an error occurred when filling in an array of records, err: %w", err)
		}
		rs = append(rs, &r)
		rids = append(rids, r.ID)
	}

	rmi, err := db.getRecordsMetadatas(ctx, tx, rids)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while retrieving an array metadatas of record, err: %w", err)
	}

	for _, r := range rs {
		if mi, ok := rmi[r.ID]; ok {
			r.Metadata = mi
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf(tmpErrCommitTxErr(), err)
	}

	return rs, nil
}

// GetRecord - used to retrieving record.
func (db *DB) GetRecord(ctx context.Context, userID string, recordID string) (*models.Record, error) {
	sql := `SELECT r.id, r.userid, r.description, r.dtype, r.created, r.modified, r.hashsum, r.version, dr.data
	FROM records as r
		LEFT JOIN datarecords as dr
		ON r.id = dr.recordid
	WHERE r.userid =  $1 AND r.id = $2`

	var r models.Record

	row := db.pool.QueryRow(ctx, sql, userID, recordID)
	if err := row.Scan(&r.ID,
		&r.Owner, &r.Description, &r.Type, &r.Created, &r.Modified, &r.Hashsum, &r.Version, &r.Data); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrRecordNotFound
		}
		return nil, fmt.Errorf("an occured error while getting record, err: %w", err)
	}

	return &r, nil
}

// AddRecord - add new record to the storage.
func (db *DB) AddRecord(ctx context.Context, userID string, record *models.RecordDTO) (*models.Record, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf(tmpErrBeginTxErr(), err)
	}

	defer func(tx pgx.Tx) {
		if err := tx.Rollback(ctx); err != nil {
			if !errors.Is(err, pgx.ErrTxClosed) {
				db.log.Error(tmpErrRollbackTxErr(), zap.Error(err))
			}
		}
	}(tx)

	var r models.Record

	sql := `INSERT INTO records(description, dtype, userid, hashsum, version)
	VALUES ($1, $2, $3, $4, 1)
	ON CONFLICT (id) DO UPDATE 
		SET (description = $1, 
			modified = CURRENT_TIMESTAMP, 
			hashsum = $4, 
			version = EXCLUDED.version + 1)
	RETURNING 
		id, userid, description, dtype, created, modified, hashsum, version;`
	row := tx.QueryRow(ctx, sql, record.Description, record.Type, userID, record.Hashsum)
	if err := row.Scan(&r.ID, &r.Owner, &r.Description, &r.Type,
		&r.Created, &r.Modified, &r.Hashsum, &r.Version); err != nil {
		return nil, fmt.Errorf("an occured error while add record, err: %w", err)
	}

	if err := db.addDataRecord(ctx, tx, r.ID, record.Data); err != nil {
		return nil, fmt.Errorf("an occured error while add data for record, err: %w", err)
	}

	r.Data = record.Data

	if err := db.updateMetadatas(ctx, tx, r.ID, record.Metadata); err != nil {
		return nil, fmt.Errorf("an occured error while add record metadata, err: %w", err)
	}
	r.Metadata = record.Metadata

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf(tmpErrCommitTxErr(), err)
	}

	return &r, nil
}

// UpdateRecord - update record to the storage.
func (db *DB) UpdateRecord(ctx context.Context, userID string, record *models.Record) (*models.Record, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf(tmpErrBeginTxErr(), err)
	}

	defer func(tx pgx.Tx) {
		if err := tx.Rollback(ctx); err != nil {
			if !errors.Is(err, pgx.ErrTxClosed) {
				db.log.Error(tmpErrRollbackTxErr(), zap.Error(err))
			}
		}
	}(tx)

	var r models.Record

	sql := `INSERT INTO records(id, description, dtype, userid, hashsum, version) VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (id) DO UPDATE SET description = $2, modified = CURRENT_TIMESTAMP, hashsum = $5, version = $6
	RETURNING 
		id, userid, description, dtype, created, modified, hashsum, version;`
	row := tx.QueryRow(ctx, sql, record.ID, record.Description, record.Type, userID, record.Hashsum, record.Version)
	if err := row.Scan(
		&r.ID, &r.Owner, &r.Description, &r.Type, &r.Created, &r.Modified, &r.Hashsum, &r.Version); err != nil {
		return nil, fmt.Errorf("an occured error while update record, err: %w", err)
	}

	if err := db.addDataRecord(ctx, tx, r.ID, record.Data); err != nil {
		return nil, fmt.Errorf("an occured error while update data for record, err: %w", err)
	}

	if err := db.updateMetadatas(ctx, tx, r.ID, record.Metadata); err != nil {
		return nil, fmt.Errorf("an occured error while update record metadata, err: %w", err)
	}

	r.Data = record.Data
	r.Metadata = record.Metadata

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf(tmpErrCommitTxErr(), err)
	}

	return &r, nil
}

// DeleteRecord - used to delete records as deleted.
func (db *DB) DeleteRecord(ctx context.Context, userID string, recordID string) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf(tmpErrBeginTxErr(), err)
	}

	defer func(tx pgx.Tx) {
		if err := tx.Rollback(ctx); err != nil {
			if !errors.Is(err, pgx.ErrTxClosed) {
				db.log.Error(tmpErrRollbackTxErr(), zap.Error(err))
			}
		}
	}(tx)

	sql := `DELETE FROM datarecords WHERE userid = $1 and recordid = $2`
	if _, err := tx.Exec(ctx, sql, userID, recordID); err != nil {
		return fmt.Errorf("an occured error while delete data for record, err: %w", err)
	}

	sql = `DELETE FROM records WHERE userid = $1 and recordid = $2`
	if _, err := tx.Exec(ctx, sql, userID, recordID); err != nil {
		return fmt.Errorf("an occured error while delete record, err: %w", err)
	}

	if err := db.cleanUpRecordMetadata(ctx, tx, recordID); err != nil {
		return fmt.Errorf("an occured error while delete records metadata, err: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf(tmpErrCommitTxErr(), err)
	}

	return nil
}

func (db *DB) addDataRecord(ctx context.Context, tx pgx.Tx, recordID string, recordData []byte) error {
	sql := `INSERT INTO datarecords(recordid, data)
	VALUES ($1, $2)
	ON CONFLICT (recordid) DO UPDATE SET data = $2`
	if _, err := tx.Exec(ctx, sql, recordID, recordData); err != nil {
		return fmt.Errorf("add or update data for record was failed, err: %w", err)
	}

	return nil
}

func (db *DB) getRecordsMetadatas(ctx context.Context,
	tx pgx.Tx, recordsID []string) (map[string][]*models.Metadata, error) {
	sql := `SELECT recordid, key, value
	FROM metadata
	WHERE recordid = any ($1);`
	rows, err := tx.Query(ctx, sql, recordsID)
	if err != nil {
		return nil, fmt.Errorf("an occured error while geting metaoinfos, err: %w", err)
	}
	defer rows.Close()

	rmi := make(map[string][]*models.Metadata)

	for rows.Next() {
		recordID := ""
		var mi models.Metadata

		if err := rows.Scan(&recordID, &mi.Key, &mi.Value); err != nil {
			return nil, fmt.Errorf("an error occurred when filling in an array of metadatas, err: %w", err)
		}

		rmi[recordID] = append(rmi[recordID], &mi)
	}

	return rmi, nil
}

func (db *DB) updateMetadatas(ctx context.Context, tx pgx.Tx, recordID string, mis []*models.Metadata) error {
	if err := db.cleanUpRecordMetadata(ctx, tx, recordID); err != nil {
		return fmt.Errorf("an occured error while do clean up metadata, err: %w", err)
	}
	for _, mi := range mis {
		if err := db.addMetadata(ctx, tx, recordID, mi); err != nil {
			return fmt.Errorf("an error occured while update metadata (recordID=%s), err: %w", recordID, err)
		}
	}
	return nil
}

func (db *DB) cleanUpRecordMetadata(ctx context.Context, tx pgx.Tx, recordID string) error {
	sql := `DELETE FROM metadata WHERE recordid = $1;`

	if _, err := tx.Exec(ctx, sql, recordID); err != nil {
		return fmt.Errorf("an occured error while cleaning up metadata, err: %w", err)
	}

	return nil
}

func (db *DB) addMetadata(ctx context.Context, tx pgx.Tx, recordID string, mi *models.Metadata) error {
	sql := `INSERT INTO metadata(recordid, key, value) VALUES ($1, $2, $3);`

	if _, err := tx.Exec(ctx, sql, recordID, mi.Key, mi.Value); err != nil {
		return fmt.Errorf("an occured error while do add or update metadata, err: %w", err)
	}

	return nil
}
