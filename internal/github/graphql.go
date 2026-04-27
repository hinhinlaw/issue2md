package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GraphQL query strings
const (
	issueQuery = `
	query($owner: String!, $repo: String!, $number: Int!) {
		repository(owner: $owner, name: $repo) {
			issue(number: $number) {
				title
				body
				number
				url
				author { login }
				comments(first: 100) {
					nodes {
						body
						databaseId
						author { login }
						reactions(first: 100) {
							nodes {
								content
								user { login }
							}
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
				reactions(first: 100) {
					nodes {
						content
						user { login }
					}
				}
			}
		}
	}`

	pullRequestQuery = `
	query($owner: String!, $repo: String!, $number: Int!) {
		repository(owner: $owner, name: $repo) {
			pullRequest(number: $number) {
				title
				body
				number
				url
				author { login }
				comments(first: 100) {
					nodes {
						body
						databaseId
						author { login }
						reactions(first: 100) {
							nodes {
								content
								user { login }
							}
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
				reactions(first: 100) {
					nodes {
						content
						user { login }
					}
				}
			}
		}
	}`

	discussionQuery = `
	query($owner: String!, $repo: String!, $number: Int!) {
		repository(owner: $owner, name: $repo) {
			discussion(number: $number) {
				title
				body
				number
				url
				author { login }
				comments(first: 100) {
					nodes {
						body
						databaseId
						author { login }
						reactions(first: 100) {
							nodes {
								content
								user { login }
							}
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	}`

	issueCommentsQuery = `
	query($owner: String!, $repo: String!, $number: Int!, $after: String) {
		repository(owner: $owner, name: $repo) {
			issue(number: $number) {
				comments(first: 100, after: $after) {
					nodes {
						body
						databaseId
						author { login }
						reactions(first: 100) {
							nodes {
								content
								user { login }
							}
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	}`

	pullRequestCommentsQuery = `
	query($owner: String!, $repo: String!, $number: Int!, $after: String) {
		repository(owner: $owner, name: $repo) {
			pullRequest(number: $number) {
				comments(first: 100, after: $after) {
					nodes {
						body
						databaseId
						author { login }
						reactions(first: 100) {
							nodes {
								content
								user { login }
							}
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	}`

	discussionCommentsQuery = `
	query($owner: String!, $repo: String!, $number: Int!, $after: String) {
		repository(owner: $owner, name: $repo) {
			discussion(number: $number) {
				comments(first: 100, after: $after) {
					nodes {
						body
						databaseId
						author { login }
						reactions(first: 100) {
							nodes {
								content
								user { login }
							}
						}
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	}`
)

const (
	graphQLEndpoint = "https://api.github.com/graphql"
	apiVersion      = "2022-11-28"
)

