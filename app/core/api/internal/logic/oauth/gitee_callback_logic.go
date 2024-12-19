package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/yitter/idgenerator-go/idgen"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/internal/logic/user"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/model"
)

type GiteeCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}
type GiteeUser struct {
	AvatarURL         string      `json:"avatar_url"`
	Bio               string      `json:"bio"`
	Blog              string      `json:"blog"`
	CreatedAt         time.Time   `json:"created_at"`
	Email             string      `json:"email"`
	EventsURL         string      `json:"events_url"`
	Followers         int         `json:"followers"`
	FollowersURL      string      `json:"followers_url"`
	Following         int         `json:"following"`
	FollowingURL      string      `json:"following_url"`
	GistsURL          string      `json:"gists_url"`
	HTMLURL           string      `json:"html_url"`
	ID                int         `json:"id"`
	Login             string      `json:"login"`
	Name              string      `json:"name"`
	OrganizationsURL  string      `json:"organizations_url"`
	PublicGists       int         `json:"public_gists"`
	PublicRepos       int         `json:"public_repos"`
	ReceivedEventsURL string      `json:"received_events_url"`
	Remark            string      `json:"remark"`
	ReposURL          string      `json:"repos_url"`
	Stared            int         `json:"stared"`
	StarredURL        string      `json:"starred_url"`
	SubscriptionsURL  string      `json:"subscriptions_url"`
	Type              string      `json:"type"`
	UpdatedAt         time.Time   `json:"updated_at"`
	URL               string      `json:"url"`
	Watched           int         `json:"watched"`
	Weibo             interface{} `json:"weibo"`
}
type Token struct {
	AccessToken string `json:"access_token"`
}

var Script = `
        <script>
        window.opener.postMessage('%s', '%s');
        window.close();
        </script>
        `

func NewGiteeCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GiteeCallbackLogic {
	return &GiteeCallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GiteeCallbackLogic) GiteeCallback(w http.ResponseWriter, r *http.Request, req *types.OAuthCallbackRequest) (string, error) {
	// 获取 token
	tokenAuthUrl := l.GetGiteeTokenAuthUrl(req.Code)
	token, err := l.GetGiteeToken(tokenAuthUrl)
	if err != nil {
		return "", err
	}
	if token == nil {
		return "", errors.New("get gitee token failed")
	}

	// 获取用户信息
	userInfo, err := l.GetGiteeUserInfo(token)
	if err != nil {
		return "", err
	}
	if userInfo == nil {
		return "", errors.New("get gitee user info failed")
	}

	var giteeUser GiteeUser
	marshal, err := json.Marshal(userInfo)
	if err != nil {
		return "", err
	}
	if err = json.Unmarshal(marshal, &giteeUser); err != nil {
		return "", err
	}
	Id := strconv.Itoa(giteeUser.ID)

	tx := l.svcCtx.DB.Begin()

	userSocial := l.svcCtx.DB.ScaAuthUserSocial
	socialUser, err := tx.ScaAuthUserSocial.Where(userSocial.OpenID.Eq(Id), userSocial.Source.Eq(constant.OAuthSourceGitee)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	if socialUser == nil {
		// 创建用户
		uid := idgen.NextId()
		uidStr := strconv.FormatInt(uid, 10)
		addUser := &model.ScaAuthUser{
			UID:      uidStr,
			Avatar:   giteeUser.AvatarURL,
			Username: giteeUser.Login,
			Nickname: giteeUser.Name,
			Blog:     giteeUser.Blog,
			Email:    giteeUser.Email,
			Gender:   constant.Male,
		}
		err = tx.ScaAuthUser.Create(addUser)
		if err != nil {
			_ = tx.Rollback()
			return "", err
		}
		gitee := constant.OAuthSourceGitee
		newSocialUser := &model.ScaAuthUserSocial{
			UserID: uidStr,
			OpenID: Id,
			Source: gitee,
		}
		err = tx.ScaAuthUserSocial.Create(newSocialUser)
		if err != nil {
			_ = tx.Rollback()
			return "", err
		}

		if res, err := l.svcCtx.CasbinEnforcer.AddRoleForUser(uidStr, constant.User); !res || err != nil {
			_ = tx.Rollback()
			return "", err
		}
		data, err := HandleOauthLoginResponse(addUser, l.svcCtx, r, w, l.ctx)
		if err != nil {
			_ = tx.Rollback()
			return "", err
		}
		if err = tx.Commit(); err != nil {
			return "", err
		}
		return data, nil
	} else {
		authUser := l.svcCtx.DB.ScaAuthUser

		authUserInfo, err := tx.ScaAuthUser.Where(authUser.UID.Eq(socialUser.UserID)).First()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			_ = tx.Rollback()
			return "", err
		}

		data, err := HandleOauthLoginResponse(authUserInfo, l.svcCtx, r, w, l.ctx)
		if err != nil {
			_ = tx.Rollback()
			return "", err
		}
		if err = tx.Commit(); err != nil {
			return "", err
		}
		return data, nil
	}
}

// HandleOauthLoginResponse 处理登录响应
func HandleOauthLoginResponse(scaAuthUser *model.ScaAuthUser, svcCtx *svc.ServiceContext, r *http.Request, w http.ResponseWriter, ctx context.Context) (string, error) {
	data, err := user.HandleUserLogin(scaAuthUser, svcCtx, true, r, w, ctx)
	if err != nil {
		return "", err
	}
	responseData := response.SuccessWithData(data)
	marshalData, err := json.Marshal(responseData)
	if err != nil {
		return "", err
	}
	formattedScript := fmt.Sprintf(Script, marshalData, svcCtx.Config.Web.URL)
	return formattedScript, nil
}

// GetGiteeTokenAuthUrl 获取Gitee token
func (l *GiteeCallbackLogic) GetGiteeTokenAuthUrl(code string) string {
	clientId := l.svcCtx.Config.OAuth.Gitee.ClientID
	clientSecret := l.svcCtx.Config.OAuth.Gitee.ClientSecret
	redirectURI := l.svcCtx.Config.OAuth.Gitee.RedirectURI
	return fmt.Sprintf(
		"https://gitee.com/oauth/token?grant_type=authorization_code&code=%s&client_id=%s&redirect_uri=%s&client_secret=%s",
		code, clientId, redirectURI, clientSecret,
	)
}

// GetGiteeToken 获取 token
func (l *GiteeCallbackLogic) GetGiteeToken(url string) (*Token, error) {

	// 形成请求
	var req *http.Request
	var err error
	if req, err = http.NewRequest(http.MethodPost, url, nil); err != nil {
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

// GetGiteeUserInfo 获取用户信息
func (l *GiteeCallbackLogic) GetGiteeUserInfo(token *Token) (map[string]interface{}, error) {

	// 形成请求
	var userInfoUrl = "https://gitee.com/api/v5/user" // github用户信息获取接口
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
