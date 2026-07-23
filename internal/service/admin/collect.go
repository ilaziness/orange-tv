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
	shareddto "github.com/ilaziness/orange-tv/internal/dto"
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
	ListSources(ctx context.Context, req *admindto.CollectSourceListRequest) ([]shareddto.CollectSourceItem, int, error)
	CreateSource(ctx context.Context, req *admindto.CreateCollectSourceRequest) (*shareddto.CollectSourceItem, error)
	UpdateSource(ctx context.Context, id int64, req *admindto.UpdateCollectSourceRequest) (*shareddto.CollectSourceItem, error)
	DeleteSource(ctx context.Context, id int64) error
	ListCategories(ctx context.Context, sourceID int64) ([]shareddto.CollectCategoryMapItem, error)
	SetCategories(ctx context.Context, sourceID int64, req *admindto.SetCollectCategoriesRequest) ([]shareddto.CollectCategoryMapItem, error)
	ListLogs(ctx context.Context, req *admindto.CollectLogListRequest) ([]shareddto.CollectLogItem, int, error)
	FetchRemoteCategories(ctx context.Context, sourceID int64) (*admindto.RemoteCategoryResponse, error)
	EnableSchedule(ctx context.Context, sourceID int64) error
	DisableSchedule(ctx context.Context, sourceID int64) error
	CollectNow(ctx context.Context, sourceID int64, req *admindto.CollectNowRequest) error
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
	cache        cache.Cache

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
	c cache.Cache,
) CollectService {
	if c == nil {
		c = cache.NewNopCache()
	}
	return &collectService{
		repo:         repo,
		playRepo:     playRepo,
		categoryRepo: categoryRepo,
		engine:       engine,
		log:          log,
		cache:        c,
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
	// batch query play source names
	psourceIDs := make(map[uint64]struct{})
	for _, it := range items {
		if it.PlaySourceID > 0 {
			psourceIDs[it.PlaySourceID] = struct{}{}
		}
	}
	psourceNames := make(map[uint64]string)
	for pid := range psourceIDs {
		psrc, err := s.playRepo.GetSource(ctx, int64(pid))
		if err == nil && psrc != nil {
			psourceNames[pid] = psrc.Name
		}
	}
	out := make([]shareddto.CollectSourceItem, 0, len(items))
	for i := range items {
		item := toCollectSource(&items[i])
		item.PlaySourceName = psourceNames[items[i].PlaySourceID]
		out = append(out, item)
	}
	return out, total, nil
}

func (s *collectService) CreateSource(ctx context.Context, req *admindto.CreateCollectSourceRequest) (*shareddto.CollectSourceItem, error) {
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
		Status:          uint8(constant.StatusDisabled),
		ScheduleEnabled: 0,
		DataRange:       strings.TrimSpace(req.DataRange),
	}
	if err := s.repo.CreateSource(ctx, m); err != nil {
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := toCollectSource(m)
	if m.PlaySourceID > 0 {
		if psrc, err := s.playRepo.GetSource(ctx, int64(m.PlaySourceID)); err == nil && psrc != nil {
			out.PlaySourceName = psrc.Name
		}
	}
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
		return nil, errcode.Wrap(errcode.DatabaseError, err)
	}
	out := toCollectSource(m)
	if m.PlaySourceID > 0 {
		if psrc, err := s.playRepo.GetSource(ctx, int64(m.PlaySourceID)); err == nil && psrc != nil {
			out.PlaySourceName = psrc.Name
		}
	}
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
	s.stopJob(id)
	if err := s.repo.SoftDeleteSource(ctx, id); err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return s.ReloadScheduler(ctx)
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
			ID:                 it.ID,
			SourceID:           it.SourceID,
			ExternalCategoryID: it.ExternalCategoryID,
			CategoryID:         it.CategoryID,
		})
	}
	return out, nil
}

