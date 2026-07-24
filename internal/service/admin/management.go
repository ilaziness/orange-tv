package admin

import (
	"context"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/audit"
	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/crypto"
	dto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"go.uber.org/zap"
)

// ManagementService handles dashboard, batch ops, admin/user/user-group CRUD, banners.
type ManagementService interface {
	// A1: Dashboard
	Dashboard(ctx context.Context) (*dto.DashboardResponse, error)

	// A2: Batch video ops
	BatchUpdatePublishStatus(ctx context.Context, req *dto.BatchVideoRequest) (*dto.BatchVideoResponse, error)
	BatchDeleteVideos(ctx context.Context, req *dto.BatchVideoRequest) (*dto.BatchVideoResponse, error)

	// A3: Admin CRUD
	ListAdmins(ctx context.Context, req *dto.AdminListRequest) ([]dto.AdminItem, int, error)
	CreateAdmin(ctx context.Context, req *dto.CreateAdminRequest) (*dto.AdminItem, error)
	UpdateAdmin(ctx context.Context, id int64, req *dto.UpdateAdminRequest) (*dto.AdminItem, error)
	ResetAdminPassword(ctx context.Context, id int64, req *dto.ResetAdminPasswordRequest) error
	DeleteAdmin(ctx context.Context, id int64) error

	// A4: User group CRUD
	ListGroups(ctx context.Context, req *dto.UserGroupListRequest) ([]dto.UserGroupItem, int, error)
	CreateGroup(ctx context.Context, req *dto.CreateUserGroupRequest) (*dto.UserGroupItem, error)
	UpdateGroup(ctx context.Context, id int64, req *dto.UpdateUserGroupRequest) (*dto.UserGroupItem, error)
	DeleteGroup(ctx context.Context, id int64) error

	// A5: Regular user CRUD
	ListUsers(ctx context.Context, req *dto.UserListRequest) ([]dto.UserItem, int, error)
	UpdateUser(ctx context.Context, id int64, req *dto.UpdateUserRequest) (*dto.UserItem, error)
	ResetUserPassword(ctx context.Context, id int64, req *dto.ResetUserPasswordRequest) error
	DeleteUser(ctx context.Context, id int64) error
	ListUserLoginLogs(ctx context.Context, userID int64, offset, limit int) ([]model.UserLoginLogs, int, error)

	// C1: Banner CRUD
	ListBanners(ctx context.Context, offset, limit int) ([]dto.BannerItem, int, error)
	CreateBanner(ctx context.Context, req *dto.CreateBannerRequest) (*dto.BannerItem, error)
	UpdateBanner(ctx context.Context, id int64, req *dto.UpdateBannerRequest) (*dto.BannerItem, error)
	DeleteBanner(ctx context.Context, id int64) error
}

type managementService struct {
	adminRepo repository.AdminRepository
	videoRepo repository.VideoRepository
	userRepo  repository.UserFeatureRepository
	audit     *audit.Recorder
	log       *zap.Logger
}

// NewManagementService creates a ManagementService.
func NewManagementService(
	adminRepo repository.AdminRepository,
	videoRepo repository.VideoRepository,
	userRepo repository.UserFeatureRepository,
	recorder *audit.Recorder,
	log *zap.Logger,
) ManagementService {
	if log == nil {
		log = zap.NewNop()
	}
	return &managementService{
		adminRepo: adminRepo,
		videoRepo: videoRepo,
		userRepo:  userRepo,
		audit:     recorder,
		log:       log,
	}
}

// ===== A1: Dashboard =====

