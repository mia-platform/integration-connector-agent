package azureactivitylogeventhubevents

type ActivityLogEventData struct {
	Records []*ActivityLogEventRecord `json:"records"`
}

type ActivityLogEventRecord struct {
	RoleLocation    string                    `json:"RoleLocation"`   //nolint:tagliatelle
	Stamp           string                    `json:"Stamp"`          //nolint:tagliatelle
	ReleaseVersion  string                    `json:"ReleaseVersion"` //nolint:tagliatelle
	Time            string                    `json:"time"`
	ResourceID      string                    `json:"resourceId"`
	OperationName   string                    `json:"operationName"`
	Category        string                    `json:"category"`
	ResultType      string                    `json:"resultType"`
	ResultSignature string                    `json:"resultSignature"`
	DurationMs      string                    `json:"durationMs"`
	CallerIPAddress string                    `json:"callerIpAddress"`
	CorrelationID   string                    `json:"correlationId"`
	Identity        *ActivityLogEventIdentity `json:"identity"`
	Level           string                    `json:"level"`
	Properties      map[string]any            `json:"properties"`
}

type ActivityLogEventIdentity struct {
	Authorization *ActivityLogAuthorization `json:"authorization"`
	Claims        map[string]string         `json:"claims"`
}

type ActivityLogAuthorization struct {
	Scope    string                `json:"scope"`
	Action   string                `json:"action"`
	Evidence AuthorizationEvidence `json:"evidence"`
}

type AuthorizationEvidence struct {
	Role                string `json:"role"`
	RoleAssignmentScope string `json:"roleAssignmentScope"`
	RoleAssignmentID    string `json:"roleAssignmentId"`
	RoleDefinitionID    string `json:"roleDefinitionId"`
	PrincipalID         string `json:"principalId"`
	PrincipalType       string `json:"principalType"`
}