func (s *collectService) SetCategories(ctx context.Context, sourceID int64, req *admindto.SetCollectCategoriesRequest) ([]shareddto.CollectCategoryMapItem, error) {
	if _, err := s.requireSource(ctx, sourceID); err != nil {
		return nil, err
	}
	rows := make([]model.CollectSourceCategories, 0, len(req.Items))
	seen := map[uint64]bool{}
	for _, in := range req.Items {
		if in.ExternalCategoryID == 0 || in.CategoryID == 0 {
			return nil, errcode.WithMessage(errcode.ParamError, "分类映射参数无效")
		}
		if seen[in.ExternalCategoryID] {
			continue
		}
		seen[in.ExternalCategoryID] = true
		cat, err := s.categoryRepo.GetByID(ctx, int64(in.CategoryID))
		if err != nil {
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
	if _, ok := s.running[int64(source.ID)]; ok {
		s.mu.Unlock()
		return errcode.CollectAlreadyRunning
	}
	jobCtx, cancel := context.WithCancel(context.Background())
	s.running[int64(source.ID)] = &runningJob{cancel: cancel}
	s.mu.Unlock()

	utils.Go(func() { s.runJob(jobCtx, source, dataRange) })
	return nil
}

// stopJob cancels a running collection job if any.
func (s *collectService) stopJob(sourceID int64) {
	s.mu.Lock()
	job, ok := s.running[sourceID]
	if !ok {
		s.mu.Unlock()
		return
	}
	job.cancel()
	s.mu.Unlock()
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
	// batch query source names
	sourceIDs := make(map[uint64]struct{})
	for _, it := range items {
		sourceIDs[it.SourceID] = struct{}{}
	}
	sourceNames := make(map[uint64]string)
	for sid := range sourceIDs {
		src, err := s.repo.GetSource(ctx, int64(sid))
		if err == nil && src != nil {
			sourceNames[sid] = src.Name
		}
	}
	out := make([]shareddto.CollectLogItem, 0, len(items))
	for _, it := range items {
		created := ""
		if it.CreatedAt != nil {
			created = it.CreatedAt.Format(time.DateTime)
		}
		out = append(out, shareddto.CollectLogItem{
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
		delete(s.running, int64(source.ID))
		s.mu.Unlock()
	}()

	start := time.Now()
	logRow := &model.CollectLogs{
		SourceID: source.ID,
		Status:   uint8(constant.CollectLogRunning),
	}
	if err := s.repo.CreateLog(context.Background(), logRow); err != nil && s.log != nil {
		s.log.Error("create collect log failed", zap.Error(err))
	}

	res := s.engine.Run(ctx, source, dataRange, logRow.ID)

	dur := uint32(time.Since(start).Seconds())
	if res.HasError {
		logRow.Status = uint8(constant.CollectLogFailed)
	} else {
		logRow.Status = uint8(constant.CollectLogCompleted)
	}
	logRow.DurationSec = dur
	if logRow.ID > 0 {
		_ = s.repo.UpdateLog(context.Background(), logRow)
	}
	_ = s.repo.TouchLastCollect(context.Background(), int64(source.ID), time.Now())
	if s.log != nil {
		s.log.Info("collect finished",
			zap.Int64("source_id", int64(source.ID)),
			zap.Uint8("status", logRow.Status),
			zap.Int("collected", res.CollectCount),
		)
	}
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
				s.log.Warn("skip invalid cron", zap.Int64("source_id", int64(src.ID)), zap.Error(err))
			}
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
			if s.log != nil {
				s.log.Warn("add cron failed", zap.Int64("source_id", int64(sourceID)), zap.Error(err))
			}
			continue
		}
		s.cronIDs[int64(sourceID)] = entryID
		s.mu.Unlock()
	}
	return nil
}

// validateCollectPrecondition checks that a source is ready for collection.
// Called before launching the goroutine so errors are returned to the user.
func (s *collectService) validateCollectPrecondition(ctx context.Context, source *model.CollectSources) error {
	if source.Type != uint8(constant.CollectTypeAppleCMS) {
		return errcode.CollectDefaultNotSupported
	}
	maps, err := s.repo.ListCategories(ctx, int64(source.ID))
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if len(maps) == 0 {
		return errcode.CollectCategoryMapEmpty
	}
	if source.PlaySourceID <= 0 {
		return errcode.WithMessage(errcode.ParamError, "采集源未绑定播放源")
	}
	playSrc, err := s.playRepo.GetSource(ctx, int64(source.PlaySourceID))
	if err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	if playSrc == nil {
		return errcode.PlaySourceNotFound
	}
	return nil
}

func (s *collectService) validateSourceInput(ctx context.Context, typ uint8, collectURL, cronExpr string, playSourceID uint64) error {
	if typ != uint8(constant.CollectTypeDefault) && typ != uint8(constant.CollectTypeAppleCMS) {
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
	src, err := s.playRepo.GetSource(ctx, int64(playSourceID))
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

func toCollectSource(m *model.CollectSources) shareddto.CollectSourceItem {
	last := ""
	if m.LastCollectAt != nil {
		last = m.LastCollectAt.Format(time.DateTime)
	}
	return shareddto.CollectSourceItem{
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

func (s *collectService) FetchRemoteCategories(ctx context.Context, sourceID int64) (*admindto.RemoteCategoryResponse, error) {
	source, err := s.requireSource(ctx, sourceID)
	if err != nil {
		return nil, err
	}
	if source.Type != uint8(constant.CollectTypeAppleCMS) {
		return nil, errcode.CollectSourceNotAppleCMS
	}
	// Apple CMS class is returned on ac=list (and bare list URL), not on ac=detail.
	fetcher := collect.NewFetcher()
	body, err := fetcher.FetchAppleCMSCategories(ctx, source.CollectURL, source.APIKey)
	if err != nil {
		return nil, errcode.Wrap(errcode.CollectFetchFailed, err)
	}
	page, err := collect.ParseAppleCMS(body)
	if err != nil {
		return nil, errcode.Wrap(errcode.CollectParseFailed, err)
	}
	items := make([]admindto.RemoteCategoryItem, 0, len(page.Classes))
	for _, c := range page.Classes {
		if c.TypeID <= 0 {
			continue
		}
		items = append(items, admindto.RemoteCategoryItem{
			TypeID:   c.TypeID,
			TypeName: c.TypeName,
			TypePID:  c.TypePID,
		})
	}
	return &admindto.RemoteCategoryResponse{List: items}, nil
}

func (s *collectService) EnableSchedule(ctx context.Context, sourceID int64) error {
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
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return s.ReloadScheduler(ctx)
}

func (s *collectService) DisableSchedule(ctx context.Context, sourceID int64) error {
	source, err := s.requireSource(ctx, sourceID)
	if err != nil {
		return err
	}
	source.ScheduleEnabled = 0
	if err := s.repo.UpdateSource(ctx, source); err != nil {
		return errcode.Wrap(errcode.DatabaseError, err)
	}
	return s.ReloadScheduler(ctx)
}

func (s *collectService) CollectNow(ctx context.Context, sourceID int64, req *admindto.CollectNowRequest) error {
	source, err := s.requireSource(ctx, sourceID)
	if err != nil {
		return err
	}
	if err := s.validateCollectPrecondition(ctx, source); err != nil {
		return err
	}
	return s.startJob(source, req.DataRange)
}
