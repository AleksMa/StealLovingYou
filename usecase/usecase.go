package usecase

import (
	"fmt"
	"github.com/AleksMa/StealLovingYou/checking"
	"github.com/AleksMa/StealLovingYou/lex_utils"
	"github.com/AleksMa/StealLovingYou/models"
	"github.com/AleksMa/StealLovingYou/repository"
	"sort"
	"time"
)

type UseCase interface {
	PutUser(user *models.User) ([]*models.User, *models.Error)
	GetUserByUsername(nickname string) (models.User, *models.Error)
	UpdateUser(user *models.User) ([]*models.User, *models.Error)
	DeleteUserByUsername(nickname string) (models.User, *models.Error)

	PutTask(task *models.Task) ([]*models.Task, *models.Error)
	GetTaskByTaskname(taskname string) (models.Task, *models.Error)
	UpdateTask(task *models.Task) ([]*models.Task, *models.Error)
	DeleteTaskByTaskname(taskname string) (models.Task, *models.Error)

	PutAttempt(attempt *models.Attempt) *models.Error

	GetStatus() (models.Status, error)
	RemoveAllData() error

	GetAttempt(task string, user string) ([]*models.Attempt, *models.Error)
	GetResult(task string, user string) ([]*models.Result, *models.Error)

	PutHashes(attempt *models.Attempt) (*models.HashSet, *models.Error)
}

type useCase struct {
	repository repository.Repo
	timing     *int64
	count      *int64
}

func NewUseCase(repo repository.Repo, t, c *int64) UseCase {
	return &useCase{
		repository: repo,
		timing:     t,
		count:      c,
	}
}

func (u *useCase) GetStatus() (models.Status, error) {
	return u.repository.GetStatus()
}

func (u *useCase) RemoveAllData() error {
	return u.repository.ReloadDB()
}

func (u *useCase) PutUser(user *models.User) ([]*models.User, *models.Error) {

	if err := user.Validate(); err != nil {
		return nil, err
	}

	users, _ := u.repository.GetUsers(user)
	if users != nil && len(users) != 0 {
		fmt.Println("DUP: ", users)
		return users, nil
	}

	_, err := u.repository.PutUser(user)
	return nil, err
}

func (u *useCase) GetUserByUsername(nickname string) (models.User, *models.Error) {
	return u.repository.GetUserByUsername(nickname)
}

func (u *useCase) UpdateUser(user *models.User) ([]*models.User, *models.Error) {

	if err := user.Validate(); err != nil {
		return nil, err
	}

	users, _ := u.repository.GetUsers(user)
	if users == nil || len(users) == 0 {
		fmt.Println("NOT EXIST: ", users)
		return users, models.NewError(404, "no users")
	} else if len(users) > 1 {
		fmt.Println("Unexpected StudentID: ", users)
		return users, nil
	}

	_, err := u.repository.UpdateUser(user)
	return []*models.User{user}, err
}

func (u *useCase) DeleteUserByUsername(nickname string) (models.User, *models.Error) {
	return u.repository.DeleteUserByUsername(nickname)
}

//func (u *useCase) GetUserByID(id int64) (models.User, *models.Error) {
//	return u.repository.GetUserByID(id)
//}

//func (u *useCase) ChangeUser(userUpd *models.UpdateUserFields, nickname string) (models.User, *models.Error) {
//	tempUser, err := u.GetUserByUsername(nickname)
//	if err != nil {
//		return tempUser, err
//	}
//
//	if userUpd.Email != nil {
//		tempUser.Email = *userUpd.Email
//	}
//	if userUpd.Fullname != nil {
//		tempUser.Fullname = *userUpd.Fullname
//	}
//	if userUpd.About != nil {
//		tempUser.About = *userUpd.About
//	}
//
//	err = u.repository.ChangeUser(&tempUser)
//	fmt.Println(*userUpd)
//	fmt.Println(tempUser)
//	return tempUser, err
//}

func (u *useCase) PutTask(task *models.Task) ([]*models.Task, *models.Error) {
	fmt.Println(task)

	if err := task.Validate(); err != nil {
		return nil, err
	}

	tasks, _ := u.repository.GetTasks(task)
	if tasks != nil && len(tasks) != 0 {
		fmt.Println("DUP: ", tasks)
		return tasks, nil
	}

	_, err := u.repository.PutTask(task)
	return nil, err
}

