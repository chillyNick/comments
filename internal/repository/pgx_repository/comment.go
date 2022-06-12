package pgx_repository

import "golang.org/x/net/context"

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
