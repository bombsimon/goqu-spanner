package main

/*
Start the Spanner emulator with Docker:

	docker run --rm -it -p 9020:9020 -p 9010:9010 \
	  gcr.io/cloud-spanner-emulator/emulator:1.1.1

This example will create the instance and database. The example itself will then
populate the database and perform some queries.
*/

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/doug-martin/goqu/v9"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	dbadmin "cloud.google.com/go/spanner/admin/database/apiv1"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	databasepb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
	instancepb "google.golang.org/genproto/googleapis/spanner/admin/instance/v1"

	// Load our custom dialect.
	"github.com/bombsimon/goquspanner"

	_ "github.com/googleapis/go-sql-spanner"
)

const (
	emulatorProjectName  = "emulator-project"
	emulatorInstanceName = "emulator-instance"
	emulatorDBName       = "emulator"
	emulatorProjectPath  = "projects/" + emulatorProjectName
	emulatorInstancePath = emulatorProjectPath + "/instances/" + emulatorInstanceName
	emulatorDBPath       = emulatorInstancePath + "/databases/" + emulatorDBName
)

type User struct {
	UserID int64
	Name   string
	Email  sql.NullString
}

type Role struct {
	RoleID int64
	Name   string
}

type UserRoles struct {
	UserID int64
	RoleID int64
}

func main() {
	createInstanceAndDatabase()

	spannerDB, err := sql.Open(goquspanner.DialectName, emulatorDBPath)
	if err != nil {
		panic(err)
	}

	dialect := goqu.Dialect(goquspanner.DialectName)
	db := dialect.DB(spannerDB)

	insert(db)
	selects(db)
}

func insert(db *goqu.Database) {
	insertRows(db)
	insertColsVals(db)
	insertStruct(db)
	insertRoles(db)
}

func insertRows(db *goqu.Database) {
	log.Println("Inserting with goqu.Record")

	ds := db.Insert("Users").Rows([]goqu.Record{
		{"UserID": 1, "Name": "Jane", "Email": nil},
		{"UserID": 2, "Name": "John", "Email": "john@doe.com"},
	},
	).Executor()

	if _, err := ds.Exec(); err != nil {
		panic(err)
	}
}

func insertColsVals(db *goqu.Database) {
	log.Println("Inserting with goqu.Cols and goqu.Vals")

	ds := db.Insert("Users").
		Cols("UserID", "Name", "Email").
		Vals(
			goqu.Vals{3, "Anna", nil},
			goqu.Vals{4, "Adam", "adam@eden.com"},
		).Executor()

	if _, err := ds.Exec(); err != nil {
		panic(err)
	}
}

func insertStruct(db *goqu.Database) {
	log.Println("Inserting with struct")

	users := []User{
		{
			UserID: 5,
			Name:   "Jessica",
			Email:  sql.NullString{String: "jessica@rbt.com", Valid: true},
		},
		{
			UserID: 6,
			Name:   "James",
		},
	}

	ds := db.Insert("Users").Rows(users).Executor()
	if _, err := ds.Exec(); err != nil {
		panic(err)
	}
}

func insertRoles(db *goqu.Database) {
	log.Println("Inserting roles")

	ds := db.Insert("Roles").Rows([]goqu.Record{
		{"RoleId": 1, "Name": "User"},
		{"RoleId": 2, "Name": "Supporter"},
		{"RoleId": 3, "Name": "Admin"},
		{"RoleId": 4, "Name": "Super User"},
	},
	).Executor()

	if _, err := ds.Exec(); err != nil {
		panic(err)
	}

	log.Println("Mapping roles to users")

	ds = db.Insert("UserRoles").Rows([]goqu.Record{
		{"UserID": 1, "RoleID": 1},
		{"UserID": 1, "RoleID": 2},
		{"UserID": 2, "RoleID": 1},
		{"UserID": 3, "RoleID": 3},
		{"UserID": 4, "RoleID": 1},
		{"UserID": 4, "RoleID": 2},
		{"UserID": 4, "RoleID": 3},
		{"UserID": 4, "RoleID": 4},
	},
	).Executor()

	if _, err := ds.Exec(); err != nil {
		panic(err)
	}
}

func selects(db *goqu.Database) {
	simpleSelect(db)
	joinSelect(db)
}