// GraphQLQuery represents a GraphQL request
type GraphQLQuery struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   json.RawMessage  `json:"data"`
	Errors []GraphQLError   `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string        `json:"message"`
	Path    []interface{} `json:"path,omitempty"`
}

// executeGraphQL sends a GraphQL query to GitHub's API
func executeGraphQL(query string, variables map[string]interface{}, token string) (*GraphQLResponse, error) {
	queryBody := GraphQLQuery{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(queryBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL query: %w", err)
	}

	req, err := http.NewRequest("POST", graphQLEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", apiVersion)

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := sharedHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub GraphQL API returned status: %s, body: %s", resp.Status, string(body))
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GraphQL response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %v", gqlResp.Errors)
	}

	return &gqlResp, nil
}

// GraphQL response types

type GraphQLIssueResponse struct {
	Repository struct {
		Issue struct {
			Title     string `json:"title"`
			Body      string `json:"body"`
			Number    int    `json:"number"`
			URL       string `json:"url"`
			Author    struct {
				Login string `json:"login"`
			} `json:"author"`
			Comments struct {
				Nodes []struct {
					Body      string `json:"body"`
					DatabaseID int    `json:"databaseId"`
					Author    struct {
						Login string `json:"login"`
					} `json:"author"`
					Reactions struct {
						Nodes []struct {
							Content string `json:"content"`
							User    struct {
								Login string `json:"login"`
							} `json:"user"`
						} `json:"nodes"`
					} `json:"reactions"`
				} `json:"nodes"`
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
			} `json:"comments"`
			Reactions struct {
				Nodes []struct {
					Content string `json:"content"`
					User    struct {
						Login string `json:"login"`
					} `json:"user"`
				} `json:"nodes"`
			} `json:"reactions"`
		} `json:"issue"`
	} `json:"repository"`
}

type GraphQLPullRequestResponse struct {
	Repository struct {
		PullRequest struct {
			Title     string `json:"title"`
			Body      string `json:"body"`
			Number    int    `json:"number"`
			URL       string `json:"url"`
			Author    struct {
				Login string `json:"login"`
			} `json:"author"`
			Comments struct {
				Nodes []struct {
					Body      string `json:"body"`
					DatabaseID int    `json:"databaseId"`
					Author    struct {
						Login string `json:"login"`
					} `json:"author"`
					Reactions struct {
						Nodes []struct {
							Content string `json:"content"`
							User    struct {
								Login string `json:"login"`
							} `json:"user"`
						} `json:"nodes"`
					} `json:"reactions"`
				} `json:"nodes"`
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
			} `json:"comments"`
			Reactions struct {
				Nodes []struct {
					Content string `json:"content"`
					User    struct {
						Login string `json:"login"`
					} `json:"user"`
				} `json:"nodes"`
			} `json:"reactions"`
		} `json:"pullRequest"`
	} `json:"repository"`
}

type GraphQLDiscussionResponse struct {
	Repository struct {
		Discussion struct {
			Title     string `json:"title"`
			Body      string `json:"body"`
			Number    int    `json:"number"`
			URL       string `json:"url"`
			Author    struct {
				Login string `json:"login"`
			} `json:"author"`
			Comments struct {
				Nodes []struct {
					Body      string `json:"body"`
					DatabaseID int    `json:"databaseId"`
					Author    struct {
						Login string `json:"login"`
					} `json:"author"`
					Reactions struct {
						Nodes []struct {
							Content string `json:"content"`
							User    struct {
								Login string `json:"login"`
							} `json:"user"`
						} `json:"nodes"`
					} `json:"reactions"`
				} `json:"nodes"`
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
			} `json:"comments"`
		} `json:"discussion"`
	} `json:"repository"`
}

// Comments-only response types for pagination support

type GraphQLIssueCommentsResponse struct {
	Repository struct {
		Issue struct {
			Comments struct {
				Nodes []struct {
					Body       string `json:"body"`
					DatabaseID int    `json:"databaseId"`
					Author     struct {
						Login string `json:"login"`
					} `json:"author"`
					Reactions struct {
						Nodes []struct {
							Content string `json:"content"`
							User    struct {
								Login string `json:"login"`
							} `json:"user"`
						} `json:"nodes"`
					} `json:"reactions"`
				} `json:"nodes"`
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
			} `json:"comments"`
		} `json:"issue"`
	} `json:"repository"`
}

type GraphQLPullRequestCommentsResponse struct {
	Repository struct {
		PullRequest struct {
			Comments struct {
				Nodes []struct {
					Body       string `json:"body"`
					DatabaseID int    `json:"databaseId"`
					Author     struct {
						Login string `json:"login"`
					} `json:"author"`
					Reactions struct {
						Nodes []struct {
							Content string `json:"content"`
							User    struct {
								Login string `json:"login"`
							} `json:"user"`
						} `json:"nodes"`
					} `json:"reactions"`
				} `json:"nodes"`
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
			} `json:"comments"`
		} `json:"pullRequest"`
	} `json:"repository"`
}

type GraphQLDiscussionCommentsResponse struct {
	Repository struct {
		Discussion struct {
			Comments struct {
				Nodes []struct {
					Body       string `json:"body"`
					DatabaseID int    `json:"databaseId"`
					Author     struct {
						Login string `json:"login"`
					} `json:"author"`
					Reactions struct {
						Nodes []struct {
							Content string `json:"content"`
							User    struct {
								Login string `json:"login"`
							} `json:"user"`
						} `json:"nodes"`
					} `json:"reactions"`
				} `json:"nodes"`
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
			} `json:"comments"`
		} `json:"discussion"`
	} `json:"repository"`
}

// Conversion functions from GraphQL to existing types

func graphQLIssueToIssue(gqlResp *GraphQLIssueResponse) *Issue {
	gqlIssue := gqlResp.Repository.Issue
	issue := &Issue{
		Title:    gqlIssue.Title,
		Body:     gqlIssue.Body,
		Number:   gqlIssue.Number,
		URL:      gqlIssue.URL,
		User:     User{Login: gqlIssue.Author.Login},
		Comments: len(gqlIssue.Comments.Nodes),
	}
	return issue
}

func graphQLCommentsFromIssue(gqlResp *GraphQLIssueResponse) []Comment {
	var comments []Comment
	for _, node := range gqlResp.Repository.Issue.Comments.Nodes {
		comment := Comment{
			Body: node.Body,
			ID:   node.DatabaseID,
			User: User{Login: node.Author.Login},
		}
		for _, r := range node.Reactions.Nodes {
			comment.Reactions = append(comment.Reactions, Reaction{
				Content: r.Content,
				User:    User{Login: r.User.Login},
			})
		}
		comments = append(comments, comment)
	}
	return comments
}

func graphQLIssueReactions(gqlResp *GraphQLIssueResponse) []Reaction {
	var reactions []Reaction
	for _, node := range gqlResp.Repository.Issue.Reactions.Nodes {
		reactions = append(reactions, Reaction{
			Content: node.Content,
			User:    User{Login: node.User.Login},
		})
	}
	return reactions
}

func graphQLPRToPullRequest(gqlResp *GraphQLPullRequestResponse) *PullRequest {
	gqlPR := gqlResp.Repository.PullRequest
	pr := &PullRequest{
		Title:    gqlPR.Title,
		Body:     gqlPR.Body,
		Number:   gqlPR.Number,
		URL:      gqlPR.URL,
		User:     User{Login: gqlPR.Author.Login},
		Comments: len(gqlPR.Comments.Nodes),
	}
	return pr
}

func graphQLCommentsFromPR(gqlResp *GraphQLPullRequestResponse) []Comment {
	var comments []Comment
	for _, node := range gqlResp.Repository.PullRequest.Comments.Nodes {
		comment := Comment{
			Body: node.Body,
			ID:   node.DatabaseID,
			User: User{Login: node.Author.Login},
		}
		for _, r := range node.Reactions.Nodes {
			comment.Reactions = append(comment.Reactions, Reaction{
				Content: r.Content,
				User:    User{Login: r.User.Login},
			})
		}
		comments = append(comments, comment)
	}
	return comments
}

func graphQLPRReactions(gqlResp *GraphQLPullRequestResponse) []Reaction {
	var reactions []Reaction
	for _, node := range gqlResp.Repository.PullRequest.Reactions.Nodes {
		reactions = append(reactions, Reaction{
			Content: node.Content,
			User:    User{Login: node.User.Login},
		})
	}
	return reactions
}

func graphQLDiscussionToDiscussion(gqlResp *GraphQLDiscussionResponse) *Discussion {
	gqlDisc := gqlResp.Repository.Discussion
	disc := &Discussion{
		Title:    gqlDisc.Title,
		Body:     gqlDisc.Body,
		Number:   gqlDisc.Number,
		URL:      gqlDisc.URL,
		User:     User{Login: gqlDisc.Author.Login},
		Comments: len(gqlDisc.Comments.Nodes),
	}
	return disc
}

func graphQLCommentsFromDiscussion(gqlResp *GraphQLDiscussionResponse) []DiscussionComment {
	var comments []DiscussionComment
	for _, node := range gqlResp.Repository.Discussion.Comments.Nodes {
		comment := DiscussionComment{
			Body: node.Body,
			ID:   node.DatabaseID,
			User: User{Login: node.Author.Login},
		}
		for _, r := range node.Reactions.Nodes {
			comment.Reactions = append(comment.Reactions, Reaction{
				Content: r.Content,
				User:    User{Login: r.User.Login},
			})
		}
		comments = append(comments, comment)
	}
	return comments
}

func graphQLIssueCommentsResponseToComments(resp *GraphQLIssueCommentsResponse) []Comment {
	var comments []Comment
	for _, node := range resp.Repository.Issue.Comments.Nodes {
		comment := Comment{
			Body: node.Body,
			ID:   node.DatabaseID,
			User: User{Login: node.Author.Login},
		}
		for _, r := range node.Reactions.Nodes {
			comment.Reactions = append(comment.Reactions, Reaction{
				Content: r.Content,
				User:    User{Login: r.User.Login},
			})
		}
		comments = append(comments, comment)
	}
	return comments
}

func graphQLPRCommentsResponseToComments(resp *GraphQLPullRequestCommentsResponse) []Comment {
	var comments []Comment
	for _, node := range resp.Repository.PullRequest.Comments.Nodes {
		comment := Comment{
			Body: node.Body,
			ID:   node.DatabaseID,
			User: User{Login: node.Author.Login},
		}
		for _, r := range node.Reactions.Nodes {
			comment.Reactions = append(comment.Reactions, Reaction{
				Content: r.Content,
				User:    User{Login: r.User.Login},
			})
		}
		comments = append(comments, comment)
	}
	return comments
}

func graphQLDiscussionCommentsResponseToComments(resp *GraphQLDiscussionCommentsResponse) []DiscussionComment {
	var comments []DiscussionComment
	for _, node := range resp.Repository.Discussion.Comments.Nodes {
		comment := DiscussionComment{
			Body: node.Body,
			ID:   node.DatabaseID,
			User: User{Login: node.Author.Login},
		}
		for _, r := range node.Reactions.Nodes {
			comment.Reactions = append(comment.Reactions, Reaction{
				Content: r.Content,
				User:    User{Login: r.User.Login},
			})
		}
		comments = append(comments, comment)
	}
	return comments
}
