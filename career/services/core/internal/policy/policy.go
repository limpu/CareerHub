package policy

type ActionKind string

const (
	ActionPostContent  ActionKind = "post_content"
	ActionSendDM       ActionKind = "send_dm"
	ActionApplyJob     ActionKind = "apply_job"
	ActionExportSheets ActionKind = "export_sheets"
)

type ApprovalGate struct {
	Required       bool   `json:"required"`
	ActorID        string `json:"actor_id"`
	PayloadHash    string `json:"payload_hash"`
	ApprovedByUser bool   `json:"approved_by_user"`
}
