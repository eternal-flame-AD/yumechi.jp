package comment

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/google/go-github/v37/github"
)

type FormResponse struct {
	Comment        *Comment
	PullRequestRef string
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
	branchName := fmt.Sprintf("comment_%s_%s", base, entryId)

	_, _, err := githubClient.Git.GetRef(context.Background(), owner, repo, branchName)
	if err == nil {
		return branchName, nil
	}

	ref, _, err := githubClient.Git.GetRef(context.Background(), owner, repo, base)
	if err != nil {
		return "", err
	}

	_, _, err = githubClient.Git.CreateRef(context.Background(), owner, repo, ref)
	if err != nil {
		return "", err
	}

	if prID != nil {
		pr, _, err := githubClient.PullRequests.List(context.Background(), owner, repo, &github.PullRequestListOptions{
			Base: base,
			Head: fmt.Sprintf("%s:%s", owner, branchName),
		})
		if err != nil {
			return branchName, err
		}
		if len(pr) == 0 {
			pr, _, err := githubClient.PullRequests.Create(context.Background(), owner, repo, &github.NewPullRequest{
				Title: refStr(fmt.Sprintf("Comment on %s", entryId)),
				Base:  refStr(base),
				Head:  &branchName,
				Body:  refStr("This is an auto generated PR for comments."),
			})
			if err != nil {
				return branchName, err
			}
			*prID = *pr.ID
		}
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

func HandleForm(form url.Values, origin string) (*FormResponse, bool, error) {
	postToBranch := "dev"
	if origin == "yumechi.jp" || origin == "yumechi.jp:443" {
		postToBranch = "main"
	}

	entryId := form.Get("options[entryId")
	if entryId == "" {
		return nil, true, fmt.Errorf("missing entry id")
	}

	var prId int64
	prBranch, err := EnsurePRBranch(postToBranch, entryId, &prId)
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
		Name:        form.Get("fields[name]"),
		ReplyThread: form.Get("fields[replyThread]"),
		ReplyID:     form.Get("fields[replyID]"),
		ReplyName:   form.Get("fields[replyname]"),
		Website:     form.Get("fields[website]"),
		Email:       form.Get("fields[email]"),
		Date:        time.Now(),
		Body:        form.Get("fields[body]"),
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

	return &FormResponse{
		Comment:        &newComment,
		PullRequestRef: *newContentResponse.HTMLURL,
	}, false, nil
}