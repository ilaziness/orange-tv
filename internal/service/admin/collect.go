package admin

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ilaziness/orange-tv/internal/cache"
	"github.com/ilaziness/orange-tv/internal/collect"
	"github.com/ilaziness/orange-tv/internal/constant"
	admindto "github.com/ilaziness/orange-tv/internal/dto/admin"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/internal/repository"
	"github.com/ilaziness/orange-tv/internal/utils"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// CollectService manages collect sources and jobs.
type CollectService interface {
	ListSources(ctx context.Context, req *admindto.CollectSourceListRequest) ([]admindto.CollectSourceItem, int, error)
	CreateSource(ctx context.Context, req *admindto.CreateCollectSourceRequest) (*admindto.CollectSourceItem, error)
	UpdateSource(ctx context.Context, id uint32, req *admindto.UpdateCollectSourceRequest) (*admindto.CollectSourceItem, error)
	DeleteSource(ctx context.Context, id uint32) error
	ListCategories(ctx context.Context, sourceID uint32) ([]admindto.CollectCategoryMapItem, error)
	SetCategories(ctx context.Context, sourceID uint32, req *admindto.SetCollectCategoriesRequest) ([]admindto.CollectCategoryMapItem, error)
	ListLogs(ctx context.Context, req *admindto.CollectLogListRequest) ([]admindto.CollectLogItem, int, error)
	FetchRemoteCategories(ctx context.Context, sourceID uint32) (*admindto.RemoteCategoryResponse, error)
	EnableSchedule(ctx context.Context, sourceID uint32) error
	DisableSchedule(ctx context.Context, sourceID uint32) error
	CollectNow(ctx context.Context, sourceID uint32, req *admindto.CollectNowRequest) error
	// ReloadScheduler reloads cron jobs from DB (called on startup and after schedule changes).
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
	cache        *cache.Manager

	mu      sync.Mutex
	running map[uint32]*runningJob
	cron    *cron.Cron
	cronIDs map[uint32]cron.EntryID
}

// NewCollectService creates a CollectService.
func NewCollectService(
	repo repository.CollectRepository,
	playRepo repository.PlayRepository,
	categoryRepo repository.CategoryRepository,
	engine *collect.Engine,
	log *zap.Logger,
	c *cache.Manager,
) CollectService {
	if log == nil {
		log = zap.NewNop()
	}
	return &collectService{
		repo:         repo,
		playRepo:     playRepo,
		categoryRepo: categoryRepo,
		engine:       engine,
		log:          log,
		cache:        c,
		running:      map[uint32]*runningJob{},
		cronIDs:      map[uint32]cron.EntryID{},
	}
}

