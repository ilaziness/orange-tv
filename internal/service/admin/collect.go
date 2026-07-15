package admin

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ilaziness/orange-tv/internal/collect"
	"github.com/ilaziness/orange-tv/internal/constant"
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// CollectService manages collect sources and jobs.
type CollectService interface {
	ListSources(ctx context.Context, req *admindto.CollectSourceListRequest) ([]shareddto.CollectSourceItem, int, error)
	CreateSource(ctx context.Context, req *admindto.CreateCollectSourceRequest) (*shareddto.CollectSourceItem, error)
	UpdateSource(ctx context.Context, id int64, req *admindto.UpdateCollectSourceRequest) (*shareddto.CollectSourceItem, error)
	DeleteSource(ctx context.Context, id int64) error
	ListCategories(ctx context.Context, sourceID int64) ([]shareddto.CollectCategoryMapItem, error)
	SetCategories(ctx context.Context, sourceID int64, req *admindto.SetCollectCategoriesRequest) ([]shareddto.CollectCategoryMapItem, error)
	Start(ctx context.Context, sourceID int64) error
	Stop(ctx context.Context, sourceID int64) error
	ListLogs(ctx context.Context, req *admindto.CollectLogListRequest) ([]shareddto.CollectLogItem, int, error)
	// ReloadScheduler reloads cron jobs from DB (called after source changes and on startup).
	ReloadScheduler(ctx context.Context) error
	StartScheduler(ctx context.Context) error
	StopScheduler(ctx context.Context) error
}

type runningJob struct {
	cancel context.CancelFunc
}

// collectCronParser is shared by validation and the scheduler (5-field + descriptors).
func collectCronParser() cron.Parser {
	return cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
}

type collectService struct {
	repo         repository.CollectRepository
	playRepo     repository.PlayRepository
	categoryRepo repository.CategoryRepository
	engine       *collect.Engine
	log          *zap.Logger

	mu      sync.Mutex
	running map[int64]*runningJob
	cron    *cron.Cron
	cronIDs map[int64]cron.EntryID
}

// NewCollectService creates a CollectService.
func NewCollectService(
	repo repository.CollectRepository,
	playRepo repository.PlayRepository,
	categoryRepo repository.CategoryRepository,
	engine *collect.Engine,
	log *zap.Logger,
) CollectService {
	return &collectService{
		repo:         repo,
		playRepo:     playRepo,
		categoryRepo: categoryRepo,
		engine:       engine,
		log:          log,
		running:      map[int64]*runningJob{},
		cronIDs:      map[int64]cron.EntryID{},
	}
}

