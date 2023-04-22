package main

import (
	"context"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
	"io"
	"strconv"
	"time"
)

type UserRepositoryInterface interface {
	SaveUser(clientUserId string, student *Student) error
	GetStudent(clientUserId string) *Student
	GetClientUserIds(studentId uint) []string
	Commit() error
}

type UserRepository struct {
	out   io.Writer
	redis redis.UniversalClient
}

const UserExpiration = time.Hour * 24 * 30 * 7 // 7 months, 210 days

const RedisBackgroundSaveInProgress = "ERR Background save already in progress"

func (repository *UserRepository) SaveUser(clientUserId string, student *Student) (err error) {
	previousStudent := repository.GetStudent(clientUserId)

	studentSerialized, err := proto.Marshal(student)
	ctx := context.Background()
	if err == nil {
		pipe := repository.redis.Pipeline()
		if previousStudent.Id != student.Id {
			pipe.Del(ctx, clientUserId)
			pipe.SRem(ctx, student.GetIdString(), clientUserId)
		}

		if student.Id != 0 {
			pipe.Set(ctx, clientUserId, studentSerialized, UserExpiration)
			pipe.SAdd(ctx, student.GetIdString(), clientUserId)
			pipe.Expire(ctx, student.GetIdString(), UserExpiration)
		}

		_, err = pipe.Exec(ctx)
	}

	return err
}

func (repository *UserRepository) Commit() error {
	err := repository.redis.BgSave(context.Background()).Err()
	if err != nil && err.Error() == RedisBackgroundSaveInProgress {
		err = nil
	}

	return err
}

func (repository *UserRepository) GetStudent(clientUserId string) *Student {
	ctx := context.Background()
	studentSerialized, _ := repository.redis.GetEx(ctx, clientUserId, UserExpiration).Bytes()

	student := &Student{}
	if len(studentSerialized) > 0 {
		_ = proto.Unmarshal(studentSerialized, student)
	}

	if student.Id == 0 {
		repository.redis.Expire(ctx, student.GetIdString(), UserExpiration)
	}

	return student
}

func (repository *UserRepository) GetClientUserIds(studentId uint) (ids []string) {
	if studentId != 0 {
		ids = repository.redis.SMembers(
			context.Background(),
			strconv.FormatUint(uint64(studentId), 10),
		).Val()
	}

	return ids
}
