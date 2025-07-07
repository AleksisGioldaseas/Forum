package database

import (
	"database/sql"
	"fmt"
	"forum/common/custom_errs"
)

func (db *DataBase) getAdminIds(tx *sql.Tx) ([]int64, error) {

	adminIds := []int64{}

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	rows, err := exor.QueryContext(ctx, GET_ADMIN_IDS)

	if err != nil {
		return adminIds, err
	}
	defer rows.Close()

	for rows.Next() {

		var id int64

		err = rows.Scan(&id)
		if err != nil {
			return adminIds, err
		}

		adminIds = append(adminIds, id)
	}

	err = rows.Err()
	if err != nil {
		return adminIds, err
	}

	if len(adminIds) == 0 {
		return adminIds, fmt.Errorf("we have no admins... that's no good %w", custom_errs.ErrNoRowsFound)
	}

	return adminIds, nil
}