func (s *collectService) ListSources(ctx context.Context, req *admindto.CollectSourceListRequest) ([]shareddto.CollectSourceItem, int, error) {
	items, total, err := s.repo.ListSources(ctx, repository.CollectSourceListFilter{
		Status: req.Status,
		Offset: req.GetOffset(),
		Limit:  req.GetLimit(),
	})
	if err != nil {
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	return mapCollectSources(items), total, nil
}

func (s *collectService) CreateSource(ctx context.Context, req *admindto.CreateCollectSourceRequest) (*shareddto.CollectSourceItem, error) {
	if err := s.validateSourceInput(ctx, req.Type, req.CollectURL, req.CronExpr, req.PlaySourceID); err != nil {
		return nil, err
	}
	status := constant.StatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	var cfg *string
	if strings.TrimSpace(req.Config) != "" {
		c := strings.TrimSpace(req.Config)
		cfg = &c
	}
	m := &model.CollectSources{
		Name:         strings.TrimSpace(req.Name),
		Type:         req.Type,
		CollectURL:   strings.TrimSpace(req.CollectURL),
		APIKey:       strings.TrimSpace(req.APIKey),
		Config:       cfg,
		CronExpr:     strings.TrimSpace(req.CronExpr),
		PlaySourceID: req.PlaySourceID,
		Status:       status,
	}
	if err := s.repo.CreateSource(ctx, m); err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	_ = s.ReloadScheduler(ctx)
	out := toCollectSource(m)
	return &out, nil
}

func (s *collectService) UpdateSource(ctx context.Context, id int64, req *admindto.UpdateCollectSourceRequest) (*shareddto.CollectSourceItem, error) {
	m, err := s.repo.GetSource(ctx, id)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if m == nil {
		return nil, errcode.CollectSourceNotFound
	}
	typ := m.Type
	collectURL := m.CollectURL
	cronExpr := m.CronExpr
	playSourceID := m.PlaySourceID
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, errcode.WithMessage(errcode.ParamError, "采集源名称不能为空")
		}
		m.Name = name
	}
	if req.Type != nil {
		typ = *req.Type
		m.Type = typ
	}
	if req.CollectURL != nil {
		collectURL = strings.TrimSpace(*req.CollectURL)
		m.CollectURL = collectURL
	}
	if req.APIKey != nil {
		m.APIKey = strings.TrimSpace(*req.APIKey)
	}
	if req.Config != nil {
		c := strings.TrimSpace(*req.Config)
		if c == "" {
			m.Config = nil
		} else {
			m.Config = &c
		}
	}
	if req.CronExpr != nil {
		cronExpr = strings.TrimSpace(*req.CronExpr)
		m.CronExpr = cronExpr
	}
	if req.PlaySourceID != nil {
		playSourceID = *req.PlaySourceID
		m.PlaySourceID = playSourceID
	}
	if req.Status != nil {
		m.Status = *req.Status
	}
	if err := s.validateSourceInput(ctx, typ, collectURL, cronExpr, playSourceID); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateSource(ctx, m); err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	_ = s.ReloadScheduler(ctx)
	out := toCollectSource(m)
	return &out, nil
}

func (s *collectService) DeleteSource(ctx context.Context, id int64) error {
	m, err := s.repo.GetSource(ctx, id)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if m == nil {
		return errcode.CollectSourceNotFound
	}
	_ = s.Stop(ctx, id)
	if err := s.repo.SoftDeleteSource(ctx, id); err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	_ = s.ReloadScheduler(ctx)
	return nil
}

func (s *collectService) ListCategories(ctx context.Context, sourceID int64) ([]shareddto.CollectCategoryMapItem, error) {
	if _, err := s.requireSource(ctx, sourceID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListCategories(ctx, sourceID)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]shareddto.CollectCategoryMapItem, 0, len(items))
	for _, it := range items {
		out = append(out, shareddto.CollectCategoryMapItem{
			ID:               it.ID,
			SourceID:         it.SourceID,
			ExternalCategory: it.ExternalCategory,
			CategoryID:       it.CategoryID,
		})
	}
	return out, nil
}

