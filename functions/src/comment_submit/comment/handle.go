package comment

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v37/github"
)

type FormResponse struct {
	Comment *Comment
	Links   FormResponseLinks
}

type FormResponseLinks struct {
	PullRequest string
	CommentFile string
}

type Form struct {
	Fields struct {
		ReplyThread string `json:"replyThread"`
		ReplyID     string `json:"replyID"`
		ReplyName   string `json:"replyName"`
		Name        string `json:"name"`
		Website     string `json:"website"`
		Email       string `json:"email"`
		Body        string `json:"body"`
	} `json:"fields"`
	Options struct {
		EntryID string `json:"entryId"`
	} `json:"options"`
}
type Comment struct {
	ID          string    `json:"_id"`
	EntryID     string    `json:"entryId"`
	Name        string    `json:"name"`
	ReplyThread string    `json:"replyThread"`
	ReplyID     string    `json:"replyId"`
	ReplyName   string    `json:"replyName"`
	Website     string    `json:"website"`
	Email       string    `json:"email"`
	Date        time.Time `json:"date"`
	Body        string    `json:"body"`
}

func EnsurePRBranch(base string, entryId string, prID *int64) (string, error) {
	branchName := fmt.Sprintf("cmt_%s_%s", base, entryId)

	// create branch if not exist
	_, _, err := githubClient.Git.GetRef(context.Background(), owner, repo, "heads/"+branchName)
	if err != nil {
		ref, _, err := githubClient.Git.GetRef(context.Background(), owner, repo, "heads/"+base)
		if err != nil {
			return "", err
		}

		_, _, err = githubClient.Git.CreateRef(context.Background(), owner, repo, &github.Reference{
			Ref: refStr("heads/" + branchName),
			Object: &github.GitObject{
				SHA: refStr(*ref.Object.SHA),
			},
		})
		if err != nil {
			return "", err
		}
	}

	if prID != nil {
		prList, _, err := githubClient.PullRequests.List(context.Background(), owner, repo, &github.PullRequestListOptions{
			Base: base,
			Head: fmt.Sprintf("%s:%s", owner, branchName),
		})
		if err != nil {
			return branchName, err
		}
		for _, pr := range prList {
			if pr.Title != nil && strings.Contains(*pr.Title, fmt.Sprintf("post %s", entryId)) && strings.Contains(*pr.Title, fmt.Sprintf("[%s]", base)) {
				*prID = *pr.ID
				return branchName, nil
			}
		}

		pr, _, err := githubClient.PullRequests.Create(context.Background(), owner, repo, &github.NewPullRequest{
			Title: refStr(fmt.Sprintf("[%s] Comment on post %s", base, entryId)),
			Base:  refStr(base),
			Head:  &branchName,
			Body:  refStr("This is an auto generated PR for comments."),
		})
		if err != nil {
			return branchName, err
		}
		*prID = *pr.ID
	}
	return branchName, nil
}

func validateNewComment(comment *Comment, existingComments []Comment) error {
	if comment.Name == "" {
		return errors.New("you must provide a name")
	}
	if comment.Body == "" {
		return errors.New("you must provide a comment body")
	}
	if comment.ReplyThread == "" {
		comment.ReplyID = ""
		comment.ReplyName = ""
	} else {
		replyName := ""
		for _, c := range existingComments {
			if c.ID == comment.ReplyThread {
				replyName = c.Name
			}
		}
		if replyName == "" {
			return errors.New("the thread you are replying to does not exist")
		}
		if comment.ReplyID == "" {
			comment.ReplyName = replyName
		} else {
			replyName = ""
			for _, c := range existingComments {
				if c.ID == comment.ReplyID {
					replyName = c.Name
				}
			}
			if replyName == "" {
				return errors.New("the comment you are replying to does not exist")
			}
			comment.ReplyID = replyName
		}
	}
	return nil
}

func HandleForm(form Form, origin string) (*FormResponse, bool, error) {
	postToBranch := "dev"
	if origin == "yumechi.jp" || origin == "yumechi.jp:443" {
		postToBranch = "main"
	}

	entryId := form.Options.EntryID
	if entryId == "" {
		return nil, true, fmt.Errorf("missing entry id")
	}

	prBranch, err := EnsurePRBranch(postToBranch, entryId, nil)
	if err != nil {
		return nil, false, err
	}

	commentFile := fmt.Sprintf("data/comments/%s.json", entryId)
	curFile, _, err := githubClient.Repositories.DownloadContents(context.Background(), owner, repo, commentFile, &github.RepositoryContentGetOptions{
		Ref: prBranch,
	})

	curFile.Close()

	if err != nil {
		if err, ok := err.(*github.ErrorResponse); !ok ||
			err.Response == nil ||
			err.Response.StatusCode != 404 {
			return nil, false, err
		}
	}

	var existingComments []Comment
	var curFileBytes []byte
	if curFile != nil {
		curFileBytes, err = io.ReadAll(curFile)
		if err != nil {
			return nil, false, err
		}
		if err := json.Unmarshal(curFileBytes, &existingComments); err != nil {
			return nil, false, err
		}
	}
	newComment := Comment{
		ID:          strconv.FormatInt(time.Now().UnixNano(), 16),
		EntryID:     entryId,
		Name:        form.Fields.Name,
		ReplyThread: form.Fields.ReplyThread,
		ReplyID:     form.Fields.ReplyID,
		ReplyName:   form.Fields.ReplyName,
		Website:     form.Fields.Website,
		Email:       form.Fields.Email,
		Date:        time.Now(),
		Body:        form.Fields.Body,
	}
	if err := validateNewComment(&newComment, existingComments); err != nil {
		return nil, true, err
	}
	existingComments = append(existingComments, newComment)

	newGhContentBytes, err := json.MarshalIndent(existingComments, "", "\t")
	if err != nil {
		return nil, false, err
	}
	newGhContent := &github.RepositoryContentFileOptions{
		Message: refStr(fmt.Sprintf("New comment for %s", entryId)),
		Content: newGhContentBytes,
	}
	if curFile != nil {
		newGhContent.SHA = refStr(hex.EncodeToString(curFileBytes))
	}

	newContentResponse, _, err := githubClient.Repositories.UpdateFile(context.Background(), owner, repo, commentFile, newGhContent)
	if err != nil {
		return nil, false, err
	}

	var prId int64
	_, err = EnsurePRBranch(postToBranch, entryId, &prId)
	if err != nil {
		return nil, false, err
	}

	return &FormResponse{
		Comment: &newComment,
		Links:   FormResponseLinks{CommentFile: *newContentResponse.HTMLURL, PullRequest: fmt.Sprintf("https://github.com/%s/%s/%d", owner, repo, prId)},
	}, false, nil
}
