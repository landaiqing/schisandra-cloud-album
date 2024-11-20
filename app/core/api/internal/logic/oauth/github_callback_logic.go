package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/yitter/idgenerator-go/idgen"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GithubCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}
type GitHubUser struct {
	AvatarURL         string      `json:"avatar_url"`
	Bio               interface{} `json:"bio"`
	Blog              string      `json:"blog"`
	Company           interface{} `json:"company"`
	CreatedAt         string      `json:"created_at"`
	Email             string      `json:"email"`
	EventsURL         string      `json:"events_url"`
	Followers         int         `json:"followers"`
	FollowersURL      string      `json:"followers_url"`
	Following         int         `json:"following"`
	FollowingURL      string      `json:"following_url"`
	GistsURL          string      `json:"gists_url"`
	GravatarID        string      `json:"gravatar_id"`
	Hireable          interface{} `json:"hireable"`
	HTMLURL           string      `json:"html_url"`
	ID                int         `json:"id"`
	Location          interface{} `json:"location"`
	Login             string      `json:"login"`
	Name              string      `json:"name"`
	NodeID            string      `json:"node_id"`
	NotificationEmail interface{} `json:"notification_email"`
	OrganizationsURL  string      `json:"organizations_url"`
	PublicGists       int         `json:"public_gists"`
	PublicRepos       int         `json:"public_repos"`
	ReceivedEventsURL string      `json:"received_events_url"`
	ReposURL          string      `json:"repos_url"`
	SiteAdmin         bool        `json:"site_admin"`
	StarredURL        string      `json:"starred_url"`
	SubscriptionsURL  string      `json:"subscriptions_url"`
	TwitterUsername   interface{} `json:"twitter_username"`
	Type              string      `json:"type"`
	UpdatedAt         string      `json:"updated_at"`
	URL               string      `json:"url"`
}

func NewGithubCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GithubCallbackLogic {
	return &GithubCallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GithubCallbackLogic) GithubCallback(w http.ResponseWriter, r *http.Request, req *types.OAuthCallbackRequest) error {

	// 获取 token
	tokenAuthUrl := l.GetTokenAuthUrl(req.Code)
	token, err := l.GetToken(tokenAuthUrl)
	if err != nil {
		return err
	}
	if token == nil {

		return nil
	}

	// 获取用户信息
	userInfo, err := l.GetUserInfo(token)
	if err != nil {

		return err
	}

	if userInfo == nil {
		return nil
	}

	// 处理用户信息
	userInfoBytes, err := json.Marshal(userInfo)
	if err != nil {
		return err
	}
	var gitHubUser GitHubUser
	err = json.Unmarshal(userInfoBytes, &gitHubUser)
	if err != nil {
		return err
	}
	Id := strconv.Itoa(gitHubUser.ID)
	tx := l.svcCtx.DB.NewSession()
	defer tx.Close()
	if err = tx.Begin(); err != nil {
		return err
	}
	userSocial := model.ScaAuthUserSocial{
		OpenId:  Id,
		Source:  constant.OAuthSourceGithub,
		Deleted: constant.NotDeleted,
	}
	has, err := tx.Get(&userSocial)
	if err != nil {
		return err
	}

	if !has {
		// 创建用户
		uid := idgen.NextId()
		uidStr := strconv.FormatInt(uid, 10)

		addUser := model.ScaAuthUser{
			UID:      uidStr,
			Avatar:   gitHubUser.AvatarURL,
			Username: gitHubUser.Login,
			Nickname: gitHubUser.Name,
			Blog:     gitHubUser.Blog,
			Email:    gitHubUser.Email,
			Deleted:  constant.NotDeleted,
			Gender:   constant.Male,
		}
		affected, err := tx.Insert(&addUser)
		if err != nil || affected == 0 {
			return err
		}

		socialUser := model.ScaAuthUserSocial{
			UserId: uidStr,
			OpenId: Id,
			Source: constant.OAuthSourceGithub,
		}
		insert, err := tx.Insert(&socialUser)
		if err != nil || insert == 0 {
			return err
		}

		if res, err := l.svcCtx.CasbinEnforcer.AddRoleForUser(uidStr, constant.User); !res || err != nil {
			return err
		}

		if err = HandleOauthLoginResponse(addUser, l.svcCtx, r, w, l.ctx); err != nil {
			return err
		}
	} else {
		user := model.ScaAuthUser{
			UID:     userSocial.UserId,
			Deleted: constant.NotDeleted,
		}
		have, err := tx.Get(&user)
		if err != nil || !have {
			return err
		}

		if err = HandleOauthLoginResponse(user, l.svcCtx, r, w, l.ctx); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// GetTokenAuthUrl 通过code获取token认证url
func (l *GithubCallbackLogic) GetTokenAuthUrl(code string) string {
	clientId := l.svcCtx.Config.OAuth.Github.ClientID
	clientSecret := l.svcCtx.Config.OAuth.Github.ClientSecret
	return fmt.Sprintf(
		"https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s",
		clientId, clientSecret, code,
	)
}

// GetToken 获取 token
func (l *GithubCallbackLogic) GetToken(url string) (*Token, error) {

	// 形成请求
	var req *http.Request
	var err error
	if req, err = http.NewRequest(http.MethodGet, url, nil); err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json")

	// 发送请求并获得响应
	var httpClient = http.Client{}
	var res *http.Response
	if res, err = httpClient.Do(req); err != nil {
		return nil, err
	}

	// 将响应体解析为 token，并返回
	var token Token
	if err = json.NewDecoder(res.Body).Decode(&token); err != nil {
		return nil, err
	}
	return &token, nil
}

// GetUserInfo 获取用户信息
func (l *GithubCallbackLogic) GetUserInfo(token *Token) (map[string]interface{}, error) {

	// 形成请求
	var userInfoUrl = "https://api.github.com/user" // github用户信息获取接口
	var req *http.Request
	var err error
	if req, err = http.NewRequest(http.MethodGet, userInfoUrl, nil); err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token.AccessToken))

	// 发送请求并获取响应
	var client = http.Client{}
	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return nil, err
	}

	// 将响应的数据写入 userInfo 中，并返回
	var userInfo = make(map[string]interface{})
	if err = json.NewDecoder(res.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	return userInfo, nil
}
