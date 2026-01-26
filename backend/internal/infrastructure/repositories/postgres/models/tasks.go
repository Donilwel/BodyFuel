package models

//type TaskRow struct {
//	Cluster     ClusterRow         `db:"cluster"`
//	UUID        uuid.UUID          `db:"task_id"`
//	TypeNm      entities.TaskType  `db:"task_type_nm"`
//	ClusterNm   string             `db:"task_cluster_nm"`
//	State       entities.TaskState `db:"task_state"`
//	Attempts    int                `db:"attempts"`
//	MaxAttempts int                `db:"max_attempts"`
//	RetryAt     time.Time          `db:"retry_at"`
//	Attribute   []byte             `db:"attribute"`
//	CreatedAt   time.Time          `db:"created_at"`
//	UpdatedAt   time.Time          `db:"updated_at"`
//}
//
//func NewTaskRow(t *entities.Task) (*TaskRow, error) {
//	attr, err := toRaw(t)
//	if err != nil {
//		return nil, fmt.Errorf(" to raw: %w", err)
//	}
//
//	return &TaskRow{
//		UUID:        t.UUID(),
//		TypeNm:      t.TypeNm(),
//		ClusterNm:   t.ClusterNm(),
//		State:       t.State(),
//		Attempts:    t.Attempts(),
//		MaxAttempts: t.MaxAttempts(),
//		Attribute:   attr,
//		RetryAt:     t.RetryAt(),
//		CreatedAt:   t.CreatedAt(),
//		UpdatedAt:   t.UpdatedAt(),
//	}, nil
//}
//
//func (r *TaskRow) ToEntity() (*entities.Task, error) {
//	attr, err := r.attributeToEntity()
//	if err != nil {
//		return nil, fmt.Errorf(" attribute to entity: %w", err)
//	}
//
//	return entities.NewTask(r.Cluster.ToEntity(), entities.WithTaskRestoreSpec(
//		entities.TaskRestoreSpecification{
//			UUID:        r.UUID,
//			TypeNm:      r.TypeNm,
//			ClusterNm:   r.ClusterNm,
//			State:       r.State,
//			Attempts:    r.Attempts,
//			MaxAttempts: r.MaxAttempts,
//			Attribute:   attr,
//			RetryAt:     r.RetryAt,
//			CreatedAt:   r.CreatedAt,
//			UpdatedAt:   r.UpdatedAt,
//		})), nil
//}
//
//func toRaw(t *entities.Task) (json.RawMessage, error) {
//	var attr any
//
//	switch t.TypeNm() {
//	case entities.TaskTypeFairyTerminateTask, entities.TaskTypeGPTerminateTask:
//		q := t.Attribute().(*entities.Query)
//		attr = queryTerminationAttribute{
//			ScrapeTs:               q.ScrapeTs(),
//			ProcPID:                q.ProcPID(),
//			SessionID:              q.SessionID(),
//			ClusterNm:              q.ClusterNm(),
//			DB:                     q.DB(),
//			Usename:                q.Usename(),
//			RsgName:                q.RsGroupName(),
//			QueryStart:             q.QueryStart(),
//			BackendStart:           q.BackendStart(),
//			ApplicationName:        q.ApplicationName(),
//			CurrentQuery:           q.CurrentQuery(),
//			CurrentQueryHash:       q.CurrentQueryHash(),
//			Waiting:                q.Waiting(),
//			CPUTime:                q.CPUTime(),
//			SpillSizeBytes:         q.SpillSizeBytes(),
//			Duration:               q.Duration(),
//			Latency:                q.Latency(),
//			Terminate:              q.Terminate(),
//			Idle:                   q.Idle(),
//			Terminated:             q.Terminated(),
//			TerminateFailure:       q.TerminateFailure(),
//			TerminateType:          q.TerminateType(),
//			RuleDuration:           q.RuleDuration(),
//			RuleCPUTime:            q.RuleCPUTime(),
//			RuleSpillSize:          q.RuleSpillSize(),
//			RuleID:                 q.RuleID(),
//			WhiteListID:            q.WhiteListIDs(),
//			MasksWhiteListID:       q.MasksWhiteListIDs(),
//			OtherRulesIDs:          q.OtherRulesIDs(),
//			NTerminationsBeforeBan: q.NTerminationsBeforeBan(),
//		}
//	case entities.TaskTypeBanUser, entities.TaskTypeUnbanUser:
//		b := t.Attribute().(*entities.TaskBanUserAttribute)
//		attr = taskBanUserAttribute{
//			Username:  b.Username,
//			ClusterNm: b.ClusterNm,
//			Host:      b.Host,
//		}
//	case entities.TaskTypeTrinoTerminateTask:
//		q := t.Attribute().(*entities.TrinoQuery)
//		attr = trinoQueryTerminationAttribute{
//			ScrapeTs:             q.ScrapeTs(),
//			ClusterNm:            q.ClusterNm(),
//			QueryID:              q.QueryID(),
//			User:                 q.User(),
//			QueryState:           q.QueryState(),
//			Query:                q.Query(),
//			Source:               q.Source(),
//			ResourceGroupID:      q.ResourceGroupID(),
//			CreateTime:           q.CreateTime(),
//			ExecutionStartTime:   q.ExecutionStartTime(),
//			LastHeartbeatTime:    q.LastHeartbeatTime(),
//			QueuedTimeMs:         q.QueuedTimeMs(),
//			AnalysisTimeMs:       q.AnalysisTimeMs(),
//			PlanningTimeMs:       q.PlanningTimeMs(),
//			Splits:               q.Splits(),
//			QueuedSplits:         q.QueuedSplits(),
//			RunningSplits:        q.RunningSplits(),
//			CompletedSplits:      q.CompletedSplits(),
//			SplitScheduledTimeMs: q.SplitScheduledTimeMs(),
//			SplitCpuTimeS:        q.SplitCPUTimeS(),
//			SplitBlockedTimeMs:   q.SplitBlockedTimeMs(),
//			RawInputBytes:        q.RawInputBytes(),
//			RawInputRows:         q.RawInputRows(),
//			ProcessedInputBytes:  q.ProcessedInputBytes(),
//			ProcessedInputRows:   q.ProcessedInputRows(),
//			OutputBytes:          q.OutputBytes(),
//			OutputRows:           q.OutputRows(),
//			PhysicalInputBytes:   q.PhysicalInputBytes(),
//			PhysicalWrittenBytes: q.PhysicalWrittenBytes(),
//			Terminate:            q.Terminate(),
//			Idle:                 q.Idle(),
//			Terminated:           q.Terminated(),
//			TerminateFailure:     q.TerminateFailure(),
//			TerminateType:        q.TerminateType(),
//			RuleDuration:         q.RuleDuration(),
//			RuleCPUTime:          q.RuleCPUTime(),
//			RuleInputBytes:       q.RuleInputBytes(),
//			RuleID:               q.RuleID(),
//			WhiteListIDs:         q.WhiteListIDs(),
//			MasksWhiteListIDs:    q.MasksWhiteListIDs(),
//			OtherRulesIDs:        q.OtherRulesIDs(),
//		}
//	default:
//		return nil, fmt.Errorf(" unsupported task type: %s", t.TypeNm())
//	}
//
//	raw, err := json.Marshal(attr)
//	if err != nil {
//		return nil, fmt.Errorf(" marshal attribute: %w", err)
//	}
//
//	return raw, nil
//}
//
//func (r *TaskRow) attributeToEntity() (any, error) {
//	var attr attribute
//
//	switch r.TypeNm {
//	case entities.TaskTypeFairyTerminateTask, entities.TaskTypeGPTerminateTask:
//		attr = new(queryTerminationAttribute)
//	case entities.TaskTypeBanUser, entities.TaskTypeUnbanUser:
//		attr = new(taskBanUserAttribute)
//	case entities.TaskTypeTrinoTerminateTask:
//		attr = new(trinoQueryTerminationAttribute)
//	default:
//		return nil, fmt.Errorf(" unsupported task type: %s", r.TypeNm)
//	}
//
//	if err := json.Unmarshal(r.Attribute, &attr); err != nil {
//		return nil, fmt.Errorf(" unmarshal attribute: %w", err)
//	}
//
//	return attr.ToEntity(), nil
//}
//
//type attribute interface {
//	ToEntity() any
//}
//
//type queryTerminationAttribute struct {
//	ScrapeTs               int64                       `json:"scrape_ts"`
//	ProcPID                int64                       `json:"proc_pid"`
//	SessionID              int64                       `json:"session_id"`
//	ClusterNm              string                      `json:"cluster_nm"`
//	DB                     string                      `json:"database"`
//	Usename                string                      `json:"usename"`
//	RsgName                string                      `json:"rs_group_name"`
//	QueryStart             int64                       `json:"query_start"`
//	BackendStart           int64                       `json:"backend_start"`
//	ApplicationName        string                      `json:"application_name"`
//	CurrentQuery           string                      `json:"current_query"`
//	CurrentQueryHash       string                      `json:"current_query_hash"`
//	Waiting                bool                        `json:"waiting"`
//	CPUTime                float64                     `json:"cpu_time"`
//	SpillSizeBytes         int64                       `json:"spill_size_bytes"`
//	Duration               uint64                      `json:"duration"`
//	Latency                int64                       `json:"latency"`
//	Terminate              bool                        `json:"terminate"`
//	Idle                   bool                        `json:"idle"`
//	Terminated             bool                        `json:"terminated"`
//	TerminateFailure       string                      `json:"terminate_failure"`
//	TerminateType          entities.QueryTerminateType `json:"terminate_type"`
//	RuleDuration           uint64                      `json:"rule_duration"`
//	RuleCPUTime            float64                     `json:"rule_cpu_time"`
//	RuleSpillSize          int64                       `json:"rule_spill_size"`
//	RuleID                 int                         `json:"rule_id"`
//	WhiteListID            []int                       `json:"whitelist_ids"`
//	MasksWhiteListID       []int                       `json:"masks_whitelist_ids"`
//	OtherRulesIDs          []int                       `json:"other_rules_ids"`
//	NTerminationsBeforeBan int64                       `json:"n_terminations_before_ban"`
//}
//
//func (r *queryTerminationAttribute) ToEntity() any {
//	return entities.NewQuery(entities.WithQueryRestoreSpec(entities.QueryRestoreSpec{
//		ScrapeTs:               r.ScrapeTs,
//		ProcPID:                r.ProcPID,
//		SessionID:              r.SessionID,
//		ClusterNm:              r.ClusterNm,
//		DB:                     r.DB,
//		Usename:                r.Usename,
//		RsgName:                r.RsgName,
//		QueryStart:             r.QueryStart,
//		BackendStart:           r.BackendStart,
//		ApplicationName:        r.ApplicationName,
//		CurrentQuery:           r.CurrentQuery,
//		CurrentQueryHash:       r.CurrentQueryHash,
//		Waiting:                r.Waiting,
//		CPUTime:                r.CPUTime,
//		SpillSizeBytes:         r.SpillSizeBytes,
//		Duration:               r.Duration,
//		Latency:                r.Latency,
//		Terminate:              r.Terminate,
//		Idle:                   r.Idle,
//		Terminated:             r.Terminated,
//		TerminateFailure:       r.TerminateFailure,
//		TerminateType:          r.TerminateType,
//		RuleDuration:           r.RuleDuration,
//		RuleCPUTime:            r.RuleCPUTime,
//		RuleSpillSize:          r.RuleSpillSize,
//		RuleID:                 r.RuleID,
//		WhiteListIDs:           r.WhiteListID,
//		MasksWhiteListIDs:      r.MasksWhiteListID,
//		OtherRulesIDs:          r.OtherRulesIDs,
//		NTerminationsBeforeBan: r.NTerminationsBeforeBan,
//	}))
//}
//
//type taskBanUserAttribute struct {
//	Username  string `json:"username"`
//	ClusterNm string `json:"cluster_nm"`
//	Host      string `json:"host"`
//}
//
//func (r *taskBanUserAttribute) ToEntity() any {
//	return &entities.TaskBanUserAttribute{
//		Username:  r.Username,
//		ClusterNm: r.ClusterNm,
//		Host:      r.Host,
//	}
//}
