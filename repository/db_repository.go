package repository

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lib/pq"

	"github.com/AleksMa/StealLovingYou/models"
)

type Repo interface {
	PutUser(user *models.User) (uint64, *models.Error)
	GetUsers(user *models.User) ([]*models.User, *models.Error)
	GetUserByUsername(nickname string) (models.User, *models.Error)
	UpdateUser(user *models.User) (uint64, *models.Error)
	DeleteUserByUsername(nickname string) (models.User, *models.Error)

	PutTask(task *models.Task) (uint64, *models.Error)
	GetTasks(task *models.Task) ([]*models.Task, *models.Error)
	GetTaskByTaskname(taskname string) (models.Task, *models.Error)
	UpdateTask(task *models.Task) (uint64, *models.Error)
	DeleteTaskByTaskname(taskname string) (models.Task, *models.Error)

	PutAttempt(attempt *models.Attempt) (uint64, *models.Error)
	GetAttempt(task string, user string) ([]*models.Attempt, *models.Error)

	GetStatus() (models.Status, *models.Error)
	ReloadDB() *models.Error

	GetResult(task string, user string) ([]*models.Result, *models.Error)

	//PutHashes(ID uint64, hashSet models.HashSet) *models.Error
	//GetHashes(ID uint64) (*models.HashSet, *models.Error)
	//GetSimilarHashes(attempt *models.Attempt) ([]*models.HashObject, *models.Error)

	PutHashes2(ID uint64, hashSet models.HashSet) *models.Error
	GetHashes2(ID uint64) (*models.HashSet, *models.Error)
	GetSimilarHashes2(attempt *models.Attempt) ([]*models.HashObject, *models.Error)

	PutBorrowing(borrowing *models.Borrowing) *models.Error
}

type DBStore struct {
	DB  *pgxpool.Pool
	ctx context.Context
}

func NewDBStore(pool *pgxpool.Pool, ctx context.Context) Repo {
	return &DBStore{
		pool,
		ctx,
	}
}

func (store *DBStore) GetStatus() (models.Status, *models.Error) {
	tx, _ := store.DB.Begin(context.Background())
	defer tx.Rollback(context.Background())

	status := &models.Status{}

	row := tx.QueryRow(store.ctx, `SELECT count(*) FROM users`)
	row.Scan(&status.User)

	tx.Commit(store.ctx)

	return *status, nil
}

