package database

import (
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"log"
	"slices"
	"sync"
)

var CategoryNameToIdMap = make(map[string]int64)
var CategoryIdToNameMap = make([]string, 20000)
var CategoryNames = []string{}
var CategoriesMutex = sync.Mutex{}

// Creates category or returns a category id if category name exists
func (db *DataBase) AddCategory(name string) (catId int64, err error) {
	if len(CategoryNames) == 0 {
		_, err := db.GetAllCategories()
		if err != nil {
			return 0, fmt.Errorf("get categories from add category when category vars empty %w", err)
		}
	}

	errorMsg := "AddCategory %w"

	_, ok := CategoryNameToIdMap[name]
	if ok {
		return 0, fmt.Errorf(errorMsg, "category already exists")
	}

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()

	err = exor.QueryRowContext(ctx, CREATE_CAT_OR_RETURN_ID, name).Scan(&catId)
	if err != nil {
		log.Println(custom_errs.ErrCreatingCategory, err)
		return 0, fmt.Errorf(errorMsg, err)
	}

	// if db.UseCache {
	// 	go db.cache.AddCategory(name, catId)
	// }

	CategoriesMutex.Lock()
	CategoryNameToIdMap[name] = catId
	CategoryIdToNameMap[catId] = name
	CategoryNames = append(CategoryNames, name)
	CategoriesMutex.Unlock()

	return
}

func (db *DataBase) DeleteCategory(name string) error {
	if len(CategoryNames) == 0 {
		_, err := db.GetAllCategories()
		if err != nil {
			return fmt.Errorf("get categories from delete category when category vars empty %w", err)
		}
	}

	errorMsg := "delete category: %w"

	_, ok := CategoryNameToIdMap[name]
	if !ok {
		return fmt.Errorf(errorMsg, "category does not exist")
	}

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()

	r, err := exor.ExecContext(ctx, DELETE_CATEGORY, name)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	if err = CatchNoRowsErr(r); err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	CategoriesMutex.Lock()
	CategoryIdToNameMap[CategoryNameToIdMap[name]] = ""
	delete(CategoryNameToIdMap, name)
	CategoryNames = slices.DeleteFunc(CategoryNames, func(catName string) bool {
		return catName == name
	})
	CategoriesMutex.Unlock()

	return nil
}

func (db *DataBase) GetCategoryIdByName(name string) (catId int64, err error) {
	if len(CategoryNames) == 0 {
		_, err := db.GetAllCategories()
		if err != nil {
			return 0, fmt.Errorf("get categories from GetCategoryIdByName when category vars empty %w", err)
		}
	}

	catId, ok := CategoryNameToIdMap[name]
	if !ok {
		return 0, errors.New("cant' find category name in map: " + name)
	}

	return catId, nil
}

func (db *DataBase) GetCatNameById(catId int64) (string, error) {
	if len(CategoryNames) == 0 {
		_, err := db.GetAllCategories()
		if err != nil {
			return "", fmt.Errorf("get categories from GetCatNameById when category vars empty %w", err)
		}
	}

	res := CategoryIdToNameMap[catId]
	if res == "" {
		return "", fmt.Errorf("categoryId doesn't exit in slicemap: %d", catId)
	}
	return res, nil
}

func (db *DataBase) GetAllCategories() ([]string, error) {
	loadCategoryVars := false
	if len(CategoryNames) != 0 {
		return CategoryNames, nil
	} else {
		loadCategoryVars = true
	}

	categories := []string{}

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()
	rows, err := exor.QueryContext(ctx, GET_ALL_CATEGORIES)
	if err != nil {
		return categories, err
	}
	defer rows.Close()

	if loadCategoryVars {
		CategoriesMutex.Lock()
		defer CategoriesMutex.Unlock()
	}

	for rows.Next() {
		var category string
		var id int64
		err := rows.Scan(&category, &id)
		if err != nil {
			log.Printf("Categories GetAllCategories err %v\n", err)
			continue
		}
		categories = append(categories, category)

		if loadCategoryVars {
			CategoryNameToIdMap[category] = id
			CategoryIdToNameMap[id] = category
			CategoryNames = append(CategoryNames, category)
		}

	}

	log.Println("Categories updated by sql database")
	return categories, nil
}
