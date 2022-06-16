package pgx_repository

import (
	"github.com/homework3/comments/internal/db_model"
	"golang.org/x/net/context"
)

func (r *repository) AddComment(ctx context.Context, comment string, itemId, userId int32, statusId string) (int64, error) {
	const query = `
		INSERT INTO comment (
			user_id, item_id, comment, status_id
		) VALUES (
			$1, $2, $3, $4
		) RETURNING id
	`

	var commentId int64
	err := r.pool.QueryRow(ctx, query,
		userId,
		itemId,
		comment,
		statusId,
	).Scan(&commentId)

	return commentId, err
}

func (r *repository) UpdateCommentStatus(ctx context.Context, id int64, statusId string) error {
	const query = `
		UPDATE comment
		set status_id = $2
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id, statusId)

	return err
}

func (r *repository) GetComments(ctx context.Context, itemId int32) ([]db_model.Comment, error) {
	const query = `
		SELECT
			id,
			user_id,
			item_id,
			comment,
			status_id
		FROM comment
		WHERE item_id = $1
	`

	rows, err := r.pool.Query(ctx, query, itemId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]db_model.Comment, 0)
	for rows.Next() {
		var c db_model.Comment
		err = rows.Scan(
			&c.Id,
			&c.UserId,
			&c.ItemId,
			&c.Comment,
			&c.StatusId,
		)
		if err != nil {
			return nil, err
		}

		comments = append(comments, c)
	}

	return comments, nil
}
