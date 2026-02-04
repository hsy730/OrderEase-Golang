package tag

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"orderease/models"
)

// setupTestDB 创建测试用的 mock 数据库
func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dialector := mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	return db, mock, sqlDB
}

// ==================== UpdateProductTags Tests ====================

func TestTagService_UpdateProductTags(t *testing.T) {
	tests := []struct {
		name          string
		dto           UpdateProductTagsDTO
		setupMock     func(mock sqlmock.Sqlmock)
		expectedErr   bool
		expectedMsg   string
		validateResult func(t *testing.T, result *UpdateProductTagsResult)
	}{
		{
			name: "successfully add new tags",
			dto: UpdateProductTagsDTO{
				CurrentTags: []models.Tag{
					{ID: 1, Name: "Tag1"},
				},
				NewTagIDs: []int{1, 2, 3},
				ProductID: 123,
				ShopID:    456,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock transaction begin
				mock.ExpectBegin()

				// Mock insert new product tags
				mock.ExpectExec("INSERT INTO `product_tags`").
					WillReturnResult(sqlmock.NewResult(1, 2))

				// Mock transaction commit
				mock.ExpectCommit()
			},
			expectedErr: false,
			validateResult: func(t *testing.T, result *UpdateProductTagsResult) {
				assert.Equal(t, 2, result.AddedCount)   // Tags 2 and 3 are new
				assert.Equal(t, 0, result.DeletedCount) // No tags deleted
			},
		},
		{
			name: "successfully delete tags",
			dto: UpdateProductTagsDTO{
				CurrentTags: []models.Tag{
					{ID: 1, Name: "Tag1"},
					{ID: 2, Name: "Tag2"},
					{ID: 3, Name: "Tag3"},
				},
				NewTagIDs: []int{1},
				ProductID: 123,
				ShopID:    456,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock transaction begin
				mock.ExpectBegin()

				// Mock delete product tags
				mock.ExpectExec("DELETE FROM `product_tags`").
					WillReturnResult(sqlmock.NewResult(1, 2))

				// Mock transaction commit
				mock.ExpectCommit()
			},
			expectedErr: false,
			validateResult: func(t *testing.T, result *UpdateProductTagsResult) {
				assert.Equal(t, 0, result.AddedCount)   // No tags added
				assert.Equal(t, 2, result.DeletedCount) // Tags 2 and 3 deleted
			},
		},
		{
			name: "successfully both add and delete tags",
			dto: UpdateProductTagsDTO{
				CurrentTags: []models.Tag{
					{ID: 1, Name: "Tag1"},
					{ID: 2, Name: "Tag2"},
					{ID: 3, Name: "Tag3"},
				},
				NewTagIDs: []int{2, 4, 5},
				ProductID: 123,
				ShopID:    456,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock transaction begin
				mock.ExpectBegin()

				// Mock insert new product tags
				mock.ExpectExec("INSERT INTO `product_tags`").
					WillReturnResult(sqlmock.NewResult(1, 2))

				// Mock delete product tags
				mock.ExpectExec("DELETE FROM `product_tags`").
					WillReturnResult(sqlmock.NewResult(1, 2))

				// Mock transaction commit
				mock.ExpectCommit()
			},
			expectedErr: false,
			validateResult: func(t *testing.T, result *UpdateProductTagsResult) {
				assert.Equal(t, 2, result.AddedCount)   // Tags 4 and 5 added
				assert.Equal(t, 2, result.DeletedCount) // Tags 1 and 3 deleted
			},
		},
		{
			name: "no changes needed - tags already match",
			dto: UpdateProductTagsDTO{
				CurrentTags: []models.Tag{
					{ID: 1, Name: "Tag1"},
					{ID: 2, Name: "Tag2"},
				},
				NewTagIDs: []int{1, 2},
				ProductID: 123,
				ShopID:    456,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock transaction begin
				mock.ExpectBegin()
				// No insert or delete operations
				// Mock transaction commit
				mock.ExpectCommit()
			},
			expectedErr: false,
			validateResult: func(t *testing.T, result *UpdateProductTagsResult) {
				assert.Equal(t, 0, result.AddedCount)
				assert.Equal(t, 0, result.DeletedCount)
			},
		},
		{
			name: "replace all tags - delete all and add new",
			dto: UpdateProductTagsDTO{
				CurrentTags: []models.Tag{
					{ID: 1, Name: "Tag1"},
					{ID: 2, Name: "Tag2"},
				},
				NewTagIDs: []int{3, 4, 5},
				ProductID: 123,
				ShopID:    456,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock transaction begin
				mock.ExpectBegin()

				// Mock insert new product tags
				mock.ExpectExec("INSERT INTO `product_tags`").
					WillReturnResult(sqlmock.NewResult(1, 3))

				// Mock delete product tags
				mock.ExpectExec("DELETE FROM `product_tags`").
					WillReturnResult(sqlmock.NewResult(1, 2))

				// Mock transaction commit
				mock.ExpectCommit()
			},
			expectedErr: false,
			validateResult: func(t *testing.T, result *UpdateProductTagsResult) {
				assert.Equal(t, 3, result.AddedCount)   // Tags 3, 4, 5 added
				assert.Equal(t, 2, result.DeletedCount) // Tags 1, 2 deleted
			},
		},
		{
			name: "clear all tags - no new tags",
			dto: UpdateProductTagsDTO{
				CurrentTags: []models.Tag{
					{ID: 1, Name: "Tag1"},
					{ID: 2, Name: "Tag2"},
				},
				NewTagIDs: []int{},
				ProductID: 123,
				ShopID:    456,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock transaction begin
				mock.ExpectBegin()

				// Mock delete all product tags
				mock.ExpectExec("DELETE FROM `product_tags`").
					WillReturnResult(sqlmock.NewResult(1, 2))

				// Mock transaction commit
				mock.ExpectCommit()
			},
			expectedErr: false,
			validateResult: func(t *testing.T, result *UpdateProductTagsResult) {
				assert.Equal(t, 0, result.AddedCount)   // No tags added
				assert.Equal(t, 2, result.DeletedCount) // All tags deleted
			},
		},
		{
			name: "add first tag to product with no existing tags",
			dto: UpdateProductTagsDTO{
				CurrentTags: []models.Tag{},
				NewTagIDs:    []int{1},
				ProductID:    123,
				ShopID:       456,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock transaction begin
				mock.ExpectBegin()

				// Mock insert new product tag
				mock.ExpectExec("INSERT INTO `product_tags`").
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Mock transaction commit
				mock.ExpectCommit()
			},
			expectedErr: false,
			validateResult: func(t *testing.T, result *UpdateProductTagsResult) {
				assert.Equal(t, 1, result.AddedCount)
				assert.Equal(t, 0, result.DeletedCount)
			},
		},
		{
			name: "insert operation fails",
			dto: UpdateProductTagsDTO{
				CurrentTags: []models.Tag{},
				NewTagIDs:    []int{1, 2},
				ProductID:    123,
				ShopID:       456,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock transaction begin
				mock.ExpectBegin()

				// Mock insert failure
				mock.ExpectExec("INSERT INTO `product_tags`").
					WillReturnError(fmt.Errorf("database connection error"))

				// Mock transaction rollback
				mock.ExpectRollback()
			},
			expectedErr: true,
			expectedMsg: "database connection error",
		},
		{
			name: "delete operation fails",
			dto: UpdateProductTagsDTO{
				CurrentTags: []models.Tag{
					{ID: 1, Name: "Tag1"},
				},
				NewTagIDs: []int{},
				ProductID: 123,
				ShopID:    456,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock transaction begin
				mock.ExpectBegin()

				// Mock delete failure
				mock.ExpectExec("DELETE FROM `product_tags`").
					WillReturnError(fmt.Errorf("database connection error"))

				// Mock transaction rollback
				mock.ExpectRollback()
			},
			expectedErr: true,
			expectedMsg: "database connection error",
		},
		{
			name: "transaction commit fails",
			dto: UpdateProductTagsDTO{
				CurrentTags: []models.Tag{},
				NewTagIDs:    []int{1},
				ProductID:    123,
				ShopID:       456,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock transaction begin
				mock.ExpectBegin()

				// Mock insert success
				mock.ExpectExec("INSERT INTO `product_tags`").
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Mock commit failure
				mock.ExpectCommit().WillReturnError(fmt.Errorf("commit failed"))
			},
			expectedErr: true,
			expectedMsg: "commit failed",
		},
		{
			name: "large number of tags to add and delete",
			dto: UpdateProductTagsDTO{
				CurrentTags: []models.Tag{
					{ID: 1, Name: "Tag1"},
					{ID: 2, Name: "Tag2"},
					{ID: 3, Name: "Tag3"},
					{ID: 4, Name: "Tag4"},
					{ID: 5, Name: "Tag5"},
				},
				NewTagIDs: []int{3, 4, 6, 7, 8, 9, 10},
				ProductID: 123,
				ShopID:    456,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Mock transaction begin
				mock.ExpectBegin()

				// Mock insert new product tags (5 new tags: 6,7,8,9,10)
				mock.ExpectExec("INSERT INTO `product_tags`").
					WillReturnResult(sqlmock.NewResult(1, 5))

				// Mock delete product tags (2 tags: 1,2)
				mock.ExpectExec("DELETE FROM `product_tags`").
					WillReturnResult(sqlmock.NewResult(1, 2))

				// Mock transaction commit
				mock.ExpectCommit()
			},
			expectedErr: false,
			validateResult: func(t *testing.T, result *UpdateProductTagsResult) {
				assert.Equal(t, 5, result.AddedCount)   // Tags 6,7,8,9,10 added
				assert.Equal(t, 3, result.DeletedCount) // Tags 1,2,5 deleted
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, sqlDB := setupTestDB(t)
			defer sqlDB.Close()

			service := NewService(db)

			tt.setupMock(mock)

			result, err := service.UpdateProductTags(tt.dto)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ==================== Edge Case Tests ====================

func TestTagService_UpdateProductTags_EdgeCases(t *testing.T) {
	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	service := NewService(db)

	t.Run("duplicate tag IDs in new tags", func(t *testing.T) {
		dto := UpdateProductTagsDTO{
			CurrentTags: []models.Tag{},
			NewTagIDs:    []int{1, 1, 2, 2}, // Duplicates
			ProductID:    123,
			ShopID:       456,
		}

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `product_tags`").
			WillReturnResult(sqlmock.NewResult(1, 4)) // Will insert all duplicates
		mock.ExpectCommit()

		result, err := service.UpdateProductTags(dto)

		assert.NoError(t, err)
		assert.Equal(t, 4, result.AddedCount) // Will attempt to add duplicates
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty current tags and empty new tags", func(t *testing.T) {
		dto := UpdateProductTagsDTO{
			CurrentTags: []models.Tag{},
			NewTagIDs:    []int{},
			ProductID:    123,
			ShopID:       456,
		}

		mock.ExpectBegin()
		mock.ExpectCommit()

		result, err := service.UpdateProductTags(dto)

		assert.NoError(t, err)
		assert.Equal(t, 0, result.AddedCount)
		assert.Equal(t, 0, result.DeletedCount)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("zero product ID", func(t *testing.T) {
		dto := UpdateProductTagsDTO{
			CurrentTags: []models.Tag{},
			NewTagIDs:    []int{1},
			ProductID:    0,
			ShopID:       456,
		}

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `product_tags`").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		result, err := service.UpdateProductTags(dto)

		assert.NoError(t, err)
		assert.Equal(t, 1, result.AddedCount)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("negative tag IDs", func(t *testing.T) {
		dto := UpdateProductTagsDTO{
			CurrentTags: []models.Tag{},
			NewTagIDs:    []int{-1, -2}, // Negative IDs (unusual but possible)
			ProductID:    123,
			ShopID:       456,
		}

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `product_tags`").
			WillReturnResult(sqlmock.NewResult(1, 2))
		mock.ExpectCommit()

		result, err := service.UpdateProductTags(dto)

		assert.NoError(t, err)
		assert.Equal(t, 2, result.AddedCount)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ==================== Performance Tests ====================

func TestTagService_UpdateProductTags_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	db, mock, sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	service := NewService(db)

	t.Run("batch operations with 100 tags", func(t *testing.T) {
		// Create 100 current tags
		currentTags := make([]models.Tag, 100)
		for i := 0; i < 100; i++ {
			currentTags[i] = models.Tag{ID: i + 1, Name: fmt.Sprintf("Tag%d", i+1)}
		}

		// Create 100 new tags (replace all)
		newTagIDs := make([]int, 100)
		for i := 0; i < 100; i++ {
			newTagIDs[i] = i + 101 // Different set
		}

		dto := UpdateProductTagsDTO{
			CurrentTags: currentTags,
			NewTagIDs:    newTagIDs,
			ProductID:    123,
			ShopID:       456,
		}

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO `product_tags`").
			WillReturnResult(sqlmock.NewResult(1, 100))
		mock.ExpectExec("DELETE FROM `product_tags`").
			WillReturnResult(sqlmock.NewResult(1, 100))
		mock.ExpectCommit()

		result, err := service.UpdateProductTags(dto)

		assert.NoError(t, err)
		assert.Equal(t, 100, result.AddedCount)
		assert.Equal(t, 100, result.DeletedCount)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
