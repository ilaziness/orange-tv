package client

import "github.com/ilaziness/orange-tv/internal/dto"

// ===== User auth (C5) =====

// RegisterRequest is the user registration payload.
type RegisterRequest struct {
	// 邮箱（必填）
	Email string `json:"email" binding:"required,email,max=128"`
	// 密码（5-30 位）
	Password string `json:"password" binding:"required,min=5,max=30"`
	// 昵称（可选，3-15 位；为空时默认取邮箱 @ 前部分）
	Nickname string `json:"nickname" binding:"omitempty,min=3,max=15"`
}

// LoginRequest is the user login payload.
type LoginRequest struct {
	// 邮箱
	Email string `json:"email" binding:"required,email,max=128"`
	// 密码（5-30 位）
	Password string `json:"password" binding:"required,min=5,max=30"`
}

// LoginResponse is returned after successful user login.
type LoginResponse struct {
	// 访问令牌
	AccessToken string `json:"access_token"`
	// 令牌类型（如 Bearer）
	TokenType string `json:"token_type"`
	// 令牌有效期（秒）
	ExpiresIn int `json:"expires_in"`
	// 用户公开资料
	User *Profile `json:"user"`
}

// Profile is the authenticated user public profile.
type Profile struct {
	// 用户ID
	ID uint32 `json:"id"`
	// 字符串形式用户ID
	StrID string `json:"str_id"`
	// 昵称
	Nickname string `json:"nickname"`
	// 邮箱
	Email string `json:"email"`
	// 头像地址
	Avatar string `json:"avatar"`
	// 账号状态（0=禁用，1=正常）
	Status uint8 `json:"status"`
}

// UpdateProfileRequest updates the current user's profile.
type UpdateProfileRequest struct {
	// 昵称（3-15 位，可选）
	Nickname string `json:"nickname" binding:"omitempty,min=3,max=15"`
	// 邮箱（可选）
	Email string `json:"email" binding:"omitempty,email,max=128"`
	// 头像地址（可选）
	Avatar string `json:"avatar" binding:"omitempty,url,max=120"`
}

// ChangePasswordRequest changes the current user's password.
type ChangePasswordRequest struct {
	// 当前密码
	CurrentPassword string `json:"current_password" binding:"required,min=5,max=30"`
	// 新密码（5-30 位）
	NewPassword string `json:"new_password" binding:"required,min=5,max=30"`
}

// LoginHistoryListRequest filters user login history.
type LoginHistoryListRequest struct {
	dto.PaginationRequest
}

// LoginHistoryItem is a single user login log entry.
type LoginHistoryItem struct {
	// 日志ID
	ID uint32 `json:"id"`
	// 登录 IP 地址
	IP string `json:"ip"`
	// User-Agent 信息
	UserAgent string `json:"user_agent"`
	// 登录结果（1=成功，2=失败）
	Status uint8 `json:"status"`
	// 登录时间
	CreatedAt string `json:"created_at"`
}

// ===== Favorites (C6) =====

// FavoriteListRequest filters user favorites.
type FavoriteListRequest struct {
	dto.PaginationRequest
}

// FavoriteItem is the favorite list item.
type FavoriteItem struct {
	// 视频ID
	VideoID uint32 `json:"video_id"`
	// 视频标题
	Title string `json:"title"`
	// 封面地址
	Cover string `json:"cover"`
	// 上映年份
	Year uint32 `json:"year"`
	// 评分
	Rating float64 `json:"rating"`
	// 分类名称
	CategoryName string `json:"category_name"`
	// 收藏时间
	CreatedAt string `json:"created_at"`
}

// FavoriteCheckResult is the favorite check result.
type FavoriteCheckResult struct {
	// 是否已收藏
	Favorited bool `json:"favorited"`
}

// ===== Play history (C6) =====

// HistoryListRequest filters user play history.
type HistoryListRequest struct {
	dto.PaginationRequest
}

