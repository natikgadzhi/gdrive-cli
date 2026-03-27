package api

import (
	"fmt"

	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"

	"github.com/natikgadzhi/gdrive-cli/internal/config"
)

// commentsFields is the fields parameter for the comments.list API call.
// The Drive API requires explicit field selection for comments.
const commentsFields = "comments(id,content,htmlContent,author(displayName,me,photoLink),createdTime,modifiedTime,resolved,deleted,replies(id,content,htmlContent,author(displayName,me,photoLink),createdTime,modifiedTime,action,deleted),quotedFileContent),nextPageToken"

// ListComments fetches all comments (with inline replies) for a Google Drive file.
// It auto-paginates through all results. Deleted comments are excluded.
func ListComments(svc *drive.Service, fileID string) ([]CommentThread, error) {
	config.DebugLog("Fetching comments for file %s", fileID)

	var threads []CommentThread

	call := svc.Comments.List(fileID).
		Fields(googleapi.Field(commentsFields)).
		PageSize(100)

	err := call.Pages(nil, func(cl *drive.CommentList) error {
		for _, c := range cl.Comments {
			if c.Deleted {
				continue
			}

			thread := CommentThread{
				ID:           c.Id,
				Content:      c.Content,
				CreatedTime:  c.CreatedTime,
				ModifiedTime: c.ModifiedTime,
				Resolved:     c.Resolved,
			}

			if c.Author != nil {
				thread.Author = CommentAuthor{
					DisplayName: c.Author.DisplayName,
					IsMe:        c.Author.Me,
				}
			}

			if c.QuotedFileContent != nil {
				thread.QuotedText = c.QuotedFileContent.Value
			}

			for _, r := range c.Replies {
				if r.Deleted {
					continue
				}
				reply := CommentReply{
					ID:          r.Id,
					Content:     r.Content,
					CreatedTime: r.CreatedTime,
					Action:      r.Action,
				}
				if r.Author != nil {
					reply.Author = CommentAuthor{
						DisplayName: r.Author.DisplayName,
						IsMe:        r.Author.Me,
					}
				}
				thread.Replies = append(thread.Replies, reply)
			}

			threads = append(threads, thread)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}

	config.DebugLog("Fetched %d comment threads", len(threads))
	return threads, nil
}
