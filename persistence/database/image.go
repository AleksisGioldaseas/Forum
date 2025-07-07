package database

import (
	"database/sql"
	"fmt"
	"log"
)

func (db *DataBase) AddImage(tx *sql.Tx, fileName string) error {
	errMsg := "add image: %w"
	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	r, err := exor.ExecContext(ctx, INSERT_IMAGE, fileName)

	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if err = CatchNoRowsErr(r); err != nil {
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

func (db *DataBase) IsImageHidden(fileName string) bool {
	var hide int
	if _, ok := (*db.systemImages)[fileName]; ok {
		return false
	}

	err := db.conn.QueryRow(IS_IMAGE_HIDDEN, fileName).Scan(&hide)
	if err != nil {
		log.Println("isImageHidden:", fileName, err)
		return true
	}

	if hide == 1 {
		return true
	}
	return false
}

func (db *DataBase) toggleHideImage(tx *sql.Tx, table string, rowId int64, status int) error {
	errMsg := "toggle hide img: %w"

	imgColName, err := returnImgColName(table)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	query := fmt.Sprintf(TOGGLE_HIDE_IMAGE, imgColName, table)

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	_, err = exor.ExecContext(ctx, query, status, rowId)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	// Skipping catch no rows err because if there are no rows the content doesnt have an image
	return nil
}
