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

func (db *DB) List(ctx context.Context, userID string) ([]*models.Record, error) {
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
		ON records.id = datarecords.recordid;
	WHERE r.userid =  $1`

	rows, err := tx.Query(ctx, sql, userID)
	if err != nil {
		return nil, fmt.Errorf("an occured error while geting records, err: %w", err)
	}
	defer rows.Close()

	var rs []*models.Record
	var rids []string
	for rows.Next() {
		var r models.Record
		err := rows.Scan(
			&r.ID,
			&r.Owner,
			&r.Description,
			&r.Type,
			&r.Created,
			&r.Modified,
			&r.Hashsum,
			&r.Version,
			&r.Data)
		if err != nil {
			return nil, fmt.Errorf("an error occurred when filling in an array of records, err: %w", err)
		}
		rs = append(rs, &r)
		rids = append(rids, r.ID)
	}

	rmi, err := db.getRecordsMetainfos(ctx, tx, rids)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while retrieving an array metainfos of record, err: %w", err)
	}

	for _, r := range rs {
		if mi, ok := rmi[r.ID]; ok {
			r.Metainfo = mi
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf(tmpErrCommitTxErr(), err)
	}

	return rs, nil
}

func (db *DB) Get(ctx context.Context, userID string, recordID string) (*models.Record, error) {
	sql := `SELECT r.id, r.userid, r.description, r.dtype, r.created, r.modified, r.hashsum, r.version, dr.data
	FROM records as r
		LEFT JOIN datarecords as dr
		ON records.id = datarecords.recordid;
	WHERE r.userid =  $1 AND records.id = $2`

	var r models.Record

	row := db.pool.QueryRow(ctx, sql, userID, recordID)
	if err := row.Scan(&r.ID,
		&r.Owner,
		&r.Description,
		&r.Type,
		&r.Created,
		&r.Modified,
		&r.Hashsum,
		&r.Version,
		&r.Data); err != nil {
		return nil, fmt.Errorf("an occured error while getting record, err: %w", err)
	}

	return &r, nil
}

func (db *DB) Add(ctx context.Context, userID string, record *models.RecordDTO) (*models.Record, error) {
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

	sql := `INSERT INTO records(description, dtype, userid, created, modified, hashsum, version)
	VALUES ($1, $2, $3, $4, 1)
	ON CONFLICT (id) DO UPDATE records SET (description = $1, modified = CURRENT_TIMESTAMP, hashsum = $4)
	RETURNING 
		id, userid, description, dtype, created, modified, hashsum;`
	row := tx.QueryRow(ctx, sql, record.Description, record.Type, userID, record.Hashsum)
	if err := row.Scan(&r.ID, &r.Owner, &r.Description, &r.Type, &r.Created, &r.Modified, &r.Hashsum); err != nil {
		return nil, fmt.Errorf("an occured error while add record, err: %w", err)
	}

	sql = `INSERT INTO datarecords(recordid, data)
	VALUES ($1, $2)
	ON CONFLICT (recordid) DO UPDATE datarecords SET data = $2 WHERE recordid = $1;`
	if _, err := tx.Exec(ctx, sql, r.ID, record.Data); err != nil {
		return nil, fmt.Errorf("an occured error while add data for record, err: %w", err)
	}

	r.Data = record.Data

	if err := db.updateMetainfos(ctx, tx, r.ID, record.Metainfo); err != nil {
		return nil, fmt.Errorf("an occured error while add record metainfo, err: %w", err)
	}
	r.Metainfo = record.Metainfo

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf(tmpErrCommitTxErr(), err)
	}

	return &r, nil
}

func (db *DB) Update(ctx context.Context, userID string, record *models.Record) (*models.Record, error) {
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

	sql := `UPDATE records SET (description = $1, modified = CURRENT_TIMESTAMP, hashsum = $4, version = $5)
	RETURNING id, userid, description, dtype, created, modified, hashsum;`
	row := tx.QueryRow(ctx, sql, record.Description, record.Type, userID, record.Hashsum, record.Version)
	if err := row.Scan(&r.ID, &r.Owner, &r.Description, &r.Type, &r.Created, &r.Modified, &r.Hashsum); err != nil {
		return nil, fmt.Errorf("an occured error while update record, err: %w", err)
	}

	sql = `UPDATE datarecords SET data = $2 WHERE recordid = $1;`
	if _, err := tx.Exec(ctx, sql, r.ID, record.Data); err != nil {
		return nil, fmt.Errorf("an occured error while update data for record, err: %w", err)
	}

	if err := db.updateMetainfos(ctx, tx, r.ID, record.Metainfo); err != nil {
		return nil, fmt.Errorf("an occured error while update record metainfo, err: %w", err)
	}

	r.Data = record.Data
	r.Metainfo = record.Metainfo

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf(tmpErrCommitTxErr(), err)
	}

	return &r, nil
}

func (db *DB) Delete(ctx context.Context, userID string, recordID string) error {
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

	if err := db.cleanUpRecordMetainfo(ctx, tx, recordID); err != nil {
		return fmt.Errorf("an occured error while delete records metainfo, err: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf(tmpErrCommitTxErr(), err)
	}

	return nil
}

func (db *DB) getRecordsMetainfos(ctx context.Context,
	tx pgx.Tx, recordsID []string) (map[string][]*models.Metainfo, error) {
	sql := `SELECT recordid, key, value
	FROM metainfo
	WHERE recordid IN ($1);`
	rows, err := tx.Query(ctx, sql, recordsID)
	if err != nil {
		return nil, fmt.Errorf("an occured error while geting metaoinfos, err: %w", err)
	}
	defer rows.Close()

	rmi := make(map[string][]*models.Metainfo)

	for rows.Next() {
		recordID := ""
		var mi models.Metainfo

		if err := rows.Scan(&recordID, &mi.Key, &mi.Value); err != nil {
			return nil, fmt.Errorf("an error occurred when filling in an array of metainfos, err: %w", err)
		}

		rmi[recordID] = append(rmi[recordID], &mi)
	}

	return rmi, nil
}

func (db *DB) updateMetainfos(ctx context.Context, tx pgx.Tx, recordID string, mis []*models.Metainfo) error {
	if err := db.cleanUpRecordMetainfo(ctx, tx, recordID); err != nil {
		return fmt.Errorf("an occured error while do clean up metainfo, err: %w", err)
	}
	for _, mi := range mis {
		if err := db.updateMetainfo(ctx, tx, recordID, mi); err != nil {
			return fmt.Errorf("an error occured while update metainfo (recordID=%s), err: %w", recordID, err)
		}
	}
	return nil
}

func (db *DB) cleanUpRecordMetainfo(ctx context.Context, tx pgx.Tx, recordID string) error {
	sql := `DELETE FROM metainfo WHERE recordid = $1;`

	if _, err := tx.Exec(ctx, sql, recordID); err != nil {
		return fmt.Errorf("an occured error while cleaning up metainfo, err: %w", err)
	}

	return nil
}

func (db *DB) updateMetainfo(ctx context.Context, tx pgx.Tx, recordID string, mi *models.Metainfo) error {
	sql := `INSERT INTO metainfo(recordid, key, value)
	VALUES ($1, $2, $3)
	ON CONFLICT (recordid, key) DO UPDATE SET value = $3 WHERE recordid = $1 AND key = $2;`

	if _, err := tx.Exec(ctx, sql, recordID, mi.Key, mi.Value); err != nil {
		return fmt.Errorf("an occured error while do add or update metainfo, err: %w", err)
	}

	return nil
}