func (s *collectService) SetCategories(ctx context.Context, sourceID int64, req *admindto.SetCollectCategoriesRequest) ([]shareddto.CollectCategoryMapItem, error) {
	if _, err := s.requireSource(ctx, sourceID); err != nil {
		return nil, err
	}
	rows := make([]model.CollectSourceCategories, 0, len(req.Items))
	seen := map[string]bool{}
	for _, in := range req.Items {
		ext := strings.TrimSpace(in.ExternalCategory)
		if ext == "" || in.CategoryID <= 0 {
			return nil, errcode.WithMessage(errcode.ParamError, "分类映射参数无效")
		}
		if seen[ext] {
			continue
		}
		seen[ext] = true
		cat, err := s.categoryRepo.GetByID(ctx, in.CategoryID)
		if err != nil {
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		if cat == nil {
			return nil, errcode.CategoryNotFound
		}
		rows = append(rows, model.CollectSourceCategories{
			ExternalCategory: ext,
			CategoryID:       in.CategoryID,
		})
	}
	if err := s.repo.ReplaceCategories(ctx, sourceID, rows); err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return s.ListCategories(ctx, sourceID)
}

func (s *collectService) Start(ctx context.Context, sourceID int64) error {
	source, err := s.requireSource(ctx, sourceID)
	if err != nil {
		return err
	}
	if source.Status != constant.StatusEnabled {
		return errcode.CollectSourceDisabled
	}
	maps, err := s.repo.ListCategories(ctx, sourceID)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if len(maps) == 0 {
		return errcode.CollectCategoryMapEmpty
	}

	s.mu.Lock()
	if _, ok := s.running[sourceID]; ok {
		s.mu.Unlock()
		return errcode.CollectAlreadyRunning
	}
	jobCtx, cancel := context.WithCancel(context.Background())
	s.running[sourceID] = &runningJob{cancel: cancel}
	s.mu.Unlock()

	go s.runJob(jobCtx, source)
	return nil
}

func (s *collectService) Stop(ctx context.Context, sourceID int64) error {
	s.mu.Lock()
	job, ok := s.running[sourceID]
	if !ok {
		s.mu.Unlock()
		return errcode.CollectNotRunning
	}
	// Only cancel; runJob defer removes the entry so Start cannot race with a half-deleted job.
	job.cancel()
	s.mu.Unlock()
	return nil
}

func (s *collectService) ListLogs(ctx context.Context, req *admindto.CollectLogListRequest) ([]shareddto.CollectLogItem, int, error) {
	items, total, err := s.repo.ListLogs(ctx, repository.CollectLogListFilter{
		SourceID: req.SourceID,
		Offset:   req.GetOffset(),
		Limit:    req.GetLimit(),
	})
	if err != nil {
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]shareddto.CollectLogItem, 0, len(items))
	for _, it := range items {
		msg := ""
		if it.ErrorMessage != nil {
			msg = *it.ErrorMessage
		}
		created := ""
		if it.CreatedAt != nil {
			created = it.CreatedAt.Format(time.RFC3339)
		}
		out = append(out, shareddto.CollectLogItem{
			ID:           it.ID,
			SourceID:     it.SourceID,
			Status:       it.Status,
			TotalCount:   it.TotalCount,
			SuccessCount: it.SuccessCount,
			FailedCount:  it.FailedCount,
			ErrorMessage: msg,
			DurationMs:   it.DurationMs,
			CreatedAt:    created,
		})
	}
	return out, total, nil
}

func (s *collectService) runJob(ctx context.Context, source *model.CollectSources) {
	defer func() {
		s.mu.Lock()
		delete(s.running, source.ID)
		s.mu.Unlock()
	}()

	start := time.Now()
	logRow := &model.CollectLogs{
		SourceID: source.ID,
		Status:   constant.CollectLogRunning,
	}
	if err := s.repo.CreateLog(context.Background(), logRow); err != nil && s.log != nil {
		s.log.Error("create collect log failed", zap.Error(err))
	}

	res := s.engine.Run(ctx, source)
	status := constant.CollectLogSuccess
	switch {
	case ctx.Err() != nil:
		if res.Success > 0 && res.Failed > 0 {
			status = constant.CollectLogPartialSuccess
		} else if res.Success > 0 {
			status = constant.CollectLogSuccess
		} else {
			status = constant.CollectLogCancelled
		}
		if res.Message == "" {
			res.Message = "采集已取消"
		}
	case res.Message != "" && res.Success == 0 && res.Failed == 0:
		status = constant.CollectLogFailed
	case res.Success == 0 && res.Failed > 0:
		status = constant.CollectLogFailed
	case res.Failed > 0:
		status = constant.CollectLogPartialSuccess
	}

	dur := int32(time.Since(start).Milliseconds())
	logRow.Status = status
	logRow.TotalCount = int32(res.Total)
	logRow.SuccessCount = int32(res.Success)
	logRow.FailedCount = int32(res.Failed)
	logRow.DurationMs = dur
	if res.Message != "" {
		msg := res.Message
		logRow.ErrorMessage = &msg
	}
	if logRow.ID > 0 {
		_ = s.repo.UpdateLog(context.Background(), logRow)
	}
	_ = s.repo.TouchLastCollect(context.Background(), source.ID, time.Now())
	if s.log != nil {
		s.log.Info("collect finished",
			zap.Int64("source_id", source.ID),
			zap.Int8("status", status),
			zap.Int("success", res.Success),
			zap.Int("failed", res.Failed),
		)
	}
}