func simpleSelect(db *goqu.Database) {
	ds := db.From("Users").
		Select(goqu.Star()).
		Where(
			goqu.I("UserID").Lt(6),
			goqu.I("Email").IsNotNull(),
		)

	sqlStr, _, _ := ds.ToSQL()
	log.Println("Selecting simple:")
	log.Println("  ", sqlStr)

	var users []User
	if err := ds.
		ScanStructs(&users); err != nil {
		panic(err)
	}

	log.Printf("%-10s | %-20s | %-s\n", "ID", "NAME", "EMAIL")
	for _, u := range users {
		log.Printf("%-10d | %-20s | %-s\n", u.UserID, u.Name, u.Email.String)
	}

	fmt.Println("")
}

func joinSelect(db *goqu.Database) {
	ds := db.From(goqu.T("Users").As("u")).
		Join(
			goqu.T("UserRoles").As("ur"),
			goqu.On(goqu.Ex{
				"u.UserID": goqu.I("ur.UserID"),
			}),
		).
		Join(
			goqu.T("Roles").As("r"),
			goqu.On(goqu.Ex{
				"ur.RoleID": goqu.I("r.RoleID"),
			}),
		).
		Select(
			goqu.I("u.Name").As("UserName"),
			goqu.I("u.Email"),
			goqu.I("r.Name").As("RoleName"),
		).
		Where(
			goqu.I("u.Email").IsNotNull(),
		)

	sqlStr, _, _ := ds.ToSQL()
	log.Println("Selecting with join:")
	log.Println("  ", sqlStr)

	type UserRole struct {
		Name     string         `db:"UserName"`
		Email    sql.NullString `db:"Email"`
		RoleName string         `db:"RoleName"`
	}

	var userRoles []UserRole
	if err := ds.
		ScanStructs(&userRoles); err != nil {
		panic(err)
	}

	log.Printf("%-20s | %-20s | %-s\n", "NAME", "EMAIL", "ROLE")
	for _, u := range userRoles {
		log.Printf("%-20s | %-20s | %-s\n", u.Name, u.Email.String, u.RoleName)
	}

	fmt.Println("")
}

func createInstanceAndDatabase() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	os.Setenv("SPANNER_EMULATOR_HOST", "localhost:9010")

	opts := []option.ClientOption{
		option.WithEndpoint("localhost:9010"),
		option.WithGRPCDialOption(grpc.WithInsecure()),
		option.WithoutAuthentication(),
	}

	log.Printf("Creating Spanner instance admin client")
	instanceAdminClient, err := instance.NewInstanceAdminClient(ctx, opts...)
	if err != nil {
		panic(err)
	}

	log.Printf("Creating Spanner database admin client")
	databaseAdminClient, err := dbadmin.NewDatabaseAdminClient(ctx, opts...)
	if err != nil {
		panic(err)
	}

	log.Printf("Deleting instance %s (ignoring non existing)", emulatorInstanceName)
	_ = instanceAdminClient.DeleteInstance(ctx, &instancepb.DeleteInstanceRequest{
		Name: emulatorInstancePath,
	})

	log.Printf("Creating instance %s", emulatorInstanceName)
	if _, err := instanceAdminClient.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     emulatorProjectPath,
		InstanceId: emulatorInstanceName,
		Instance: &instancepb.Instance{
			Config:      emulatorProjectName,
			DisplayName: emulatorInstanceName,
			Name:        emulatorInstancePath,
		},
	}); err != nil {
		panic(err)
	}

	log.Printf("Creating database %s", emulatorDBName)
	if _, err := databaseAdminClient.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          emulatorInstancePath,
		CreateStatement: fmt.Sprintf("CREATE DATABASE %s", emulatorDBName),
		ExtraStatements: []string{
			`CREATE TABLE Users (
				UserID INT64 NOT NULL,
				Name   STRING(128) NOT NULL,
				Email  STRING(128),
			) PRIMARY KEY(UserID)`,
			`CREATE UNIQUE INDEX UsersName ON Users(Name)`,
			`CREATE TABLE Roles (
			    RoleID INT64 NOT NULL,
			    Name   STRING(128) NOT NULL,
			) PRIMARY KEY(RoleID)`,
			`CREATE UNIQUE INDEX RoleName ON Roles(Name)`,
			`CREATE TABLE UserRoles (
			    UserID INT64 NOT NULL,
			    RoleID INT64 NOT NULL,
			) PRIMARY KEY (UserID, RoleID)`,
			`CREATE UNIQUE INDEX UserRole ON UserRoles(UserID, RoleID)`,
		},
	}); err != nil {
		panic(err)
	}
}
