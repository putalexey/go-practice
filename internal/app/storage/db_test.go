package storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var db *sql.DB

func withDockerDB(f func(databaseDSN string, db *sql.DB)) error {
	databaseDSN := "postgres://postgres:postgres@postgres:5432/praktikum"
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	f(databaseDSN, db)
	return nil

	//// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	//pool, err := dockertest.NewPool("")
	//if err != nil {
	//	return fmt.Errorf("could not connect to docker: %w", err)
	//}
	//
	//// pulls an image, creates a container based on it and runs it
	//resource, err := pool.Run("postgres", "latest", []string{"POSTGRES_PASSWORD=secret"})
	//if err != nil {
	//	return fmt.Errorf("could not start resource: %w", err)
	//}
	//defer func() {
	//	// You can't defer this because os.Exit doesn't care for defer
	//	if err := pool.Purge(resource); err != nil {
	//		log.Fatalf("Could not purge resource: %s", err)
	//	}
	//}()
	//
	//databaseDSN := fmt.Sprintf("postgres://postgres:secret@localhost:%s/postgres", resource.GetPort("5432/tcp"))
	//// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	//err = pool.Retry(func() error {
	//	var err error
	//	db, err = sql.Open("pgx", databaseDSN)
	//	if err != nil {
	//		return err
	//	}
	//	return db.Ping()
	//})
	//if err != nil {
	//	return fmt.Errorf("could not connect to database: %w", err)
	//}
	//
	//f(databaseDSN, db)

	//return nil
}