func (s *collectService) StartScheduler(ctx context.Context) error {
	s.mu.Lock()
	if s.cron == nil {
		s.cron = cron.New(cron.WithParser(collectCronParser()), cron.WithChain(cron.Recover(cron.DefaultLogger)))
		s.cron.Start()
	}
	s.mu.Unlock()
	return s.ReloadScheduler(ctx)
}

func (s *collectService) StopScheduler(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cron != nil {
		stopCtx := s.cron.Stop()
		select {
		case <-stopCtx.Done():
		case <-ctx.Done():
		case <-time.After(5 * time.Second):
		}
		s.cron = nil
		s.cronIDs = map[int64]cron.EntryID{}
	}
	// cancel all running jobs
	for id, job := range s.running {
		job.cancel()
		delete(s.running, id)
	}
	return nil
}

func (s *collectService) ReloadScheduler(ctx context.Context) error {
	s.mu.Lock()
	if s.cron == nil {
		s.mu.Unlock()
		return nil
	}
	// remove old entries
	for id, entryID := range s.cronIDs {
		s.cron.Remove(entryID)
		delete(s.cronIDs, id)
	}
	s.mu.Unlock()

	sources, err := s.repo.ListEnabledCronSources(ctx)
	if err != nil {
		return err
	}
	parser := collectCronParser()
	for i := range sources {
		src := sources[i]
		expr := strings.TrimSpace(src.CronExpr)
		if expr == "" {
			continue
		}
		if _, err := parser.Parse(expr); err != nil {
			if s.log != nil {
				s.log.Warn("skip invalid cron", zap.Int64("source_id", src.ID), zap.Error(err))
			}
			continue
		}
		sourceID := src.ID
		s.mu.Lock()
		entryID, err := s.cron.AddFunc(expr, func() {
			_ = s.Start(context.Background(), sourceID)
		})
		if err != nil {
			s.mu.Unlock()
			if s.log != nil {
				s.log.Warn("add cron failed", zap.Int64("source_id", sourceID), zap.Error(err))
			}
			continue
		}
		s.cronIDs[sourceID] = entryID
		s.mu.Unlock()
	}
	return nil
}

func (s *collectService) validateSourceInput(ctx context.Context, typ int8, collectURL, cronExpr string, playSourceID int64) error {
	if typ != constant.CollectTypeDefault && typ != constant.CollectTypeAppleCMS {
		return errcode.WithMessage(errcode.ParamError, "采集源类型无效")
	}
	u := strings.TrimSpace(collectURL)
	if u == "" {
		return errcode.WithMessage(errcode.ParamError, "采集地址不能为空")
	}
	parsed, err := url.Parse(u)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return errcode.WithMessage(errcode.ParamError, "采集地址格式无效")
	}
	if strings.TrimSpace(cronExpr) != "" {
		if _, err := collectCronParser().Parse(strings.TrimSpace(cronExpr)); err != nil {
			return errcode.CollectInvalidCron
		}
	}
	src, err := s.playRepo.GetSource(ctx, playSourceID)
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if src == nil {
		return errcode.PlaySourceNotFound
	}
	return nil
}

func (s *collectService) requireSource(ctx context.Context, id int64) (*model.CollectSources, error) {
	m, err := s.repo.GetSource(ctx, id)
	if err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if m == nil {
		return nil, errcode.CollectSourceNotFound
	}
	return m, nil
}

func mapCollectSources(items []model.CollectSources) []shareddto.CollectSourceItem {
	out := make([]shareddto.CollectSourceItem, 0, len(items))
	for i := range items {
		out = append(out, toCollectSource(&items[i]))
	}
	return out
}

func toCollectSource(m *model.CollectSources) shareddto.CollectSourceItem {
	cfg := ""
	if m.Config != nil {
		cfg = *m.Config
	}
	last := ""
	if m.LastCollectAt != nil {
		last = m.LastCollectAt.Format(time.RFC3339)
	}
	return shareddto.CollectSourceItem{
		ID:            m.ID,
		Name:          m.Name,
		Type:          m.Type,
		CollectURL:    m.CollectURL,
		Config:        cfg,
		CronExpr:      m.CronExpr,
		PlaySourceID:  m.PlaySourceID,
		LastCollectAt: last,
		Status:        m.Status,
	}
}