// HistoryItem is the play history list item.
type HistoryItem struct {
	// 视频ID
	VideoID uint32 `json:"video_id"`
	// 视频标题
	Title string `json:"title"`
	// 封面地址
	Cover string `json:"cover"`
	// 上映年份
	Year string `json:"year"`
	// 分类名称
	CategoryName string `json:"category_name"`
	// 播放源ID
	PlaySourceID uint32 `json:"play_source_id"`
	// 剧集ID
	EpisodeID uint32 `json:"episode_id"`
	// 播放进度（秒）
	Progress uint32 `json:"progress"`
	// 视频总时长（秒）
	Duration uint32 `json:"duration"`
	// 最近播放时间
	LastPlayedAt string `json:"last_played_at"`
}

// UpsertHistoryRequest upserts play progress.
type UpsertHistoryRequest struct {
	// 视频ID（必填）
	VideoID uint32 `json:"video_id" binding:"required,min=1"`
	// 播放源ID
	PlaySourceID uint32 `json:"play_source_id" binding:"omitempty,min=1"`
	// 剧集ID
	EpisodeID uint32 `json:"episode_id" binding:"omitempty,min=1"`
	// 播放进度（秒）
	Progress uint32 `json:"progress" binding:"omitempty,min=0"`
	// 视频总时长（秒）
	Duration uint32 `json:"duration" binding:"omitempty,min=0"`
}

// ===== Comments (C6) =====

// CommentListRequest filters video comments.
type CommentListRequest struct {
	dto.PaginationRequest
}

// CommentItem is the comment list item.
type CommentItem struct {
	// 评论ID
	ID uint32 `json:"id"`
	// 视频ID
	VideoID uint32 `json:"video_id"`
	// 用户ID
	UserID uint32 `json:"user_id"`
	// 昵称
	Nickname string `json:"nickname"`
	// 用户头像
	Avatar string `json:"avatar"`
	// 父评论ID，0 表示顶级评论
	ParentID uint32 `json:"parent_id"`
	// 评论内容
	Content string `json:"content"`
	// 点赞数
	LikeCount uint32 `json:"like_count"`
	// 踩数
	DislikeCount uint32 `json:"dislike_count"`
	// 我的投票（1=顶，-1=踩，0=未投票）
	MyVote int8 `json:"my_vote"`
	// 回复数
	ReplyCount int `json:"reply_count"`
	// 子回复列表
	Replies []*CommentItem `json:"replies"`
	// 评论时间
	CreatedAt string `json:"created_at"`
}

// CreateCommentRequest creates a comment.
type CreateCommentRequest struct {
	// 视频ID（必填）
	VideoID uint32 `json:"video_id" binding:"required,min=1"`
	// 父评论ID，回复时必填
	ParentID uint32 `json:"parent_id" binding:"omitempty,min=1"`
	// 评论内容（1-200 字）
	Content string `json:"content" binding:"required,min=1,max=200"`
}

// VoteCommentRequest votes on a comment.
type VoteCommentRequest struct {
	// 投票动作（like=顶，dislike=踩，cancel=取消）
	Action string `json:"action" binding:"required,oneof=like dislike cancel"`
}

// VoteCommentResult is returned after a comment vote.
type VoteCommentResult struct {
	// 点赞数
	LikeCount uint32 `json:"like_count"`
	// 踩数
	DislikeCount uint32 `json:"dislike_count"`
	// 我的投票（1=顶，-1=踩，0=未投票）
	MyVote int8 `json:"my_vote"`
}

// ===== Banner (C1) =====

// BannerItem is the client banner item.
type BannerItem struct {
	// 横幅ID
	ID uint32 `json:"id"`
	// 横幅标题
	Title string `json:"title"`
	// 封面地址
	Cover string `json:"cover"`
	// 跳转链接
	Link string `json:"link"`
	// 关联视频ID
	VideoID uint32 `json:"video_id"`
}

// ===== Ratings (C6) =====

// RateVideoRequest is the user video rating payload.
type RateVideoRequest struct {
	// 评分（1-10 分，必填）
	Score float64 `json:"score" binding:"required"`
}

// RatingResult is the rating response containing the user's score and video stats.
type RatingResult struct {
	// 当前用户评分，0 表示未评/未登录
	MyScore float64 `json:"my_score"`
	// 视频平均分
	Rating float64 `json:"rating"`
	// 评分人数
	RatingCount uint32 `json:"rating_count"`
}