func TestDBStorage_Delete(t *testing.T) {
	type args struct {
		ctx   context.Context
		short string
	}
	tests := []struct {
		name      string
		args      args
		wantErr   assert.ErrorAssertionFunc
		mockSetup func(sqlmock.Sqlmock)
	}{
		{
			name: "Successfully deletes record",
			args: args{
				ctx:   context.Background(),
				short: "short-1",
			},
			wantErr: assert.NoError,
			mockSetup: func(s sqlmock.Sqlmock) {
				s.ExpectExec("UPDATE shorts SET deleted = TRUE WHERE short = \\$1").
					WithArgs("short-1").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "Tries to delete with empty `short` value",
			args: args{
				ctx:   context.Background(),
				short: "",
			},
			wantErr: assert.NoError,
			mockSetup: func(s sqlmock.Sqlmock) {
				s.ExpectExec("UPDATE shorts SET deleted = TRUE WHERE short = \\$1").
					WithArgs("").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "Return error, when something wrong with db",
			args: args{
				ctx:   context.Background(),
				short: "test-1",
			},
			wantErr: assert.Error,
			mockSetup: func(s sqlmock.Sqlmock) {
				s.ExpectExec("UPDATE shorts SET deleted = TRUE WHERE short = \\$1").
					WithArgs("test-1").
					WillReturnError(driver.ErrBadConn)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()
			tt.mockSetup(mock)

			s := &DBStorage{
				db: db,
			}
			tt.wantErr(t, s.Delete(tt.args.ctx, tt.args.short), fmt.Sprintf("Delete(%v, %v)", tt.args.ctx, tt.args.short))

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDBStorage_DeleteBatch(t *testing.T) {
	type args struct {
		ctx    context.Context
		shorts []string
	}
	tests := []struct {
		name      string
		args      args
		wantErr   assert.ErrorAssertionFunc
		mockSetup func(sqlmock.Sqlmock)
	}{
		{
			name: "Successfully deletes multiple records",
			args: args{
				ctx:    context.Background(),
				shorts: []string{"short-1", "short-2", "short-3"},
			},
			wantErr: assert.NoError,
			mockSetup: func(s sqlmock.Sqlmock) {
				s.ExpectExec("UPDATE shorts SET deleted = TRUE WHERE short IN \\(\\$1,\\s*\\$2,\\s*\\$3\\)").
					WithArgs("short-1", "short-2", "short-3").
					WillReturnResult(sqlmock.NewResult(0, 3))
			},
		},
		{
			name: "Nothing executed when list is empty",
			args: args{
				ctx:    context.Background(),
				shorts: []string{},
			},
			wantErr:   assert.NoError,
			mockSetup: func(s sqlmock.Sqlmock) {},
		},
		{
			name: "Return error, when something wrong with db",
			args: args{
				ctx:    context.Background(),
				shorts: []string{"test-1"},
			},
			wantErr: assert.Error,
			mockSetup: func(s sqlmock.Sqlmock) {
				s.ExpectExec("UPDATE shorts SET deleted = TRUE WHERE short IN \\(\\$1\\)").
					WithArgs("test-1").
					WillReturnError(driver.ErrBadConn)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()
			tt.mockSetup(mock)

			s := &DBStorage{
				db: db,
			}
			tt.wantErr(t, s.DeleteBatch(tt.args.ctx, tt.args.shorts), fmt.Sprintf("DeleteBatch(%v, %v)", tt.args.ctx, tt.args.shorts))

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDBStorage_Load(t *testing.T) {
	type args struct {
		ctx   context.Context
		short string
	}
	columns := []string{
		"short",
		"original",
		"user_id",
		"deleted",
	}
	tests := []struct {
		name      string
		args      args
		want      Record
		wantErr   assert.ErrorAssertionFunc
		mockSetup func(sqlmock.Sqlmock)
	}{
		{
			name: "Return record, when found",
			args: args{
				ctx:   context.Background(),
				short: "short-1",
			},
			want: Record{
				Short:   "short-1",
				Full:    "https://example.com/asd",
				UserID:  "2",
				Deleted: false,
			},
			wantErr: assert.NoError,
			mockSetup: func(s sqlmock.Sqlmock) {
				s.ExpectQuery("SELECT (.+) FROM shorts WHERE short = \\$1 LIMIT 1").
					WithArgs("short-1").
					WillReturnRows(
						sqlmock.
							NewRows(columns).
							AddRow("short-1", "https://example.com/asd", "2", "0"),
					)
			},
		},
		{
			name: "Return NoRows error, when not found",
			args: args{
				ctx:   context.Background(),
				short: "short-1",
			},
			want: Record{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				fmt.Println(err)
				var e *RecordNotFoundError
				return assert.ErrorAs(t, err, &e, i...)
			},
			mockSetup: func(s sqlmock.Sqlmock) {
				s.ExpectQuery("SELECT (.+) FROM shorts WHERE short = \\$1 LIMIT 1").
					WithArgs("short-1").
					WillReturnError(sql.ErrNoRows)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()
			tt.mockSetup(mock)

			s := &DBStorage{
				db: db,
			}
			got, err := s.Load(tt.args.ctx, tt.args.short)
			if !tt.wantErr(t, err, fmt.Sprintf("Load(%v, %v)", tt.args.ctx, tt.args.short)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Load(%v, %v)", tt.args.ctx, tt.args.short)

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDBStorage_LoadBatch(t *testing.T) {
	type args struct {
		ctx    context.Context
		shorts []string
	}
	columns := []string{
		"short",
		"original",
		"user_id",
		"deleted",
	}
	tests := []struct {
		name      string
		args      args
		want      []Record
		wantErr   assert.ErrorAssertionFunc
		mockSetup func(sqlmock.Sqlmock)
	}{
		{
			name: "Successfully loads records",
			args: args{
				ctx:    context.Background(),
				shorts: []string{"short-1", "short-2", "short-3"},
			},
			want: []Record{
				{
					Short:   "short-1",
					Full:    "https://example.com/asd",
					UserID:  "1",
					Deleted: false,
				},
				{
					Short:   "short-2",
					Full:    "https://example.com/asd123",
					UserID:  "1",
					Deleted: false,
				},
			},
			wantErr: assert.NoError,
			mockSetup: func(s sqlmock.Sqlmock) {
				s.ExpectQuery("SELECT (.+) FROM shorts WHERE short in \\(\\$1,\\$2,\\$3\\) and deleted = FALSE").
					WithArgs("short-1", "short-2", "short-3").
					WillReturnRows(
						sqlmock.
							NewRows(columns).
							AddRow("short-1", "https://example.com/asd", "1", "0").
							AddRow("short-2", "https://example.com/asd123", "1", "0"),
					)
			},
		},
		{
			name: "Empty list, when no rows found",
			args: args{
				ctx:    context.Background(),
				shorts: []string{"short-1", "short-2", "short-3"},
			},
			want:    []Record{},
			wantErr: assert.NoError,
			mockSetup: func(s sqlmock.Sqlmock) {
				s.ExpectQuery("SELECT (.+) FROM shorts WHERE short in \\(\\$1,\\$2,\\$3\\) and deleted = FALSE").
					WithArgs("short-1", "short-2", "short-3").
					WillReturnRows(
						sqlmock.NewRows(columns),
					)
			},
		},
		{
			name: "Empty list, when query shorts list is empty",
			args: args{
				ctx:    context.Background(),
				shorts: []string{},
			},
			want:      []Record{},
			wantErr:   assert.NoError,
			mockSetup: func(s sqlmock.Sqlmock) {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()
			tt.mockSetup(mock)

			s := &DBStorage{
				db: db,
			}
			got, err := s.LoadBatch(tt.args.ctx, tt.args.shorts)
			if !tt.wantErr(t, err, fmt.Sprintf("LoadBatch(%v, %v)", tt.args.ctx, tt.args.shorts)) {
				return
			}
			assert.Equalf(t, tt.want, got, "LoadBatch(%v, %v)", tt.args.ctx, tt.args.shorts)

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDBStorage_LoadForUser(t *testing.T) {
	type args struct {
		ctx    context.Context
		userID string
	}
	columns := []string{
		"short",
		"original",
		"user_id",
		"deleted",
	}

	tests := []struct {
		name      string
		args      args
		want      []Record
		wantErr   assert.ErrorAssertionFunc
		mockSetup func(sqlmock.Sqlmock)
	}{
		{
			name: "Successfully stores record",
			args: args{
				ctx:    context.Background(),
				userID: "1",
			},
			want: []Record{
				{
					Short:   "short-1",
					Full:    "https://example.com/asd",
					UserID:  "1",
					Deleted: false,
				},
				{
					Short:   "short-2",
					Full:    "https://example.com/asd123",
					UserID:  "1",
					Deleted: false,
				},
			},
			wantErr: assert.NoError,
			mockSetup: func(s sqlmock.Sqlmock) {
				s.ExpectQuery("SELECT (.+) FROM shorts WHERE user_id = \\$1 and deleted = FALSE").
					WithArgs("1").
					WillReturnRows(
						sqlmock.
							NewRows(columns).
							AddRow("short-1", "https://example.com/asd", "1", "0").
							AddRow("short-2", "https://example.com/asd123", "1", "0"),
					)
			},
		},
		{
			name: "Empty list, when no rows found",
			args: args{
				ctx:    context.Background(),
				userID: "1",
			},
			want:    []Record{},
			wantErr: assert.NoError,
			mockSetup: func(s sqlmock.Sqlmock) {
				s.ExpectQuery("SELECT (.+) FROM shorts WHERE user_id = \\$1 and deleted = FALSE").
					WithArgs("1").
					WillReturnRows(
						sqlmock.NewRows(columns),
					)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()
			tt.mockSetup(mock)

			s := &DBStorage{
				db: db,
			}
			got, err := s.LoadForUser(tt.args.ctx, tt.args.userID)
			if !tt.wantErr(t, err, fmt.Sprintf("LoadForUser(%v, %v)", tt.args.ctx, tt.args.userID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "LoadForUser(%v, %v)", tt.args.ctx, tt.args.userID)

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDBStorage_Ping(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectPing()

	s := &DBStorage{
		db: db,
	}
	ctx := context.Background()
	assert.NoError(t, s.Ping(ctx), fmt.Sprintf("Ping(%v)", ctx))

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDBStorage_Store(t *testing.T) {
	type args struct {
		ctx    context.Context
		record Record
	}
	tests := []struct {
		name      string
		args      args
		wantErr   assert.ErrorAssertionFunc
		mockSetup func(sqlmock.Sqlmock)
	}{
		{
			name: "Successfully stores record",
			args: args{
				ctx: context.Background(),
				record: Record{
					Short:   "short-1",
					Full:    "https://example.com/asd",
					UserID:  "1",
					Deleted: false,
				},
			},
			wantErr: assert.NoError,
			mockSetup: func(s sqlmock.Sqlmock) {
				s.ExpectExec("INSERT INTO shorts").
					WithArgs("short-1", "https://example.com/asd", "1").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "Returns RecordConflictError with existing record",
			args: args{
				ctx: context.Background(),
				record: Record{
					Short:   "short-2",
					Full:    "https://example.com/asd",
					UserID:  "1",
					Deleted: false,
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				var e *RecordConflictError
				return assert.ErrorAs(t, err, &e, i...)
			},
			mockSetup: func(s sqlmock.Sqlmock) {
				s.ExpectExec("INSERT INTO shorts").
					WithArgs("short-2", "https://example.com/asd", "1").
					WillReturnResult(sqlmock.NewResult(0, 0))
				s.ExpectQuery("SELECT (.+) FROM shorts WHERE \"original\"").
					WithArgs("https://example.com/asd").
					WillReturnRows(
						sqlmock.
							NewRows([]string{"short", "original", "user_id"}).
							AddRow("short-1", "https://example.com/asd", "2"),
					)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()
			tt.mockSetup(mock)

			s := &DBStorage{
				db: db,
			}
			tt.wantErr(t, s.Store(tt.args.ctx, tt.args.record), fmt.Sprintf("Store(%v, %v)", tt.args.ctx, tt.args.record))

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDBStorage_StoreBatch(t *testing.T) {
	type args struct {
		ctx     context.Context
		records []Record
	}
	tests := []struct {
		name      string
		args      args
		wantErr   assert.ErrorAssertionFunc
		mockSetup func(sqlmock.Sqlmock)
	}{
		{
			name: "Successfully stores record",
			args: args{
				ctx: context.Background(),
				records: []Record{
					Record{
						Short:   "short-1",
						Full:    "https://example.com/asd",
						UserID:  "1",
						Deleted: false,
					},
				},
			},
			wantErr: assert.NoError,
			mockSetup: func(s sqlmock.Sqlmock) {
				s.ExpectBegin()
				s.ExpectPrepare("INSERT INTO shorts").
					ExpectExec().
					WithArgs("short-1", "https://example.com/asd", "1").
					WillReturnResult(sqlmock.NewResult(0, 1))
				s.ExpectCommit()
			},
		},
		{
			name: "Not executes anything, if recods list empty",
			args: args{
				ctx:     context.Background(),
				records: []Record{},
			},
			wantErr:   assert.NoError,
			mockSetup: func(s sqlmock.Sqlmock) {},
		},
		//{
		//	name: "Returns RecordConflictError with existing record",
		//	args: args{
		//		ctx: context.Background(),
		//		records: Record{
		//			Short:   "short-2",
		//			Full:    "https://example.com/asd",
		//			UserID:  "1",
		//			Deleted: false,
		//		},
		//	},
		//	wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
		//		var e *RecordConflictError
		//		return assert.ErrorAs(t, err, &e, i...)
		//	},
		//	mockSetup: func(s sqlmock.Sqlmock) {
		//		s.ExpectExec("INSERT INTO shorts").
		//			WithArgs("short-2", "https://example.com/asd", "1").
		//			WillReturnResult(sqlmock.NewResult(0, 0))
		//		s.ExpectQuery("SELECT (.+) FROM shorts WHERE \"original\"").
		//			WithArgs("https://example.com/asd").
		//			WillReturnRows(
		//				sqlmock.
		//					NewRows([]string{"short", "original", "user_id"}).
		//					AddRow("short-1", "https://example.com/asd", "2"),
		//			)
		//	},
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()
			tt.mockSetup(mock)

			s := &DBStorage{
				db: db,
			}
			tt.wantErr(t, s.StoreBatch(tt.args.ctx, tt.args.records), fmt.Sprintf("StoreBatch(%v, %v)", tt.args.ctx, tt.args.records))

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestNewDBStorage(t *testing.T) {
	err := withDockerDB(func(databaseDSN string, db *sql.DB) {
		type args struct {
			migrationsDir string
			databaseDSN   string
		}
		t.Run("Correctly connects to db", func(t *testing.T) {
			migrationsDir := ""
			databaseDSN := databaseDSN

			wantErr := assert.NoError
			wantCloseErr := assert.NoError
			got, err := NewDBStorage(databaseDSN, migrationsDir)
			if !wantErr(t, err, fmt.Sprintf("NewDBStorage(%v, %v)", databaseDSN, migrationsDir)) {
				return
			}
			require.NotNilf(t, got, "NewDBStorage(%v, %v)", databaseDSN, migrationsDir)
			err = got.Close()
			wantCloseErr(t, err, fmt.Sprintf("NewDBStorage(%v, %v).Close", databaseDSN, migrationsDir))
		})

		tests := []struct {
			name    string
			args    args
			wantErr assert.ErrorAssertionFunc
		}{
			{
				name: "Directory not found error",
				args: args{
					migrationsDir: "123",
					databaseDSN:   databaseDSN,
				},
				wantErr: assert.Error,
			},
			{
				name: "Wrong DSN error",
				args: args{
					migrationsDir: "",
					databaseDSN:   "postgres://wrongDSN",
				},
				wantErr: assert.Error,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := NewDBStorage(tt.args.databaseDSN, tt.args.migrationsDir)
				if !tt.wantErr(t, err, fmt.Sprintf("NewDBStorage(%v, %v)", tt.args.databaseDSN, tt.args.migrationsDir)) {
					return
				}
			})
		}
	})
	require.NoError(t, err)
}

func Test_prepareSQLPlaceholders(t *testing.T) {
	type args struct {
		startIndex int
		values     []string
	}
	tests := []struct {
		name  string
		args  args
		want  []string
		want1 []interface{}
	}{
		{
			name:  "index starts from 1",
			args:  args{startIndex: 1, values: []string{"test 1", "test 2"}},
			want:  []string{"$1", "$2"},
			want1: []interface{}{"test 1", "test 2"},
		},
		{
			name:  "index starts from 5",
			args:  args{startIndex: 5, values: []string{"test 3", "test 4"}},
			want:  []string{"$5", "$6"},
			want1: []interface{}{"test 3", "test 4"},
		},
		{
			name:  "empty values",
			args:  args{startIndex: 1, values: []string{}},
			want:  []string{},
			want1: []interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := prepareSQLPlaceholders(tt.args.startIndex, tt.args.values)
			assert.Equalf(t, tt.want, got, "prepareSQLPlaceholders(%v, %v)", tt.args.startIndex, tt.args.values)
			assert.Equalf(t, tt.want1, got1, "prepareSQLPlaceholders(%v, %v)", tt.args.startIndex, tt.args.values)
		})
	}
}