func (s *managementService) Dashboard(ctx context.Context) (*dto.DashboardResponse, error) {
	resp := &dto.DashboardResponse{}
	// CollectRunning is reserved for future collect-engine status; defaults to 0 (idle).
	resp.CollectRunning = 0

	totalVideos, err := s.videoRepo.CountVideos(ctx)
	if err != nil {
		s.log.Error("management: dashboard count videos failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	resp.TotalVideos = int64(totalVideos)

	// Today's zero point in local timezone for "today" filtering
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayVideos, err := s.videoRepo.CountVideosToday(ctx, startOfDay)
	if err != nil {
		s.log.Error("management: dashboard count videos today failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	resp.TodayVideos = int64(todayVideos)

	onlineVideos, err := s.videoRepo.CountVideosByStatus(ctx, constant.PublishStatusOnline)
	if err != nil {
		s.log.Error("management: dashboard count online videos failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	resp.OnlineVideos = int64(onlineVideos)

	offlineVideos, err := s.videoRepo.CountVideosByStatus(ctx, constant.PublishStatusOffline)
	if err != nil {
		s.log.Error("management: dashboard count offline videos failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	resp.OfflineVideos = int64(offlineVideos)

	totalCats, err := s.videoRepo.CountCategories(ctx)
	if err != nil {
		s.log.Error("management: dashboard count categories failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	resp.TotalCategories = int64(totalCats)

	admins, totalAdmins, err := s.adminRepo.ListAdmins(ctx, repository.AdminListFilter{Offset: 0, Limit: 1})
	if err != nil {
		s.log.Error("management: dashboard list admins failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	_ = admins
	resp.TotalAdmins = int64(totalAdmins)

	users, totalUsers, err := s.adminRepo.ListUsers(ctx, repository.UserListFilter{Offset: 0, Limit: 1})
	if err != nil {
		s.log.Error("management: dashboard list users failed", zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	_ = users
	resp.TotalUsers = int64(totalUsers)

	// Online count (approximate: sessions active in last 5 minutes)
	onlineCount, err := s.userRepo.CountOnlineSessions(ctx, time.Now().Add(-5*time.Minute))
	if err != nil {
		s.log.Warn("management: dashboard count online sessions failed", zap.Error(err))
		// non-fatal: return 0
		resp.OnlineCount = 0
	} else {
		resp.OnlineCount = int64(onlineCount)
	}

	// Today PV/UV (use the same local zero point for consistency with startOfDay)
	stats, err := s.userRepo.GetDailyStats(ctx, startOfDay)
	if err != nil {
		s.log.Warn("management: dashboard get daily stats failed", zap.Error(err))
		resp.TodayPV = 0
		resp.TodayUV = 0
	} else if stats != nil {
		resp.TodayPV = int64(stats.PV)
		resp.TodayUV = int64(stats.UV)
	}

	return resp, nil
}

// ===== A2: Batch video ops =====

func (s *managementService) BatchUpdatePublishStatus(ctx context.Context, req *dto.BatchVideoRequest) (*dto.BatchVideoResponse, error) {
	status := constant.PublishStatusOffline
	if req.Status != nil {
		status = *req.Status
	}
	n, err := s.videoRepo.BatchUpdatePublishStatus(ctx, req.IDs, status)
	if err != nil {
		s.log.Error("management: batch update publish status failed", zap.Int("count", len(req.IDs)), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return &dto.BatchVideoResponse{Affected: n}, nil
}

func (s *managementService) BatchDeleteVideos(ctx context.Context, req *dto.BatchVideoRequest) (*dto.BatchVideoResponse, error) {
	n, err := s.videoRepo.BatchSoftDelete(ctx, req.IDs)
	if err != nil {
		s.log.Error("management: batch delete videos failed", zap.Int("count", len(req.IDs)), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return &dto.BatchVideoResponse{Affected: n}, nil
}

// ===== A3: Admin CRUD =====

func (s *managementService) ListAdmins(ctx context.Context, req *dto.AdminListRequest) ([]dto.AdminItem, int, error) {
	items, total, err := s.adminRepo.ListAdmins(ctx, repository.AdminListFilter{
		Keyword: strings.TrimSpace(req.Keyword),
		Status:  req.Status,
		GroupID: req.GroupID,
		Offset:  req.GetOffset(),
		Limit:   req.GetPageSize(),
	})
	if err != nil {
		s.log.Error("management: list admins failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]dto.AdminItem, 0, len(items))
	for _, a := range items {
		groupName := ""
		if g, _ := s.adminRepo.GetGroupByID(ctx, int64(a.GroupID)); g != nil {
			groupName = g.Name
		}
		out = append(out, dto.AdminItem{
			ID:          a.ID,
			Username:    a.Username,
			Email:       a.Email,
			Avatar:      a.Avatar,
			GroupID:     a.GroupID,
			GroupName:   groupName,
			Status:      a.Status,
			LastLoginAt: formatTimePtr(a.LastLoginAt),
			CreatedAt:   formatTimePtr(a.CreatedAt),
		})
	}
	return out, total, nil
}

func (s *managementService) CreateAdmin(ctx context.Context, req *dto.CreateAdminRequest) (*dto.AdminItem, error) {
	username := strings.TrimSpace(req.Username)
	exists, err := s.adminRepo.ExistsUsername(ctx, username)
	if err != nil {
		s.log.Error("management: check admin username exists failed", zap.String("username", username), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if exists {
		return nil, errcode.AdminAlreadyExists
	}
	group, err := s.adminRepo.GetGroupByID(ctx, int64(req.GroupID))
	if err != nil {
		s.log.Error("management: get group by id for create admin failed", zap.Uint64("group_id", req.GroupID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if group == nil {
		return nil, errcode.UserGroupNotFound
	}
	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		s.log.Error("management: hash password for create admin failed", zap.String("username", username), zap.Error(err))
		return nil, errcode.Wrap(errcode.InternalError, err)
	}
	status := constant.StatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	admin := &model.Admins{
		Username: username,
		Password: hash,
		Email:    strings.TrimSpace(req.Email),
		Avatar:   strings.TrimSpace(req.Avatar),
		GroupID:  req.GroupID,
		Status:   status,
	}
	if err := s.adminRepo.Create(ctx, admin); err != nil {
		s.log.Error("management: create admin failed", zap.String("username", username), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return &dto.AdminItem{
		ID:        admin.ID,
		Username:  admin.Username,
		Email:     admin.Email,
		Avatar:    admin.Avatar,
		GroupID:   admin.GroupID,
		GroupName: group.Name,
		Status:    admin.Status,
		CreatedAt: formatTimePtr(admin.CreatedAt),
	}, nil
}

func (s *managementService) UpdateAdmin(ctx context.Context, id int64, req *dto.UpdateAdminRequest) (*dto.AdminItem, error) {
	admin, err := s.adminRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("management: get admin by id for update failed", zap.Int64("admin_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if admin == nil {
		return nil, errcode.AdminNotFound
	}
	if req.Email != "" {
		admin.Email = strings.TrimSpace(req.Email)
	}
	if req.Avatar != "" {
		admin.Avatar = strings.TrimSpace(req.Avatar)
	}
	if req.GroupID != nil {
		group, err := s.adminRepo.GetGroupByID(ctx, int64(*req.GroupID))
		if err != nil {
			s.log.Error("management: get group by id for update admin failed", zap.Uint64("group_id", *req.GroupID), zap.Error(err))
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		if group == nil {
			return nil, errcode.UserGroupNotFound
		}
		admin.GroupID = *req.GroupID
	}
	if req.Status != nil {
		admin.Status = *req.Status
	}
	if err := s.adminRepo.UpdateAdmin(ctx, admin); err != nil {
		s.log.Error("management: update admin failed", zap.Int64("admin_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	groupName := ""
	if g, _ := s.adminRepo.GetGroupByID(ctx, int64(admin.GroupID)); g != nil {
		groupName = g.Name
	}
	return &dto.AdminItem{
		ID:          admin.ID,
		Username:    admin.Username,
		Email:       admin.Email,
		Avatar:      admin.Avatar,
		GroupID:     admin.GroupID,
		GroupName:   groupName,
		Status:      admin.Status,
		LastLoginAt: formatTimePtr(admin.LastLoginAt),
		CreatedAt:   formatTimePtr(admin.CreatedAt),
	}, nil
}

func (s *managementService) ResetAdminPassword(ctx context.Context, id int64, req *dto.ResetAdminPasswordRequest) error {
	admin, err := s.adminRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("management: get admin by id for reset password failed", zap.Int64("admin_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if admin == nil {
		return errcode.AdminNotFound
	}
	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		s.log.Error("management: hash password for reset admin password failed", zap.Int64("admin_id", id), zap.Error(err))
		return errcode.Wrap(errcode.InternalError, err)
	}
	admin.Password = hash
	if err := s.adminRepo.UpdateAdmin(ctx, admin); err != nil {
		s.log.Error("management: update admin for reset password failed", zap.Int64("admin_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func (s *managementService) DeleteAdmin(ctx context.Context, id int64) error {
	admin, err := s.adminRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("management: get admin by id for delete failed", zap.Int64("admin_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if admin == nil {
		return errcode.AdminNotFound
	}
	if err := s.adminRepo.SoftDeleteAdmin(ctx, id); err != nil {
		s.log.Error("management: soft delete admin failed", zap.Int64("admin_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

// ===== A4: User group CRUD =====

func (s *managementService) ListGroups(ctx context.Context, req *dto.UserGroupListRequest) ([]dto.UserGroupItem, int, error) {
	items, total, err := s.adminRepo.ListGroups(ctx, repository.UserGroupListFilter{
		Keyword: strings.TrimSpace(req.Keyword),
		Offset:  req.GetOffset(),
		Limit:   req.GetPageSize(),
	})
	if err != nil {
		s.log.Error("management: list groups failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]dto.UserGroupItem, 0, len(items))
	for _, g := range items {
		out = append(out, dto.UserGroupItem{
			ID:          g.ID,
			Name:        g.Name,
			Permissions: g.Permissions,
			Description: g.Description,
			CreatedAt:   formatTimePtr(g.CreatedAt),
		})
	}
	return out, total, nil
}

func (s *managementService) CreateGroup(ctx context.Context, req *dto.CreateUserGroupRequest) (*dto.UserGroupItem, error) {
	name := strings.TrimSpace(req.Name)
	g := &model.UserGroups{
		Name:        name,
		Permissions: &req.Permissions,
		Description: strings.TrimSpace(req.Description),
	}
	if err := s.adminRepo.CreateGroup(ctx, g); err != nil {
		s.log.Error("management: create group failed", zap.String("name", name), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return &dto.UserGroupItem{
		ID:          g.ID,
		Name:        g.Name,
		Permissions: g.Permissions,
		Description: g.Description,
		CreatedAt:   formatTimePtr(g.CreatedAt),
	}, nil
}

func (s *managementService) UpdateGroup(ctx context.Context, id int64, req *dto.UpdateUserGroupRequest) (*dto.UserGroupItem, error) {
	group, err := s.adminRepo.GetGroupByID(ctx, id)
	if err != nil {
		s.log.Error("management: get group by id for update failed", zap.Int64("group_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if group == nil {
		return nil, errcode.UserGroupNotFound
	}
	if req.Name != "" {
		name := strings.TrimSpace(req.Name)
		exists, err := s.adminRepo.ExistsGroupNameExcludeID(ctx, name, id)
		if err != nil {
			s.log.Error("management: check group name exists for update failed", zap.Int64("group_id", id), zap.String("name", name), zap.Error(err))
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		if exists {
			return nil, errcode.UserGroupNameDup
		}
		group.Name = name
	}
	if req.Permissions != "" {
		group.Permissions = &req.Permissions
	}
	if req.Description != "" {
		group.Description = strings.TrimSpace(req.Description)
	}
	if err := s.adminRepo.UpdateGroup(ctx, group); err != nil {
		s.log.Error("management: update group failed", zap.Int64("group_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return &dto.UserGroupItem{
		ID:          group.ID,
		Name:        group.Name,
		Permissions: group.Permissions,
		Description: group.Description,
		CreatedAt:   formatTimePtr(group.CreatedAt),
	}, nil
}

func (s *managementService) DeleteGroup(ctx context.Context, id int64) error {
	group, err := s.adminRepo.GetGroupByID(ctx, id)
	if err != nil {
		s.log.Error("management: get group by id for delete failed", zap.Int64("group_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if group == nil {
		return errcode.UserGroupNotFound
	}
	// Prevent deleting super_admin group
	if group.Name == constant.RoleSuperAdmin {
		return errcode.InsufficientPermission
	}
	if err := s.adminRepo.SoftDeleteGroup(ctx, id); err != nil {
		s.log.Error("management: soft delete group failed", zap.Int64("group_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

// ===== A5: Regular user CRUD =====

func (s *managementService) ListUsers(ctx context.Context, req *dto.UserListRequest) ([]dto.UserItem, int, error) {
	items, total, err := s.adminRepo.ListUsers(ctx, repository.UserListFilter{
		Keyword: strings.TrimSpace(req.Keyword),
		Status:  req.Status,
		Offset:  req.GetOffset(),
		Limit:   req.GetPageSize(),
	})
	if err != nil {
		s.log.Error("management: list users failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]dto.UserItem, 0, len(items))
	for _, u := range items {
		out = append(out, dto.UserItem{
			ID:          u.ID,
			Username:    u.Username,
			Email:       u.Email,
			Avatar:      u.Avatar,
			Status:      u.Status,
			LastLoginAt: formatTimePtr(u.LastLoginAt),
			CreatedAt:   formatTimePtr(u.CreatedAt),
		})
	}
	return out, total, nil
}

func (s *managementService) UpdateUser(ctx context.Context, id int64, req *dto.UpdateUserRequest) (*dto.UserItem, error) {
	u, err := s.adminRepo.GetUserByID(ctx, id)
	if err != nil {
		s.log.Error("management: get user by id for update failed", zap.Int64("user_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if u == nil {
		return nil, errcode.UserNotFound
	}
	if req.Email != "" {
		u.Email = strings.TrimSpace(req.Email)
	}
	if req.Avatar != "" {
		u.Avatar = strings.TrimSpace(req.Avatar)
	}
	if req.Status != nil {
		u.Status = *req.Status
	}
	if err := s.adminRepo.UpdateUser(ctx, u); err != nil {
		s.log.Error("management: update user failed", zap.Int64("user_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return &dto.UserItem{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		Avatar:      u.Avatar,
		Status:      u.Status,
		LastLoginAt: formatTimePtr(u.LastLoginAt),
		CreatedAt:   formatTimePtr(u.CreatedAt),
	}, nil
}

func (s *managementService) ResetUserPassword(ctx context.Context, id int64, req *dto.ResetUserPasswordRequest) error {
	u, err := s.adminRepo.GetUserByID(ctx, id)
	if err != nil {
		s.log.Error("management: get user by id for reset password failed", zap.Int64("user_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if u == nil {
		return errcode.UserNotFound
	}
	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		s.log.Error("management: hash password for reset user password failed", zap.Int64("user_id", id), zap.Error(err))
		return errcode.Wrap(errcode.InternalError, err)
	}
	u.Password = hash
	if err := s.adminRepo.UpdateUser(ctx, u); err != nil {
		s.log.Error("management: update user for reset password failed", zap.Int64("user_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func (s *managementService) DeleteUser(ctx context.Context, id int64) error {
	u, err := s.adminRepo.GetUserByID(ctx, id)
	if err != nil {
		s.log.Error("management: get user by id for delete failed", zap.Int64("user_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if u == nil {
		return errcode.UserNotFound
	}
	if err := s.adminRepo.SoftDeleteUser(ctx, id); err != nil {
		s.log.Error("management: soft delete user failed", zap.Int64("user_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

func (s *managementService) ListUserLoginLogs(ctx context.Context, userID int64, offset, limit int) ([]model.UserLoginLogs, int, error) {
	return s.userRepo.ListUserLoginLogs(ctx, userID, offset, limit)
}

// ===== C1: Banner CRUD =====

func (s *managementService) ListBanners(ctx context.Context, offset, limit int) ([]dto.BannerItem, int, error) {
	items, total, err := s.userRepo.ListAllBanners(ctx, offset, limit)
	if err != nil {
		s.log.Error("management: list banners failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]dto.BannerItem, 0, len(items))
	for _, b := range items {
		out = append(out, *toBannerItem(&b))
	}
	return out, total, nil
}

func (s *managementService) CreateBanner(ctx context.Context, req *dto.CreateBannerRequest) (*dto.BannerItem, error) {
	status := constant.StatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	b := &model.Banners{
		Title:   strings.TrimSpace(req.Title),
		Cover:   strings.TrimSpace(req.Cover),
		Link:    strings.TrimSpace(req.Link),
		VideoID: req.VideoID,
		Sort:    req.Sort,
		Status:  status,
	}
	if err := s.userRepo.CreateBanner(ctx, b); err != nil {
		s.log.Error("management: create banner failed", zap.String("title", b.Title), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return toBannerItem(b), nil
}

func (s *managementService) UpdateBanner(ctx context.Context, id int64, req *dto.UpdateBannerRequest) (*dto.BannerItem, error) {
	b, err := s.userRepo.GetBanner(ctx, id)
	if err != nil {
		s.log.Error("management: get banner for update failed", zap.Int64("banner_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if b == nil {
		return nil, errcode.BannerNotFound
	}
	if req.Title != "" {
		b.Title = strings.TrimSpace(req.Title)
	}
	if req.Cover != "" {
		b.Cover = strings.TrimSpace(req.Cover)
	}
	if req.Link != "" {
		b.Link = strings.TrimSpace(req.Link)
	}
	if req.VideoID != nil {
		b.VideoID = *req.VideoID
	}
	if req.Sort != nil {
		b.Sort = *req.Sort
	}
	if req.Status != nil {
		b.Status = *req.Status
	}
	if err := s.userRepo.UpdateBanner(ctx, b); err != nil {
		s.log.Error("management: update banner failed", zap.Int64("banner_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return toBannerItem(b), nil
}

func (s *managementService) DeleteBanner(ctx context.Context, id int64) error {
	b, err := s.userRepo.GetBanner(ctx, id)
	if err != nil {
		s.log.Error("management: get banner for delete failed", zap.Int64("banner_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if b == nil {
		return errcode.BannerNotFound
	}
	if err := s.userRepo.DeleteBanner(ctx, id); err != nil {
		s.log.Error("management: delete banner failed", zap.Int64("banner_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return nil
}

// ===== helpers =====

func toBannerItem(b *model.Banners) *dto.BannerItem {
	return &dto.BannerItem{
		ID:      b.ID,
		Title:   b.Title,
		Cover:   b.Cover,
		Link:    b.Link,
		VideoID: b.VideoID,
		Sort:    b.Sort,
		Status:  b.Status,
	}
}

func formatTimePtr(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}