func (u *useCase) GetTaskByTaskname(taskname string) (models.Task, *models.Error) {
	return u.repository.GetTaskByTaskname(taskname)
}

func (u *useCase) UpdateTask(task *models.Task) ([]*models.Task, *models.Error) {

	if err := task.Validate(); err != nil {
		return nil, err
	}

	tasks, _ := u.repository.GetTasks(task)
	if tasks == nil || len(tasks) == 0 {
		fmt.Println("NOT EXIST: ", task)
		return tasks, models.NewError(404, "no tasks")
	}

	_, err := u.repository.UpdateTask(task)
	return nil, err
}

func (u *useCase) DeleteTaskByTaskname(taskname string) (models.Task, *models.Error) {
	return u.repository.DeleteTaskByTaskname(taskname)
}

func (u *useCase) PutAttempt(attempt *models.Attempt) *models.Error {
	ID, err := u.repository.PutAttempt(attempt)
	attempt.ID = ID
	if err == nil {
		go func() {
			t := time.Now()
			u.PlagiarismCheck(attempt)
			fmt.Println("Checking: ", time.Now().Sub(t))
			*(u.timing) += (time.Now().Sub(t)).Nanoseconds()
			*(u.count)++
			fmt.Println("TIMING:", *(u.timing) / *(u.count))
			fmt.Println("COUNT:", *(u.count))
			fmt.Println()
		}()
	}
	return err
}

func (u *useCase) GetAttempt(task string, user string) ([]*models.Attempt, *models.Error) {
	return u.repository.GetAttempt(task, user)
}

func isResultsEqual(res1, res2 *models.Result) bool {
	return res1.ID == res2.ID
}

func (u *useCase) GetResult(task string, user string) ([]*models.Result, *models.Error) {
	results, err := u.repository.GetResult(task, user)
	if err != nil {
		return results, err
	}

	if len(results) == 0 {
		return results, nil
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})

	j := 0
	for i := 1; i < len(results); i++ {
		if isResultsEqual(results[j], results[i]) {
			results[j].CopiedFrom = append(results[j].CopiedFrom, results[i].CopiedFrom...)
			continue
		}
		j++
		results[j] = results[i]
	}

	return results[:j+1], nil
}

func (u *useCase) PutHashes(attempt *models.Attempt) (*models.HashSet, *models.Error) {
	t := time.Now()
	hashSet := lex_utils.FullAnalise(attempt.SourceCode)
	fmt.Println("Storing: ", time.Now().Sub(t))
	fmt.Println()
	//time.Sleep(2 * time.Minute)
	err := u.repository.PutHashes2(attempt.ID, hashSet)
	return &hashSet, err
}

func (u *useCase) GetSimilarHashes(attempt *models.Attempt) ([]*models.HashObject, *models.Error) {
	return u.repository.GetSimilarHashes2(attempt)
}

func (u *useCase) PlagiarismCheck(attempt *models.Attempt) {
	var (
		copied []*models.Borrowing
	)

	hashSet, e := u.PutHashes(attempt)
	if e != nil {
		fmt.Println(e)
		return
	}

	if len(*hashSet) == 0 {
		fmt.Println("Too small program")
		return
	}

	similarHashes, e := u.GetSimilarHashes(attempt)
	if e != nil {
		fmt.Println(e)
		return
	}

	curObject := &models.HashObject{
		ID:  attempt.ID,
		Set: hashSet,
	}

	for _, hashObject := range similarHashes {
		res := checking.CheckObjects(curObject, hashObject)
		if res >= models.BorrowingThreshold {
			copied = append(copied, &models.Borrowing{
				AttemptID:  curObject.ID,
				CopiedFrom: hashObject.ID,
				Percent:    res,
			})
		}
	}

	for _, borrowing := range copied {
		e = u.repository.PutBorrowing(borrowing)
		if e != nil {
			fmt.Println(e)
			return
		}
	}
}

//func isResultsEqual(res1, res2 *models.Result) bool {
//	return res1.User == res2.User &&
//		res1.Task == res2.Task &&
//		res1.UploadDate == res2.UploadDate
//}

//sort.Slice(results, func(i, j int) bool {
//	if results[i].User != results[j].User {
//		return results[i].User < results[j].User
//	}
//	if results[i].Task != results[j].Task {
//		return results[i].Task < results[j].Task
//	}
//	return results[i].UploadDate.Unix() < results[j].UploadDate.Unix()
//})