func (store *DBStore) ReloadDB() *models.Error {
	_, err := store.DB.Exec(store.ctx, models.InitScript)
	if err != nil {
		return models.NewError(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func (store *DBStore) PutUser(user *models.User) (uint64, *models.Error) {
	var ID uint64

	tx, err := store.DB.Begin(store.ctx)
	if err != nil {
		return 0, models.NewError(http.StatusInternalServerError, err.Error())
	}

	insertQuery := `INSERT INTO users (userName, fullName, studentID) VALUES ($1, $2, $3) RETURNING id`
	rows := tx.QueryRow(store.ctx, insertQuery,
		user.UserName, user.FullName, user.StudentID)

	err = rows.Scan(&ID)
	if err != nil {
		tx.Rollback(store.ctx)
		return 0, models.NewError(http.StatusInternalServerError, err.Error())
	}

	tx.Commit(store.ctx)
	return ID, nil
}

func (store *DBStore) UpdateUser(user *models.User) (uint64, *models.Error) {
	var ID uint64

	tx, err := store.DB.Begin(store.ctx)
	if err != nil {
		return 0, models.NewError(http.StatusInternalServerError, err.Error())
	}

	insertQuery := `UPDATE users SET FullName = $1, StudentID = $2 WHERE username = $3 RETURNING id`
	rows := tx.QueryRow(store.ctx, insertQuery,
		user.FullName, user.StudentID, user.UserName)

	err = rows.Scan(&ID)
	if err != nil {
		tx.Rollback(store.ctx)
		return 0, models.NewError(http.StatusInternalServerError, err.Error())
	}

	tx.Commit(store.ctx)
	return ID, nil
}

func (store *DBStore) GetUsers(user *models.User) ([]*models.User, *models.Error) {
	users := []*models.User{}

	selectStr := "SELECT DISTINCT userName, fullName, studentID FROM users WHERE userName=$1 OR studentID=$2"

	rows, err := store.DB.Query(store.ctx, selectStr, user.UserName, user.StudentID)
	if err != nil {
		fmt.Println(err)
		return users, models.NewError(http.StatusInternalServerError, err.Error())
	}

	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(&user.UserName, &user.FullName, &user.StudentID)
		if err != nil {
			return users, models.NewError(http.StatusInternalServerError, err.Error())
		}
		users = append(users, user)
	}

	rows.Close()

	if err != nil {
		return users, models.NewError(http.StatusInternalServerError, err.Error())
	}

	return users, nil
}

func (store *DBStore) GetUserByUsername(nickname string) (models.User, *models.Error) {
	user := &models.User{}

	selectStr := "SELECT id, userName, fullName, studentID FROM users WHERE username = $1"
	row := store.DB.QueryRow(store.ctx, selectStr, nickname)

	err := row.Scan(&user.ID, &user.UserName, &user.FullName, &user.StudentID)

	if err != nil {
		fmt.Println(err)
		if err == pgx.ErrNoRows {
			return *user, models.NewError(http.StatusNotFound, err.Error())
		}
		return *user, models.NewError(http.StatusInternalServerError, err.Error())
	}

	return *user, nil
}

func (store *DBStore) DeleteUserByUsername(nickname string) (models.User, *models.Error) {
	user := &models.User{}

	selectStr := "DELETE FROM users WHERE username = $1 RETURNING ID, UserName, FullName, StudentID"
	row := store.DB.QueryRow(store.ctx, selectStr, nickname)

	err := row.Scan(&user.ID, &user.UserName, &user.FullName, &user.StudentID)

	if err != nil {
		fmt.Println(err)
		if err == pgx.ErrNoRows || user.ID == 0 {
			return *user, models.NewError(http.StatusNotFound, err.Error())
		}
		return *user, models.NewError(http.StatusInternalServerError, err.Error())
	}

	return *user, nil
}

func (store *DBStore) PutTask(task *models.Task) (uint64, *models.Error) {
	fmt.Println(task)
	var ID uint64

	insertQuery := `INSERT INTO tasks (taskName, fullName, maxTime, maxMemory) VALUES ($1, $2, $3, $4) RETURNING ID`
	rows := store.DB.QueryRow(store.ctx, insertQuery,
		task.TaskName, task.FullName, task.MaxTime, task.MaxMemory)

	err := rows.Scan(&ID)
	if err != nil {
		fmt.Println(err)
		return 0, models.NewError(http.StatusInternalServerError, err.Error())
	}

	return ID, nil
}

func (store *DBStore) GetTasks(task *models.Task) ([]*models.Task, *models.Error) {
	tasks := []*models.Task{}

	selectStr := "SELECT DISTINCT taskName, fullName, maxTime, maxMemory FROM tasks WHERE taskName=$1"

	rows, err := store.DB.Query(store.ctx, selectStr, task.TaskName)
	if err != nil {
		fmt.Println(err)
		return tasks, models.NewError(http.StatusInternalServerError, err.Error())
	}

	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(&task.TaskName, &task.FullName, &task.MaxTime, &task.MaxMemory)
		if err != nil {
			return tasks, models.NewError(http.StatusInternalServerError, err.Error())
		}
		tasks = append(tasks, task)
	}

	rows.Close()

	if err != nil {
		return tasks, models.NewError(http.StatusInternalServerError, err.Error())
	}

	return tasks, nil
}

func (store *DBStore) GetTaskByTaskname(taskname string) (models.Task, *models.Error) {
	task := &models.Task{}

	selectStr := "SELECT id, taskName, fullName, maxTime, maxMemory FROM tasks WHERE taskname = $1"
	row := store.DB.QueryRow(store.ctx, selectStr, taskname)

	err := row.Scan(&task.ID, &task.TaskName, &task.FullName, &task.MaxTime, &task.MaxMemory)

	if err != nil {
		fmt.Println(err)
		if err == pgx.ErrNoRows {
			return *task, models.NewError(http.StatusNotFound, err.Error())
		}
		return *task, models.NewError(http.StatusInternalServerError, err.Error())
	}

	return *task, nil
}

func (store *DBStore) UpdateTask(task *models.Task) (uint64, *models.Error) {
	fmt.Println(task)
	var ID uint64

	insertQuery := `UPDATE tasks SET FullName = $1, maxTime = $2, maxMemory = $3 WHERE taskname = $4 RETURNING id`
	rows := store.DB.QueryRow(store.ctx, insertQuery,
		task.FullName, task.MaxTime, task.MaxMemory,
		task.TaskName)

	err := rows.Scan(&ID)
	if err != nil {
		fmt.Println(err)
		return 0, models.NewError(http.StatusInternalServerError, err.Error())
	}

	return ID, nil
}

func (store *DBStore) DeleteTaskByTaskname(taskname string) (models.Task, *models.Error) {
	task := &models.Task{}

	selectStr := "DELETE FROM tasks WHERE taskname = $1 RETURNING ID, TaskName, FullName, MaxTime, MaxMemory"
	row := store.DB.QueryRow(store.ctx, selectStr, taskname)

	err := row.Scan(&task.ID, &task.TaskName, &task.FullName, &task.MaxTime, &task.MaxMemory)

	if err != nil {
		fmt.Println(err)
		if err == pgx.ErrNoRows || task.ID == 0 {
			return *task, models.NewError(http.StatusNotFound, err.Error())
		}
		return *task, models.NewError(http.StatusInternalServerError, err.Error())
	}

	return *task, nil
}

func (store *DBStore) PutAttempt(attempt *models.Attempt) (uint64, *models.Error) {
	fmt.Println(attempt)
	var ID uint64

	insertQuery := `INSERT INTO attempts (userID, taskID, memory, time, sourceCode, uploadDate, status) VALUES 
						((SELECT ID FROM users WHERE userName = $1),
						(SELECT ID FROM tasks WHERE taskName = $2), 
						$3, $4, $5, $6, $7) RETURNING ID`
	rows := store.DB.QueryRow(store.ctx, insertQuery,
		attempt.User, attempt.Task, attempt.Memory, attempt.Time, attempt.SourceCode, attempt.UploadDate, 1)

	fmt.Println(attempt.UploadDate)

	err := rows.Scan(&ID)
	if err != nil {
		fmt.Println(err)
		return 0, models.NewError(http.StatusInternalServerError, err.Error())
	}

	return ID, nil
}

func (store *DBStore) GetAttempt(task string, user string) ([]*models.Attempt, *models.Error) {
	var attempts []*models.Attempt
	var args []interface{}

	selectStr := `SELECT u.userName, t.taskName, a.time, a.memory, a.uploaddate, a.sourcecode
					FROM users u
         			JOIN attempts a on u.ID = a.userID
         			JOIN tasks t ON a.taskID = t.ID`

	if task != "" {
		selectStr += " WHERE t.taskname=$1"
		args = append(args, task)
	}

	if user != "" {
		if task != "" {
			selectStr += " AND u.username=$2"
		} else {
			selectStr += " WHERE u.username=$1"
		}
		args = append(args, user)
	}
	selectStr += ";"

	rows, err := store.DB.Query(store.ctx, selectStr, args...)
	if err != nil {
		fmt.Println(err)
		return attempts, models.NewError(http.StatusInternalServerError, err.Error())
	}

	for rows.Next() {
		attempt := &models.Attempt{}
		err := rows.Scan(&attempt.User, &attempt.Task, &attempt.Time,
			&attempt.Memory, &attempt.UploadDate, &attempt.SourceCode)
		if err != nil {
			return attempts, models.NewError(http.StatusInternalServerError, err.Error())
		}
		attempts = append(attempts, attempt)
	}

	rows.Close()

	return attempts, nil
}

func (store *DBStore) GetResult(task string, user string) ([]*models.Result, *models.Error) {
	var results []*models.Result
	var args []interface{}

	selectStr := `SELECT id, userName, taskName, uploadDate, status, percent, 
						 copiedFrom, copiedTask, copiedDate, sourceCode, copiedCode
					FROM results WHERE status=3`

	if task != "" {
		selectStr += " AND taskname=$1"
		args = append(args, task)
	}

	if user != "" {
		if task != "" {
			selectStr += " AND username=$2"
		} else {
			selectStr += " AND username=$1"
		}
		args = append(args, user)
	}
	selectStr += " AND percent > " + strconv.Itoa(int(models.BorrowingThreshold)) + " ;"

	fmt.Println(selectStr)

	rows, err := store.DB.Query(store.ctx, selectStr, args...)
	if err != nil {
		fmt.Println(err)
		return results, models.NewError(http.StatusInternalServerError, err.Error())
	}

	for rows.Next() {
		result := &models.Result{}
		copied := &models.AttemptSimplification{}
		err := rows.Scan(&result.ID, &result.User, &result.Task, &result.UploadDate, &result.Status, &copied.PlagiarismPercent,
			&copied.User, &copied.Task, &copied.UploadDate, &result.SourceCode, &copied.SourceCode)
		if err != nil {
			return results, models.NewError(http.StatusInternalServerError, err.Error())
		}
		result.CopiedFrom = append(result.CopiedFrom, copied)
		results = append(results, result)
	}

	rows.Close()

	return results, nil
}

func (store *DBStore) PutHashes(ID uint64, hashSet models.HashSet) *models.Error {
	var gID uint64
	for hash := range hashSet {
		insertQuery := `INSERT INTO hashes (attemptID, Hash) VALUES ($1, $2) RETURNING attemptID`
		rows := store.DB.QueryRow(store.ctx, insertQuery,
			ID, hash)

		err := rows.Scan(&gID)
		if err != nil || ID != gID {
			fmt.Println(err)
			return models.NewError(http.StatusInternalServerError, err.Error())
		}
	}

	updateQuery := `UPDATE attempts SET status=$1 WHERE ID=$2`
	_, err := store.DB.Exec(store.ctx, updateQuery,
		3, ID)

	if err != nil {
		return models.NewError(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func (store *DBStore) GetHashes(ID uint64) (*models.HashSet, *models.Error) {

	result := make(models.HashSet)

	selectStr := `SELECT hash FROM hashes
						WHERE attemptID=$1;`

	rows, err := store.DB.Query(store.ctx, selectStr, ID)
	defer rows.Close()
	if err != nil {
		return &result, models.NewError(http.StatusInternalServerError, err.Error())
	}

	for rows.Next() {
		var hash models.Hash

		err := rows.Scan(&hash)
		if err != nil {
			return &result, models.NewError(http.StatusInternalServerError, err.Error())
		}
		result[hash] = struct{}{}
	}

	return &result, nil
}

func (store *DBStore) GetSimilarHashes(attempt *models.Attempt) ([]*models.HashObject, *models.Error) {
	var results []*models.HashObject

	selectStr := `SELECT a.ID FROM attempts a
						JOIN tasks t on a.taskID = t.ID
						WHERE t.taskName=$1;`

	rows, err := store.DB.Query(store.ctx, selectStr, attempt.Task)
	defer rows.Close()
	if err != nil {
		fmt.Println(err)
		return results, models.NewError(http.StatusInternalServerError, err.Error())
	}

	for rows.Next() {
		var attemptID uint64
		err := rows.Scan(&attemptID)
		if err != nil {
			return results, models.NewError(http.StatusInternalServerError, err.Error())
		}
		if attemptID == attempt.ID {
			continue
		}
		hs, e := store.GetHashes(attemptID)
		if e != nil {
			return results, models.NewError(http.StatusInternalServerError, e.Error())
		}
		result := &models.HashObject{
			ID:  attemptID,
			Set: hs,
		}
		results = append(results, result)
	}

	return results, nil
}

func (store *DBStore) PutHashes2(ID uint64, hashSet models.HashSet) *models.Error {
	var gID uint64
	var hashes []models.Hash

	for k := range hashSet {
		hashes = append(hashes, k)
	}

	insertQuery := `INSERT INTO hashes2 (attemptID, Hash) VALUES ($1, $2) RETURNING attemptID`
	rows := store.DB.QueryRow(store.ctx, insertQuery,
		ID, pq.Array(hashes))

	err := rows.Scan(&gID)
	if err != nil || ID != gID {
		fmt.Println(err)
		return models.NewError(http.StatusInternalServerError, err.Error())
	}

	updateQuery := `UPDATE attempts SET status=$1 WHERE ID=$2`
	_, err = store.DB.Exec(store.ctx, updateQuery,
		3, ID)

	if err != nil {
		return models.NewError(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func (store *DBStore) GetHashes2(ID uint64) (*models.HashSet, *models.Error) {

	result := make(models.HashSet)

	selectStr := `SELECT hash FROM hashes2
						WHERE attemptID=$1;`
	var hashes []int64

	row := store.DB.QueryRow(store.ctx, selectStr, ID)

	err := row.Scan(pq.Array(&hashes))

	if err != nil {
		return &result, models.NewError(http.StatusInternalServerError, err.Error())
	}

	for _, hash := range hashes {
		result[models.Hash(hash)] = struct{}{}
	}

	return &result, nil
}

func (store *DBStore) GetSimilarHashes2(attempt *models.Attempt) ([]*models.HashObject, *models.Error) {
	var results []*models.HashObject

	selectStr := `SELECT a.ID FROM attempts a
						JOIN tasks t on a.taskID = t.ID
						JOIN users u on a.userID = u.ID
						WHERE t.taskName=$1 AND u.userName!=$2;`

	rows, err := store.DB.Query(store.ctx, selectStr, attempt.Task, attempt.User)

	defer rows.Close()
	if err != nil {
		fmt.Println(err)
		return results, models.NewError(http.StatusInternalServerError, err.Error())
	}

	for rows.Next() {
		var attemptID uint64
		err := rows.Scan(&attemptID)
		if err != nil {
			fmt.Println("222")
			return results, models.NewError(http.StatusInternalServerError, err.Error())
		}
		if attemptID == attempt.ID {
			continue
		}
		hs, e := store.GetHashes2(attemptID)
		if e != nil {
			fmt.Println("111")
			return results, models.NewError(http.StatusInternalServerError, e.Error())
		}
		result := &models.HashObject{
			ID:  attemptID,
			Set: hs,
		}
		results = append(results, result)
	}

	return results, nil
}

func (store *DBStore) PutBorrowing(borrowing *models.Borrowing) *models.Error {
	//fmt.Println(borrowing)
	var ID uint64

	insertQuery := `INSERT INTO borrowings (attemptID, copiedFrom, plagiarismPercent) VALUES ($1, $2, $3) RETURNING attemptID`
	rows := store.DB.QueryRow(store.ctx, insertQuery,
		borrowing.AttemptID, borrowing.CopiedFrom, borrowing.Percent)

	err := rows.Scan(&ID)
	if err != nil || ID != borrowing.AttemptID {
		fmt.Println(err)
		return models.NewError(http.StatusInternalServerError, err.Error())
	}

	return nil
}

//func ( store * DBStore ) PutUsers ( user * models . User ) {
//func ( store * DBStore ) PutUsers ( user *
//	 ( store * DBStore ) PutUsers ( user * models
//	   store * DBStore ) PutUsers ( user * models .
//			 * DBStore ) PutUsers ( user * models . User
//func ( store * DBStore ) PutUsers ( user * models . User )
//func ( store * DBStore ) PutUsers ( user * models . User ) {
