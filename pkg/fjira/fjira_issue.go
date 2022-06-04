package fjira

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/mk5/fjira/internal/app"
	"github.com/mk5/fjira/internal/jira"
	"math"
)

type fjiraIssueView struct {
	app.View
	bottomBar            *app.ActionBar
	topBar               *app.ActionBar
	fuzzyFind            *app.FuzzyFind
	issue                *jira.JiraIssue
	descriptionLimitX    int
	descriptionLimitY    int
	descriptionBodyLines int
	commentsLines        int
	scrollY              int
	descriptionLines     int
	maxScrollY           int
}

func NewIssueView(issue *jira.JiraIssue) *fjiraIssueView {
	bottomBar := CreateIssueBottomBar(issue)
	bottomBar.AddItem(NewStatusChangeBarItem())
	bottomBar.AddItem(NewAssigneeChangeBarItem())
	bottomBar.AddItem(CreateCommentBarItem())
	bottomBar.AddItem(NewOpenBarItem())
	bottomBar.AddItem(NewCancelBarItem())

	issueActionBar := CreateIssueTopBar(issue)

	return &fjiraIssueView{
		bottomBar: bottomBar,
		topBar:    issueActionBar,
		issue:     issue,
		scrollY:   0,
	}
}

func (view *fjiraIssueView) Init() {
	go view.handleIssueAction()
}

func (view *fjiraIssueView) Destroy() {
}

func (view *fjiraIssueView) Draw(screen tcell.Screen) {
	if view.fuzzyFind == nil {
		app.DrawText(screen, 1, 3-view.scrollY, tcell.StyleDefault, view.issue.Fields.Summary)
		for col := 1; col <= len(view.issue.Fields.Summary); col++ {
			screen.SetContent(col, 4-view.scrollY, tcell.RuneHLine, nil, tcell.StyleDefault)
		}
		rows := app.DrawTextLimited(screen, 1, 6-view.scrollY, view.descriptionLimitX, view.descriptionLimitY, tcell.StyleDefault, view.issue.Fields.Description)
		app.DrawText(screen, 1, 6+rows-view.scrollY, tcell.StyleDefault, "Comments")
		for col := 1; col <= 60; col++ {
			screen.SetContent(col, 6+rows+1-view.scrollY, tcell.RuneHLine, nil, tcell.StyleDefault)
		}
	}
	if len(view.issue.Fields.Comments) > 0 {

	}
	view.bottomBar.Draw(screen)
	view.topBar.Draw(screen)
	if view.fuzzyFind != nil {
		view.fuzzyFind.Draw(screen)
	}
}

func (view *fjiraIssueView) Update() {
	view.bottomBar.Update()
	view.topBar.Update()
	if view.fuzzyFind != nil {
		view.fuzzyFind.Update()
	}
}

func (view *fjiraIssueView) Resize(screenX, screenY int) {
	view.descriptionLimitX = app.ClampInt(int(math.Floor(float64(screenX)*0.9)), 1, 10000)
	view.descriptionLimitY = screenY - 6
	view.descriptionBodyLines = int(math.Ceil(float64(len(view.issue.Fields.Description) / view.descriptionLimitX)))
	if len(view.issue.Fields.Comments) > 0 {
		view.commentsLines = 2
		for _, comment := range view.issue.Fields.Comments {
			view.commentsLines = view.commentsLines + int(math.Ceil(float64(len(comment.Body)/view.descriptionLimitX)))
		}
	}
	view.descriptionLines = view.descriptionBodyLines + view.commentsLines
	view.maxScrollY = app.ClampInt(-((view.descriptionLimitY - 6 - 6) - view.descriptionLines), 0, 1000)
	if view.maxScrollY >= view.descriptionLimitY {
		view.maxScrollY += 1
	}
	view.bottomBar.Resize(screenX, screenY)
	view.topBar.Resize(screenX, screenY)
	if view.fuzzyFind != nil {
		view.fuzzyFind.Resize(screenX, screenY)
	}
}

func (view *fjiraIssueView) HandleKeyEvent(ev *tcell.EventKey) {
	view.bottomBar.HandleKeyEvent(ev)
	view.topBar.HandleKeyEvent(ev)
	if view.fuzzyFind != nil {
		view.fuzzyFind.HandleKeyEvent(ev)
	}
	if ev.Key() == tcell.KeyUp {
		view.scrollY = app.ClampInt(view.scrollY-1, 0, view.maxScrollY)
	}
	if ev.Key() == tcell.KeyDown {
		view.scrollY = app.ClampInt(view.scrollY+1, 0, view.maxScrollY)
	}
}

func (view *fjiraIssueView) handleIssueAction() {
	select {
	case selectedAction := <-view.bottomBar.Action:
		switch selectedAction {
		case ActionEscape:
			app.GetApp().SetView(NewIssuesSearchView(&view.issue.Fields.Project))
			return
		case ActionStatusChange:
			goIntoChangeStatus(view.issue)
			return
		case ActionAssigneeChange:
			goIntoChangeAssignment(view.issue)
			return
		case ActionComment:
			goIntoCommentView(view.issue)
			return
		case ActionOpen:
			jiraUrl, _ := GetJiraUrl()
			app.OpenLink(fmt.Sprintf("%s/browse/%s", jiraUrl, view.issue.Key))
			go view.handleIssueAction()
			return
		}
	}
}
