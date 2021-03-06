// +build test

package main

import (
	"flag"
	"fmt"
	"github.com/iris-contrib/httpexpect/v2"
	"github.com/kataras/iris/v12"
	"net/http"
	"os"
	"testing"

	"github.com/kataras/iris/v12/httptest"
	"github.com/snowlyg/IrisAdminApi/config"
	"github.com/snowlyg/IrisAdminApi/models"
	"github.com/snowlyg/IrisAdminApi/seeder"
	"github.com/snowlyg/IrisAdminApi/web_server"
)

var (
	app   *iris.Application
	token string
)

//单元测试基境
func TestMain(m *testing.M) {
	s := web_server.NewServer(AssetFile(), Asset, AssetNames) // 初始化app
	s.NewApp()
	app = s.App

	seeder.Run()

	flag.Parse()
	exitCode := m.Run()

	models.DropTables() // 删除测试数据表，保持测试环境
	os.Exit(exitCode)
}

func getHttpExpect(t *testing.T) *httpexpect.Expect {
	return httptest.New(t, app, httptest.Configuration{Debug: true, URL: "http://127.0.0.1:8085/v1/admin/"})
}

// 单元测试 login 方法
func login(t *testing.T, Object interface{}, StatusCode int, Code int, Msg string) (e *httpexpect.Expect) {
	e = getHttpExpect(t)
	e.POST("login").WithJSON(Object).
		Expect().Status(StatusCode).
		JSON().Object().Values().Contains(Code, Msg)

	return
}

// 单元测试 create 方法
func create(t *testing.T, url string, Object interface{}, StatusCode int, Code int, Msg string) (e *httpexpect.Expect) {
	e = getHttpExpect(t)
	ob := e.POST(url).WithHeader("Authorization", "Bearer "+GetOauthToken(e)).WithJSON(Object).
		Expect().Status(StatusCode).JSON().Object()
	ob.Value("code").Equal(Code)
	ob.Value("message").Equal(Msg)

	return
}

// 单元测试 update 方法
func update(t *testing.T, url string, Object interface{}, StatusCode int, Code int, Msg string) (e *httpexpect.Expect) {
	e = getHttpExpect(t)
	ob := e.PUT(url).WithHeader("Authorization", "Bearer "+GetOauthToken(e)).WithJSON(Object).
		Expect().Status(StatusCode).JSON().Object()
	ob.Value("code").Equal(Code)
	ob.Value("message").Equal(Msg)

	return
}

// 单元测试 getOne 方法
func getOne(t *testing.T, url string, StatusCode int, Code int, Msg string) (e *httpexpect.Expect) {
	e = getHttpExpect(t)
	e.GET(url).WithHeader("Authorization", "Bearer "+GetOauthToken(e)).
		Expect().Status(StatusCode).
		JSON().Object().Values().Contains(Code, Msg)
	return
}

// 单元测试 getOnAuth 方法
func getOnAuth(t *testing.T, url string, StatusCode int, Code int, Msg string) (e *httpexpect.Expect) {
	e = getHttpExpect(t)
	e.GET(url).
		Expect().Status(StatusCode).
		JSON().Object().Values().Contains(Code, Msg)
	return
}

// 单元测试 bImport 方法
func bImport(t *testing.T, url string, StatusCode int, Code int, Msg string, _ map[string]interface{}) (e *httpexpect.Expect) {
	e = getHttpExpect(t)
	e.POST(url).WithHeader("Authorization", "Bearer "+GetOauthToken(e)).
		WithMultipart().
		WithFile("file", "permissions.xlsx").
		Expect().Status(StatusCode).
		JSON().Object().Values().Contains(Code, Msg)

	return
}

// 单元测试 getMore 方法
func getMore(t *testing.T, url string, StatusCode int, Code int, Msg string) (e *httpexpect.Expect) {
	e = getHttpExpect(t)
	e.GET(url).WithHeader("Authorization", "Bearer "+GetOauthToken(e)).
		Expect().Status(StatusCode).
		JSON().Object().Values().Contains(Code, Msg)

	return
}

// 单元测试 delete 方法
func delete(t *testing.T, url string, StatusCode int, Code int, Msg string) (e *httpexpect.Expect) {
	e = getHttpExpect(t)
	e.DELETE(url).WithHeader("Authorization", "Bearer "+GetOauthToken(e)).
		Expect().Status(StatusCode).
		JSON().Object().Values().Contains(Code, Msg)
	return
}

func CreateRole(name, disName, dec string) *models.Role {

	role, err := models.GetRoleByName(name)
	if err == nil && role.ID == 0 {
		role = &models.Role{
			Name:        name,
			DisplayName: disName,
			Description: dec,
		}
		err = role.CreateRole()
		if err != nil {
			fmt.Println(fmt.Sprintf("%+v 角色创建失败: %s", role, err.Error()))
		}
	}

	return role

}

func CreateUser() *models.User {

	user, err := models.GetUserByUsername("TUsername")
	if err == nil && user.ID == 0 {
		user = &models.User{
			Username: "TUsername",
			Password: "TPassword",
			Name:     "TName",
			RoleIds:  []uint{},
		}
		_ = user.CreateUser()
		return user
	} else {
		return user
	}
}

func GetOauthToken(e *httpexpect.Expect) string {
	if len(token) > 0 {
		return token
	}

	oj := map[string]string{
		"username": config.Config.Admin.UserName,
		"password": config.Config.Admin.Pwd,
	}
	r := e.POST("login").WithJSON(oj).
		Expect().
		Status(http.StatusOK).JSON().Object()

	token = r.Value("data").Object().Value("token").String().Raw()

	return token
}