func (s *collectService) ListSources(ctx context.Context, req *admindto.CollectSourceListRequest) ([]admindto.CollectSourceItem, int, error) {
	items, total, err := s.repo.ListSources(ctx, repository.CollectSourceListFilter{
		Status: req.Status,
		Offset: req.GetOffset(),
		Limit:  req.GetLimit(),
	})
	if err != nil {
		s.log.Error("collect: list sources failed", zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	// batch query play source names
	psourceIDs := make(map[uint32]struct{})
	for _, it := range items {
		if it.PlaySourceID > 0 {
			psourceIDs[it.PlaySourceID] = struct{}{}
		}
	}
	psourceNames := make(map[uint32]string)
	for pid := range psourceIDs {
		psrc, err := s.playRepo.GetSource(ctx, pid)
		if err == nil && psrc != nil {
			psourceNames[pid] = psrc.Name
		}
	}
	out := make([]admindto.CollectSourceItem, 0, len(items))
	for i := range items {
		item := toCollectSource(&items[i])
		item.PlaySourceName = psourceNames[items[i].PlaySourceID]
		out = append(out, item)
	}
	return out, total, nil
}

func (s *collectService) CreateSource(ctx context.Context, req *admindto.CreateCollectSourceRequest) (*admindto.CollectSourceItem, error) {
	if err := s.validateSourceInput(ctx, req.Type, req.CollectURL, req.CronExpr, req.PlaySourceID); err != nil {
		return nil, err
	}
	m := &model.CollectSources{
		Name:            strings.TrimSpace(req.Name),
		Type:            req.Type,
		CollectURL:      strings.TrimSpace(req.CollectURL),
		APIKey:          strings.TrimSpace(req.APIKey),
		CronExpr:        strings.TrimSpace(req.CronExpr),
		PlaySourceID:    req.PlaySourceID,
		Status:          constant.StatusDisabled,
		ScheduleEnabled: 0,
		DataRange:       strings.TrimSpace(req.DataRange),
	}
	if err := s.repo.CreateSource(ctx, m); err != nil {
		s.log.Error("collect: create source failed", zap.String("name", m.Name), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := toCollectSource(m)
	if m.PlaySourceID > 0 {
		if psrc, err := s.playRepo.GetSource(ctx, m.PlaySourceID); err == nil && psrc != nil {
			out.PlaySourceName = psrc.Name
		}
	}
	return &out, nil
}

func (s *collectService) UpdateSource(ctx context.Context, id uint32, req *admindto.UpdateCollectSourceRequest) (*admindto.CollectSourceItem, error) {
	m, err := s.repo.GetSource(ctx, id)
	if err != nil {
		s.log.Error("collect: get source failed", zap.Uint32("source_id", id), zap.Error(err))
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
	if req.DataRange != nil {
		m.DataRange = strings.TrimSpace(*req.DataRange)
	}
	if req.CronExpr != nil {
		cronExpr = strings.TrimSpace(*req.CronExpr)
		m.CronExpr = cronExpr
	}
	if req.PlaySourceID != nil {
		playSourceID = *req.PlaySourceID
		m.PlaySourceID = playSourceID
	}
	if err := s.validateSourceInput(ctx, typ, collectURL, cronExpr, playSourceID); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateSource(ctx, m); err != nil {
		s.log.Error("collect: update source failed", zap.Uint32("source_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := toCollectSource(m)
	if m.PlaySourceID > 0 {
		if psrc, err := s.playRepo.GetSource(ctx, m.PlaySourceID); err == nil && psrc != nil {
			out.PlaySourceName = psrc.Name
		}
	}
	return &out, nil
}

func (s *collectService) DeleteSource(ctx context.Context, id uint32) error {
	m, err := s.repo.GetSource(ctx, id)
	if err != nil {
		s.log.Error("collect: get source for delete failed", zap.Uint32("source_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if m == nil {
		return errcode.CollectSourceNotFound
	}
	s.stopJob(id)
	if err := s.repo.SoftDeleteSource(ctx, id); err != nil {
		s.log.Error("collect: delete source failed", zap.Uint32("source_id", id), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return s.ReloadScheduler(ctx)
}

func (s *collectService) ListCategories(ctx context.Context, sourceID uint32) ([]admindto.CollectCategoryMapItem, error) {
	if _, err := s.requireSource(ctx, sourceID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListCategories(ctx, sourceID)
	if err != nil {
		s.log.Error("collect: list categories failed", zap.Uint32("source_id", sourceID), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := make([]admindto.CollectCategoryMapItem, 0, len(items))
	for _, it := range items {
		out = append(out, admindto.CollectCategoryMapItem{
			ID:                 it.ID,
			SourceID:           it.SourceID,
			ExternalCategoryID: it.ExternalCategoryID,
			CategoryID:         it.CategoryID,
		})
	}
	return out, nil
}

func (s *collectService) SetCategories(ctx context.Context, sourceID uint32, req *admindto.SetCollectCategoriesRequest) ([]admindto.CollectCategoryMapItem, error) {
	if _, err := s.requireSource(ctx, sourceID); err != nil {
		return nil, err
	}
	rows := make([]model.CollectSourceCategories, 0, len(req.Items))
	seen := map[uint32]bool{}
	for _, in := range req.Items {
		if in.ExternalCategoryID == 0 || in.CategoryID == 0 {
			return nil, errcode.WithMessage(errcode.ParamError, "分类映射参数无效")
		}
		if seen[in.ExternalCategoryID] {
			continue
		}
		seen[in.ExternalCategoryID] = true
		cat, err := s.categoryRepo.GetByID(ctx, in.CategoryID)
		if err != nil {
			s.log.Error("collect: get category by id failed", zap.Uint32("category_id", in.CategoryID), zap.Error(err))
			return nil, errcode.Wrap(errcode.DatabaseError, err)
		}
		if cat == nil {
			return nil, errcode.CategoryNotFound
		}
		rows = append(rows, model.CollectSourceCategories{
			ExternalCategoryID: in.ExternalCategoryID,
			CategoryID:         in.CategoryID,
		})
	}
	if err := s.repo.ReplaceCategories(ctx, sourceID, rows); err != nil {
		s.log.Error("collect: replace categories failed", zap.Uint32("source_id", sourceID), zap.Int("count", len(rows)), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	return s.ListCategories(ctx, sourceID)
}

// startJob launches a collection goroutine for the given source.
func (s *collectService) startJob(source *model.CollectSources, dataRange string) error {
	if strings.TrimSpace(dataRange) == "" {
		dataRange = "all"
	}
	s.mu.Lock()
	if _, ok := s.running[source.ID]; ok {
		s.mu.Unlock()
		return errcode.CollectAlreadyRunning
	}
	jobCtx, cancel := context.WithCancel(context.Background())
	s.running[source.ID] = &runningJob{cancel: cancel}
	s.mu.Unlock()

	utils.Go(func() { s.runJob(jobCtx, source, dataRange) })
	return nil
}

// stopJob cancels a running collection job if any.
func (s *collectService) stopJob(sourceID uint32) {
	s.mu.Lock()
	job, ok := s.running[sourceID]
	if !ok {
		s.mu.Unlock()
		return
	}
	job.cancel()
	s.mu.Unlock()
}

func (s *collectService) ListLogs(ctx context.Context, req *admindto.CollectLogListRequest) ([]admindto.CollectLogItem, int, error) {
	items, total, err := s.repo.ListLogs(ctx, repository.CollectLogListFilter{
		SourceID: req.SourceID,
		Offset:   req.GetOffset(),
		Limit:    req.GetLimit(),
	})
	if err != nil {
		s.log.Error("collect: list logs failed", zap.Uint32("source_id", req.SourceID), zap.Error(err))
		return nil, 0, errcode.Wrap(errcode.DatabaseError, err)
	}
	// batch query source names
	sourceIDs := make(map[uint32]struct{})
	for _, it := range items {
		sourceIDs[it.SourceID] = struct{}{}
	}
	sourceNames := make(map[uint32]string)
	for sid := range sourceIDs {
		src, err := s.repo.GetSource(ctx, sid)
		if err == nil && src != nil {
			sourceNames[sid] = src.Name
		}
	}
	out := make([]admindto.CollectLogItem, 0, len(items))
	for _, it := range items {
		created := it.CreatedAt.Format(time.DateTime)
		out = append(out, admindto.CollectLogItem{
			ID:           it.ID,
			SourceID:     it.SourceID,
			SourceName:   sourceNames[it.SourceID],
			Status:       it.Status,
			CollectCount: it.CollectCount,
			DurationSec:  it.DurationSec,
			CreatedAt:    created,
		})
	}
	return out, total, nil
}

func (s *collectService) runJob(ctx context.Context, source *model.CollectSources, dataRange string) {
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
	if err := s.repo.CreateLog(context.Background(), logRow); err != nil {
		s.log.Error("create collect log failed", zap.Error(err))
	}

	res := s.engine.Run(ctx, source, dataRange, logRow.ID)

	dur := uint32(time.Since(start).Seconds())
	if res.HasError {
		logRow.Status = constant.CollectLogFailed
	} else {
		logRow.Status = constant.CollectLogCompleted
	}
	logRow.DurationSec = dur
	if logRow.ID > 0 {
		_ = s.repo.UpdateLog(context.Background(), logRow)
	}
	_ = s.repo.TouchLastCollect(context.Background(), source.ID, time.Now())
	s.log.Info("collect finished",
		zap.Uint32("source_id", source.ID),
		zap.Uint8("status", logRow.Status),
		zap.Int("collected", res.CollectCount),
	)
	if res.CollectCount > 0 {
		_ = s.cache.Clear(context.Background())
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
		s.cronIDs = map[uint32]cron.EntryID{}
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
		s.log.Error("collect: list enabled cron sources failed", zap.Error(err))
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
			s.log.Warn("skip invalid cron", zap.Uint32("source_id", src.ID), zap.Error(err))
			continue
		}
		sourceID := src.ID
		dataRange := src.DataRange
		s.mu.Lock()
		entryID, err := s.cron.AddFunc(expr, func() {
			_ = s.startJob(&src, dataRange)
		})
		if err != nil {
			s.mu.Unlock()
			s.log.Warn("add cron failed", zap.Uint32("source_id", sourceID), zap.Error(err))
			continue
		}
		s.cronIDs[sourceID] = entryID
		s.mu.Unlock()
	}
	return nil
}

// validateCollectPrecondition checks that a source is ready for collection.
// Called before launching the goroutine so errors are returned to the user.
func (s *collectService) validateCollectPrecondition(ctx context.Context, source *model.CollectSources) error {
	maps, err := s.repo.ListCategories(ctx, source.ID)
	if err != nil {
		s.log.Error("collect: list categories for precondition failed", zap.Uint32("source_id", source.ID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if len(maps) == 0 {
		return errcode.CollectCategoryMapEmpty
	}
	if source.PlaySourceID <= 0 {
		return errcode.WithMessage(errcode.ParamError, "采集源未绑定播放源")
	}
	playSrc, err := s.playRepo.GetSource(ctx, source.PlaySourceID)
	if err != nil {
		s.log.Error("collect: get play source for precondition failed", zap.Uint32("play_source_id", source.PlaySourceID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if playSrc == nil {
		return errcode.PlaySourceNotFound
	}
	return nil
}

func (s *collectService) validateSourceInput(ctx context.Context, typ uint8, collectURL, cronExpr string, playSourceID uint32) error {
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
		s.log.Error("collect: get play source for validation failed", zap.Uint32("play_source_id", playSourceID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if src == nil {
		return errcode.PlaySourceNotFound
	}
	return nil
}

func (s *collectService) requireSource(ctx context.Context, id uint32) (*model.CollectSources, error) {
	m, err := s.repo.GetSource(ctx, id)
	if err != nil {
		s.log.Error("collect: get source failed", zap.Uint32("source_id", id), zap.Error(err))
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	if m == nil {
		return nil, errcode.CollectSourceNotFound
	}
	return m, nil
}

func toCollectSource(m *model.CollectSources) admindto.CollectSourceItem {
	last := ""
	if m.LastCollectAt != nil {
		last = m.LastCollectAt.Format(time.DateTime)
	}
	return admindto.CollectSourceItem{
		ID:              m.ID,
		Name:            m.Name,
		Type:            m.Type,
		CollectURL:      m.CollectURL,
		CronExpr:        m.CronExpr,
		PlaySourceID:    m.PlaySourceID,
		LastCollectAt:   last,
		Status:          m.Status,
		ScheduleEnabled: m.ScheduleEnabled,
		DataRange:       m.DataRange,
	}
}

func (s *collectService) FetchRemoteCategories(ctx context.Context, sourceID uint32) (*admindto.RemoteCategoryResponse, error) {
	source, err := s.requireSource(ctx, sourceID)
	if err != nil {
		return nil, err
	}
	// Both Apple CMS (class) and default (Open API /categories) formats expose remote categories.
	cats, err := collect.FetchCategories(ctx, source, s.log)
	if err != nil {
		s.log.Error("collect: fetch remote categories failed", zap.Uint32("source_id", sourceID), zap.Error(err))
		return nil, errcode.Wrap(errcode.CollectFetchFailed, err)
	}
	items := make([]admindto.RemoteCategoryItem, 0, len(cats))
	for _, c := range cats {
		if c.ID == 0 {
			continue
		}
		items = append(items, admindto.RemoteCategoryItem{
			TypeID:   c.ID,
			TypeName: c.Name,
			TypePID:  c.ParentID,
		})
	}
	return &admindto.RemoteCategoryResponse{List: items}, nil
}

func (s *collectService) EnableSchedule(ctx context.Context, sourceID uint32) error {
	source, err := s.requireSource(ctx, sourceID)
	if err != nil {
		return err
	}
	if err := s.validateCollectPrecondition(ctx, source); err != nil {
		return err
	}
	expr := strings.TrimSpace(source.CronExpr)
	if expr == "" {
		return errcode.WithMessage(errcode.ParamError, "请先设置定时时间")
	}
	if _, err := collectCronParser().Parse(expr); err != nil {
		return errcode.CollectInvalidCron
	}
	source.ScheduleEnabled = 1
	if err := s.repo.UpdateSource(ctx, source); err != nil {
		s.log.Error("collect: enable schedule failed", zap.Uint32("source_id", sourceID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return s.ReloadScheduler(ctx)
}

func (s *collectService) DisableSchedule(ctx context.Context, sourceID uint32) error {
	source, err := s.requireSource(ctx, sourceID)
	if err != nil {
		return err
	}
	source.ScheduleEnabled = 0
	if err := s.repo.UpdateSource(ctx, source); err != nil {
		s.log.Error("collect: disable schedule failed", zap.Uint32("source_id", sourceID), zap.Error(err))
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return s.ReloadScheduler(ctx)
}

func (s *collectService) CollectNow(ctx context.Context, sourceID uint32, req *admindto.CollectNowRequest) error {
	source, err := s.requireSource(ctx, sourceID)
	if err != nil {
		return err
	}
	if err := s.validateCollectPrecondition(ctx, source); err != nil {
		return err
	}
	return s.startJob(source, req.DataRange)
}
