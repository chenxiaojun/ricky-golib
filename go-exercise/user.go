package exercise

import (
	"fmt"
)

type User struct {
	*Page

	// userId 表示用户的知乎 ID (用户名)
	userID string
}

func NewUser(link string, userID string) *User {
	if link == "" && !isAnonymous(userID) {
		panic("调用NewUser 的参数不合法")
	}

	return &User{
		Page:   newZhihuPage(link),
		userID: userID,
	}
}

// 返回知乎用户的ID
func (user *User) GetUserID() string {
	if user.userID != "" {
		return user.userID
	}

	doc := user.Doc()
	//user.userID = strip(doc.Find("div.title-section.ellipsis").Find("span.name").Text())
	user.userID = strip(doc.Find("div.ProfileHeader-content").Find("span.ProfileHeader-name").Text())
	return user.userID
}

func (user *User) GetLocation() string {
	return user.getProfile("location")
}

func (user *User) getProfile(cacheKey string) string {
	if user.IsAnonymous() {
		return ""
	}

	if got, ok := user.getStringField(cacheKey); ok {
		return got
	}

	doc := user.Doc()
	// NEED UPDATE
	value, _ := doc.Find(fmt.Sprintf("span.%s", cacheKey)).Attr("title")
	user.setField(cacheKey, value)
	return value
}

func (user *User) IsAnonymous() bool {
	return isAnonymous(user.userID)
}

func isAnonymous(userID string) bool {
	return userID == "匿名用户" || userID == "知乎用户"
}

















