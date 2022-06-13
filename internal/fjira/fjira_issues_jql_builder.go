package fjira

import (
	"fmt"
	"github.com/mk5/fjira/internal/jira"
	"strings"
)

func buildSearchIssuesJql(project *jira.JiraProject, query string, status *jira.JiraIssueStatus, user *jira.JiraUser) string {
	jql := fmt.Sprintf("project=%s", project.Id)
	orderBy := "ORDER BY status"
	query = strings.TrimSpace(query)
	if query != "" {
		jql = jql + fmt.Sprintf(" AND summary~\"%s*\"", query)
	}
	if status != nil && status.Name != MessageAll {
		jql = jql + fmt.Sprintf(" AND status=%s", status.Id)
	}
	if user != nil && user.DisplayName != MessageAll {
		jql = jql + fmt.Sprintf(" AND assignee=%s", user.AccountId)
	}
	if query != "" && issueRegExp.MatchString(query) {
		jql = jql + fmt.Sprintf(" OR issuekey=\"%s\"", query)
	}
	return fmt.Sprintf("%s %s", jql, orderBy)
}