package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Issue struct {
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Number    int        `json:"number"`
	URL       string     `json:"html_url"` // Use html_url for consistency
	Comments  int        `json:"comments"`
	User      User       `json:"user"`
	Reactions []Reaction `json:"-"`
}

func (i *Issue) ItemNumber() int {
	return i.Number
}

func (i *Issue) SetReactions(reactions []Reaction) {
	i.Reactions = reactions
}

type PullRequest struct {
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Number    int        `json:"number"`
	URL       string     `json:"html_url"` // Use html_url for the web link
	Comments  int        `json:"comments"`
	User      User       `json:"user"`
	Reactions []Reaction `json:"-"`
}

func (pr *PullRequest) ItemNumber() int {
	return pr.Number
}

func (pr *PullRequest) SetReactions(reactions []Reaction) {
	pr.Reactions = reactions
}

type Discussion struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Number   int    `json:"number"`
	URL      string `json:"html_url"` // Use html_url for discussions
	Comments int    `json:"comments_count"`
	User     User   `json:"user"`
}

type DiscussionComment struct {
	Body      string     `json:"body"`
	User      User       `json:"user"`
	ID        int        `json:"id"`
	Reactions []Reaction `json:"-"`
}

type Comment struct {
	Body      string     `json:"body"`
	User      User       `json:"user"`
	ID        int        `json:"id"`
	Reactions []Reaction `json:"-"`
}

type Reaction struct {
	Content string `json:"content"`
	User    User   `json:"user"`
}

type User struct {
	Login string `json:"login"`
}

type reactionable interface {
	ItemNumber() int
	SetReactions([]Reaction)
}

var sharedHTTPClient = &http.Client{}

func ParseURL(issueURL string) (owner, repo string, number int, issueType string, err error) {
	parsedURL, err := url.Parse(issueURL)
	if err != nil {
		return
	}

	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(parts) < 4 {
		err = fmt.Errorf("invalid GitHub URL format")
		return
	}

	if parts[2] == "issues" && len(parts) == 4 {
		issueType = "issue"
	} else if parts[2] == "discussions" && len(parts) == 4 {
		issueType = "discussion"
	} else if parts[2] == "pull" && len(parts) == 4 {
		issueType = "pull"
	} else {
		err = fmt.Errorf("invalid GitHub issue, discussion, or pull request URL format")
		return
	}
	owner = parts[0]
	repo = parts[1]
	issueNumber, err := strconv.Atoi(parts[3])
	if err != nil {
		return
	}

	return owner, repo, issueNumber, issueType, nil
}

func fetchAndSetReactions(item reactionable, owner, repo, token string) error {
	reactions, err := FetchReactionsForPullRequestOrIssue(owner, repo, item.ItemNumber(), token)
	if err != nil {
		return fmt.Errorf("failed to fetch reactions in %s/%s for item %d: %v. Ensure you have set a valid GITHUB_TOKEN", owner, repo, item.ItemNumber(), err)
	}
	item.SetReactions(reactions)
	return nil
}

func FetchIssue(owner, repo string, issueNumber int, token string, enableReactions bool) (*Issue, error) {
	variables := map[string]interface{}{
		"owner":  owner,
		"repo":   repo,
		"number": issueNumber,
	}

	gqlResp, err := executeGraphQL(issueQuery, variables, token)
	if err != nil {
		return nil, err
	}

	var issueResp GraphQLIssueResponse
	if err := json.Unmarshal(gqlResp.Data, &issueResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal issue response: %w", err)
	}

	issue := graphQLIssueToIssue(&issueResp)

	if enableReactions {
		reactions := graphQLIssueReactions(&issueResp)
		issue.SetReactions(reactions)
	}

	return issue, nil
}

func FetchPullRequest(owner, repo string, pullNumber int, token string, enableReactions bool) (*PullRequest, error) {
	variables := map[string]interface{}{
		"owner":  owner,
		"repo":   repo,
		"number": pullNumber,
	}

	gqlResp, err := executeGraphQL(pullRequestQuery, variables, token)
	if err != nil {
		return nil, err
	}

	var prResp GraphQLPullRequestResponse
	if err := json.Unmarshal(gqlResp.Data, &prResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pull request response: %w", err)
	}

	pr := graphQLPRToPullRequest(&prResp)

	if enableReactions {
		reactions := graphQLPRReactions(&prResp)
		pr.SetReactions(reactions)
	}

	return pr, nil
}

func FetchComments(owner, repo string, issueNumber int, token string, enableReactions bool, enableUserLinks bool) ([]Comment, error) {
	var allComments []Comment
	var after string

	for {
		variables := map[string]interface{}{
			"owner":  owner,
			"repo":   repo,
			"number": issueNumber,
		}
		if after != "" {
			variables["after"] = after
		}

		gqlResp, err := executeGraphQL(issueCommentsQuery, variables, token)
		if err != nil {
			return nil, err
		}

		var commentsResp GraphQLIssueCommentsResponse
		if err := json.Unmarshal(gqlResp.Data, &commentsResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal comments response: %w", err)
		}

		comments := graphQLIssueCommentsResponseToComments(&commentsResp)
		allComments = append(allComments, comments...)

		if !commentsResp.Repository.Issue.Comments.PageInfo.HasNextPage {
			break
		}
		after = commentsResp.Repository.Issue.Comments.PageInfo.EndCursor
	}

	return allComments, nil
}

func fetchReactions(url, token string) ([]Reaction, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.squirrel-girl-preview+json")

	resp, err := sharedHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var reactions []Reaction
	if err := json.NewDecoder(resp.Body).Decode(&reactions); err != nil {
		return nil, err
	}
	return reactions, nil
}

func FetchReactionsForPullRequestOrIssue(owner, repo string, prOrIssueNumber int, token string) ([]Reaction, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/reactions", owner, repo, prOrIssueNumber)
	return fetchReactions(url, token)
}

func FetchReactionsForComment(owner, repo string, commentID int, token string) ([]Reaction, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/comments/%d/reactions", owner, repo, commentID)
	return fetchReactions(url, token)
}

func FetchDiscussion(owner, repo string, discussionNumber int, token string) (*Discussion, error) {
	variables := map[string]interface{}{
		"owner":  owner,
		"repo":   repo,
		"number": discussionNumber,
	}

	gqlResp, err := executeGraphQL(discussionQuery, variables, token)
	if err != nil {
		return nil, err
	}

	var discResp GraphQLDiscussionResponse
	if err := json.Unmarshal(gqlResp.Data, &discResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal discussion response: %w", err)
	}

	discussion := graphQLDiscussionToDiscussion(&discResp)
	return discussion, nil
}

func FetchDiscussionComments(owner, repo string, discussionNumber int, token string, enableReactions bool) ([]DiscussionComment, error) {
	var allComments []DiscussionComment
	var after string

	for {
		variables := map[string]interface{}{
			"owner":  owner,
			"repo":   repo,
			"number": discussionNumber,
		}
		if after != "" {
			variables["after"] = after
		}

		gqlResp, err := executeGraphQL(discussionCommentsQuery, variables, token)
		if err != nil {
			return nil, err
		}

		var commentsResp GraphQLDiscussionCommentsResponse
		if err := json.Unmarshal(gqlResp.Data, &commentsResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal discussion comments response: %w", err)
		}

		comments := graphQLDiscussionCommentsResponseToComments(&commentsResp)
		allComments = append(allComments, comments...)

		if !commentsResp.Repository.Discussion.Comments.PageInfo.HasNextPage {
			break
		}
		after = commentsResp.Repository.Discussion.Comments.PageInfo.EndCursor
	}

	return allComments, nil
}
